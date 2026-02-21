// FinBud Backend — Entry Point
//
// This is the composition root.  It wires together every layer of
// the Clean / Hexagonal Architecture:
//
//   Config → Infrastructure (DB, Redis, Kafka)
//            → Repositories → Services → Handlers → Router → Server
//
// All configuration is read from environment variables (see .env.example).
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/finbud/finbud-backend/api"
	"github.com/finbud/finbud-backend/internal/config"
	"github.com/finbud/finbud-backend/internal/handler"
	"github.com/finbud/finbud-backend/internal/repository"
	"github.com/finbud/finbud-backend/internal/service"
	"github.com/finbud/finbud-backend/pkg/database"
	kafkapkg "github.com/finbud/finbud-backend/pkg/kafka"
	redispkg "github.com/finbud/finbud-backend/pkg/redis"
)

func main() {
	// ── 1. Load configuration ───────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	gin.SetMode(cfg.GinMode)

	// ── 2. Connect to PostgreSQL ────────────────────────────────
	db, err := database.NewPostgresDB(cfg.PostgresDSN())
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	log.Println("main: PostgreSQL ready")

	// ── 3. Connect to Redis ─────────────────────────────────────
	rdb, err := redispkg.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	log.Println("main: Redis ready")

	// ── 4. Create Kafka producer ────────────────────────────────
	producer := kafkapkg.NewProducer(cfg.KafkaBrokers, cfg.KafkaTopicProfileUpdated)
	defer func() {
		if err := producer.Close(); err != nil {
			log.Printf("main: kafka producer close error: %v", err)
		}
	}()
	log.Println("main: Kafka producer ready")

	// ── 5. Repositories (driven adapters) ───────────────────────
	userRepo := repository.NewUserRepository(db)
	profileRepo := repository.NewProfileRepository(db)
	scenarioRepo := repository.NewScenarioRepository(db)

	// ── 6. Services (use-case layer) ────────────────────────────
	userSvc := service.NewUserService(userRepo)
	profileSvc := service.NewProfileService(profileRepo, rdb, producer)
	simulatorSvc := service.NewSimulatorService(scenarioRepo, rdb)

	// ── 7. HTTP handlers (driving adapters) ─────────────────────
	userHandler := handler.NewUserHandler(userSvc)
	profileHandler := handler.NewProfileHandler(profileSvc)
	simulatorHandler := handler.NewSimulatorHandler(simulatorSvc)

	// ── 8. Router ───────────────────────────────────────────────
	router := api.NewRouter(userHandler, profileHandler, simulatorHandler)

	// ── 9. HTTP server with graceful shutdown ───────────────────
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.ServerPort),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	// Start server in a goroutine so it doesn't block.
	go func() {
		log.Printf("main: FinBud API listening on :%s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("main: server error: %v", err)
		}
	}()

	// ── 10. Wait for interrupt signal ───────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("main: shutting down gracefully…")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("main: forced shutdown: %v", err)
	}

	log.Println("main: server stopped ✓")
}

