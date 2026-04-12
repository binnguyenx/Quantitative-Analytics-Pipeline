package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/finbud/finbud-backend/internal/domain"
)

type ingestedEventRepo struct {
	db *gorm.DB
}

// NewIngestedEventRepository returns a production IngestedEventRepository.
func NewIngestedEventRepository(db *gorm.DB) IngestedEventRepository {
	return &ingestedEventRepo{db: db}
}

func (r *ingestedEventRepo) CreateBatch(ctx context.Context, events []domain.IngestedEvent) (int64, error) {
	if len(events) == 0 {
		return 0, nil
	}

	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "event_id"}},
			DoNothing: true,
		}).
		Create(&events)
	if result.Error != nil {
		return 0, fmt.Errorf("ingestedEventRepo.CreateBatch: %w", result.Error)
	}
	return result.RowsAffected, nil
}
