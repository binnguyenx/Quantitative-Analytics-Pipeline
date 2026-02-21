package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/finbud/finbud-backend/internal/domain"
)

// scenarioRepo is the GORM-backed implementation of ScenarioRepository.
type scenarioRepo struct {
	db *gorm.DB
}

// NewScenarioRepository returns a production ScenarioRepository.
func NewScenarioRepository(db *gorm.DB) ScenarioRepository {
	return &scenarioRepo{db: db}
}

func (r *scenarioRepo) Create(ctx context.Context, scenario *domain.DecisionScenario) error {
	if err := r.db.WithContext(ctx).Create(scenario).Error; err != nil {
		return fmt.Errorf("scenarioRepo.Create: %w", err)
	}
	return nil
}

func (r *scenarioRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.DecisionScenario, error) {
	var s domain.DecisionScenario
	if err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("scenarioRepo.GetByID: %w", err)
	}
	return &s, nil
}

func (r *scenarioRepo) ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.DecisionScenario, error) {
	var scenarios []domain.DecisionScenario
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&scenarios).Error
	if err != nil {
		return nil, fmt.Errorf("scenarioRepo.ListByUserID: %w", err)
	}
	return scenarios, nil
}

