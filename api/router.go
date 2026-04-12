// Package api wires up all HTTP routes.  It is the composition root
// for the "driving" side of the hexagonal architecture.
package api

import (
	"github.com/gin-gonic/gin"

	"github.com/finbud/finbud-backend/internal/handler"
	"github.com/finbud/finbud-backend/internal/metrics"
	"github.com/finbud/finbud-backend/internal/middleware"
)

// NewRouter creates the Gin engine with all routes registered.
func NewRouter(
	userH *handler.UserHandler,
	profileH *handler.ProfileHandler,
	simulatorH *handler.SimulatorHandler,
) *gin.Engine {
	r := gin.New()

	// ── Global middleware ────────────────────────────────────────
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.CORS())

	// ── Health check ────────────────────────────────────────────
	r.GET("/health", handler.HealthCheck)
	r.GET("/metrics", gin.WrapH(metrics.Handler()))

	// ── API v1 ──────────────────────────────────────────────────
	v1 := r.Group("/api/v1")
	{
		// Users
		users := v1.Group("/users")
		{
			users.POST("", userH.CreateUser)
			users.GET("/:id", userH.GetUser)
			users.DELETE("/:id", userH.DeleteUser)

			// Financial Profile (nested under user)
			users.GET("/:id/profile", profileH.GetProfile)
			users.PUT("/:id/profile", profileH.UpsertProfile)

			// Decision Engine — Simulations
			users.POST("/:id/simulate/compound-interest", simulatorH.CompoundInterest)
			users.POST("/:id/simulate/debt-payoff", simulatorH.DebtPayoff)
			users.GET("/:id/scenarios", simulatorH.ScenarioHistory)
		}
	}

	return r
}

