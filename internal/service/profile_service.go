// Package service contains the application's use-case logic.
// Services orchestrate domain models, repositories, cache and
// event producers without knowing about HTTP or transport details.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/finbud/finbud-backend/internal/domain"
	"github.com/finbud/finbud-backend/internal/repository"
	kafkapkg "github.com/finbud/finbud-backend/pkg/kafka"
	redispkg "github.com/finbud/finbud-backend/pkg/redis"
)

// ProfileService handles all operations on a user's FinancialProfile.
type ProfileService struct {
	profileRepo repository.ProfileRepository
	rdb         *redis.Client
	producer    *kafkapkg.Producer
}

// NewProfileService constructs a ProfileService with its dependencies.
func NewProfileService(
	repo repository.ProfileRepository,
	rdb *redis.Client,
	producer *kafkapkg.Producer,
) *ProfileService {
	return &ProfileService{
		profileRepo: repo,
		rdb:         rdb,
		producer:    producer,
	}
}

// GetProfile retrieves the financial profile for a user, checking
// Redis first for a cached copy.
func (s *ProfileService) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.FinancialProfile, error) {
	cacheKey := fmt.Sprintf("profile:%s", userID.String())

	// ── 1. Try cache ────────────────────────────────────────────
	cached, err := s.rdb.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var profile domain.FinancialProfile
		if json.Unmarshal(cached, &profile) == nil {
			log.Printf("profileService: cache HIT for user %s", userID)
			return &profile, nil
		}
	}

	// ── 2. Fallback to DB ───────────────────────────────────────
	profile, err := s.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("profileService.GetProfile: %w", err)
	}

	// ── 3. Warm cache asynchronously ────────────────────────────
	go func() {
		data, _ := json.Marshal(profile)
		if err := s.rdb.Set(context.Background(), cacheKey, data, redispkg.DefaultCacheTTL).Err(); err != nil {
			log.Printf("profileService: cache SET error: %v", err)
		}
	}()

	return profile, nil
}

// UpsertProfile creates or updates the financial profile and
// emits an EVENT_PROFILE_UPDATED Kafka message so downstream
// services can react (e.g. recalculate recommendations).
func (s *ProfileService) UpsertProfile(ctx context.Context, profile *domain.FinancialProfile) error {
	// ── 1. Persist ──────────────────────────────────────────────
	if err := s.profileRepo.Upsert(ctx, profile); err != nil {
		return fmt.Errorf("profileService.UpsertProfile: %w", err)
	}

	// ── 2. Invalidate cache ─────────────────────────────────────
	cacheKey := fmt.Sprintf("profile:%s", profile.UserID.String())
	if err := s.rdb.Del(ctx, cacheKey).Err(); err != nil {
		log.Printf("profileService: cache DEL error: %v", err)
	}

	// ── 3. Emit event (non-blocking) ────────────────────────────
	go s.emitProfileUpdatedEvent(profile)

	return nil
}

// emitProfileUpdatedEvent publishes the EVENT_PROFILE_UPDATED message
// to Kafka.  It runs in a goroutine so the API response is not blocked
// by broker latency.
func (s *ProfileService) emitProfileUpdatedEvent(profile *domain.FinancialProfile) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	eventID := uuid.NewString()
	now := time.Now().UTC()

	payload, err := json.Marshal(map[string]interface{}{
		"event":          "EVENT_PROFILE_UPDATED",
		"event_id":       eventID,
		"correlation_id": eventID,
		"user_id":        profile.UserID.String(),
		"timestamp":      now.Format(time.RFC3339),
		"data":           profile,
	})
	if err != nil {
		log.Printf("profileService: marshal event error: %v", err)
		return
	}

	if err := s.producer.Publish(ctx, []byte(profile.UserID.String()), payload); err != nil {
		log.Printf("profileService: kafka publish error: %v", err)
	} else {
		log.Printf("profileService: EVENT_PROFILE_UPDATED emitted for user %s", profile.UserID)
	}
}

