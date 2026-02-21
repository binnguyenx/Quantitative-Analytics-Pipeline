package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/finbud/finbud-backend/internal/domain"
)

// profileRepo is the GORM-backed implementation of ProfileRepository.
type profileRepo struct {
	db *gorm.DB
}

// NewProfileRepository returns a production ProfileRepository.
func NewProfileRepository(db *gorm.DB) ProfileRepository {
	return &profileRepo{db: db}
}

// Upsert creates a FinancialProfile or updates it if one already
// exists for the given UserID (conflict on the unique index).
func (r *profileRepo) Upsert(ctx context.Context, profile *domain.FinancialProfile) error {
	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"monthly_income", "monthly_expenses", "total_debt",
				"debt_apr", "savings_balance", "credit_score", "updated_at",
			}),
		}).
		Create(profile)

	if result.Error != nil {
		return fmt.Errorf("profileRepo.Upsert: %w", result.Error)
	}
	return nil
}

func (r *profileRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.FinancialProfile, error) {
	var profile domain.FinancialProfile
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return nil, fmt.Errorf("profileRepo.GetByUserID: %w", err)
	}
	return &profile, nil
}

func (r *profileRepo) Delete(ctx context.Context, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&domain.FinancialProfile{}).Error; err != nil {
		return fmt.Errorf("profileRepo.Delete: %w", err)
	}
	return nil
}

