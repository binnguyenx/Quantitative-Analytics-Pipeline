// Package repository defines the ports (interfaces) that the domain
// layer requires for data persistence.  Concrete implementations
// live alongside in this package but depend only on GORM, never on
// HTTP or messaging concerns — keeping layers cleanly separated.
package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/finbud/finbud-backend/internal/domain"
)

// UserRepository describes the persistence contract for User entities.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ProfileRepository handles FinancialProfile CRUD.
type ProfileRepository interface {
	Upsert(ctx context.Context, profile *domain.FinancialProfile) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.FinancialProfile, error)
	Delete(ctx context.Context, userID uuid.UUID) error
}

// ScenarioRepository handles DecisionScenario persistence.
type ScenarioRepository interface {
	Create(ctx context.Context, scenario *domain.DecisionScenario) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.DecisionScenario, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.DecisionScenario, error)
}

// IngestedEventRepository handles persistence for consumed Kafka events.
type IngestedEventRepository interface {
	CreateBatch(ctx context.Context, events []domain.IngestedEvent) (int64, error)
}

