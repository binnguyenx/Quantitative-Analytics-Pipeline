package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/finbud/finbud-backend/internal/domain"
	"github.com/finbud/finbud-backend/internal/service"
)

// SimulatorHandler groups all Decision Engine endpoints.
type SimulatorHandler struct {
	simSvc *service.SimulatorService
}

// NewSimulatorHandler constructs a SimulatorHandler.
func NewSimulatorHandler(svc *service.SimulatorService) *SimulatorHandler {
	return &SimulatorHandler{simSvc: svc}
}

// CompoundInterest godoc
// POST /api/v1/users/:id/simulate/compound-interest
//
// Decision Engine — Compound Interest Simulation
// ──────────────────────────────────────────────────────────────
// Accepts principal, monthly contribution, annual rate and years.
// Returns the projected future value, total contributions and
// total interest earned so users can see the power of compounding.
func (h *SimulatorHandler) CompoundInterest(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var input domain.CompoundInterestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.simSvc.SimulateCompoundInterest(c.Request.Context(), userID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// DebtPayoff godoc
// POST /api/v1/users/:id/simulate/debt-payoff
//
// Decision Engine — Debt Payoff Simulation
// ──────────────────────────────────────────────────────────────
// Accepts total debt, APR and monthly payment.  Returns the
// number of months to pay off the debt, total paid, total
// interest and a feasibility flag.
func (h *SimulatorHandler) DebtPayoff(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var input domain.DebtPayoffInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.simSvc.SimulateDebtPayoff(c.Request.Context(), userID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ScenarioHistory godoc
// GET /api/v1/users/:id/scenarios?limit=20&offset=0
func (h *SimulatorHandler) ScenarioHistory(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	scenarios, err := h.simSvc.GetScenarioHistory(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   scenarios,
		"limit":  limit,
		"offset": offset,
	})
}

