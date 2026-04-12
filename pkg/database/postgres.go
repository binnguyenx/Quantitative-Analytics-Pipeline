// Package database provides a thin wrapper around GORM to connect
// to PostgreSQL and auto-migrate domain models.
package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/finbud/finbud-backend/internal/domain"
)

// NewPostgresDB opens a connection pool to PostgreSQL and runs
// auto-migrations for all domain models.
func NewPostgresDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("database: failed to connect to postgres: %w", err)
	}

	// Tune the underlying *sql.DB connection pool for high concurrency.
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("database: failed to get sql.DB handle: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	log.Println("database: connected to PostgreSQL ✓")

	// Auto-migrate creates / updates tables to match struct definitions.
	if err := db.AutoMigrate(
		&domain.User{},
		&domain.FinancialProfile{},
		&domain.DecisionScenario{},
		&domain.IngestedEvent{},
	); err != nil {
		return nil, fmt.Errorf("database: auto-migration failed: %w", err)
	}

	log.Println("database: auto-migration complete ✓")
	return db, nil
}

