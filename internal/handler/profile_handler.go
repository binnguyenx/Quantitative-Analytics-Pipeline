package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/finbud/finbud-backend/internal/domain"
	"github.com/finbud/finbud-backend/internal/service"
)

// ProfileHandler groups financial-profile endpoints.
type ProfileHandler struct {
	profileSvc *service.ProfileService
}

// NewProfileHandler constructs a ProfileHandler.
func NewProfileHandler(svc *service.ProfileService) *ProfileHandler {
	return &ProfileHandler{profileSvc: svc}
}

// GetProfile godoc
// GET /api/v1/users/:id/profile
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	profile, err := h.profileSvc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// UpsertProfile godoc
// PUT /api/v1/users/:id/profile
//
// This endpoint triggers:
//  1. DB upsert
//  2. Redis cache invalidation
//  3. Kafka EVENT_PROFILE_UPDATED emission
func (h *ProfileHandler) UpsertProfile(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var req domain.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build a profile from the request, setting only provided fields.
	profile := &domain.FinancialProfile{
		UserID:    userID,
		UpdatedAt: time.Now(),
	}
	if req.MonthlyIncome != nil {
		profile.MonthlyIncome = *req.MonthlyIncome
	}
	if req.MonthlyExpenses != nil {
		profile.MonthlyExpenses = *req.MonthlyExpenses
	}
	if req.TotalDebt != nil {
		profile.TotalDebt = *req.TotalDebt
	}
	if req.DebtAPR != nil {
		profile.DebtAPR = *req.DebtAPR
	}
	if req.SavingsBalance != nil {
		profile.SavingsBalance = *req.SavingsBalance
	}
	if req.CreditScore != nil {
		profile.CreditScore = *req.CreditScore
	}

	if err := h.profileSvc.UpsertProfile(c.Request.Context(), profile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}

