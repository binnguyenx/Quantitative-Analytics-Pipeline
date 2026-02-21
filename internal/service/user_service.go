package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/finbud/finbud-backend/internal/domain"
	"github.com/finbud/finbud-backend/internal/repository"
)

// UserService handles user account operations.
type UserService struct {
	userRepo repository.UserRepository
}

// NewUserService constructs a UserService.
func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{userRepo: repo}
}

// CreateUser registers a new user.
func (s *UserService) CreateUser(ctx context.Context, user *domain.User) error {
	if err := s.userRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("userService.CreateUser: %w", err)
	}
	return nil
}

// GetUser retrieves a user by ID (with their FinancialProfile eager-loaded).
func (s *UserService) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("userService.GetUser: %w", err)
	}
	return user, nil
}

// UpdateUser persists changes to a user entity.
func (s *UserService) UpdateUser(ctx context.Context, user *domain.User) error {
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("userService.UpdateUser: %w", err)
	}
	return nil
}

// DeleteUser removes a user and (via CASCADE) their profile/scenarios.
func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if err := s.userRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("userService.DeleteUser: %w", err)
	}
	return nil
}

