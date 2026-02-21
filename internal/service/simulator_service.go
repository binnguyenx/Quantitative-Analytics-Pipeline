package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/finbud/finbud-backend/internal/domain"
	"github.com/finbud/finbud-backend/internal/repository"
	redispkg "github.com/finbud/finbud-backend/pkg/redis"
)

// SimulatorService is the core **Decision Engine** of FinBud.
//
// It provides two simulation models that help Gen-Z users visualise
// the long-term impact of their financial choices:
//
//  1. Compound Interest — "What will my savings grow to?"
//  2. Debt Payoff       — "How long until I'm debt-free?"
//
// Results are cached in Redis so identical queries are served instantly
// and every simulation is persisted as a DecisionScenario for the
// user's history.
type SimulatorService struct {
	scenarioRepo repository.ScenarioRepository
	rdb          *redis.Client
}

// NewSimulatorService wires up the simulator.
func NewSimulatorService(
	repo repository.ScenarioRepository,
	rdb *redis.Client,
) *SimulatorService {
	return &SimulatorService{
		scenarioRepo: repo,
		rdb:          rdb,
	}
}

// ─────────────────────────────────────────────────────────────────
// COMPOUND INTEREST SIMULATION
// ─────────────────────────────────────────────────────────────────
//
// Formula (Future Value of a Series with initial lump sum):
//
//   FV = P × (1 + r)^n  +  PMT × [((1 + r)^n − 1) / r]
//
// Where:
//   P   = principal (initial deposit)
//   PMT = monthly contribution
//   r   = monthly interest rate  (annual_rate / 12 / 100)
//   n   = total compounding periods  (years × 12)
//
// The output tells the user exactly how much of the future value
// came from their contributions vs. earned interest — a powerful
// motivator for young savers.
// ─────────────────────────────────────────────────────────────────

// SimulateCompoundInterest runs the compound interest model.
func (s *SimulatorService) SimulateCompoundInterest(
	ctx context.Context,
	userID uuid.UUID,
	input domain.CompoundInterestInput,
) (*domain.CompoundInterestOutput, error) {

	// ── 1. Check Redis cache ────────────────────────────────────
	cacheKey := fmt.Sprintf("sim:ci:%s:%.2f:%.2f:%.2f:%d",
		userID, input.Principal, input.MonthlyContribution,
		input.AnnualRatePercent, input.Years)

	if cached, err := s.rdb.Get(ctx, cacheKey).Bytes(); err == nil {
		var out domain.CompoundInterestOutput
		if json.Unmarshal(cached, &out) == nil {
			log.Printf("simulator: cache HIT for compound interest [user=%s]", userID)
			return &out, nil
		}
	}

	// ── 2. Run the calculation ──────────────────────────────────
	r := input.AnnualRatePercent / 100.0 / 12.0 // monthly rate
	n := float64(input.Years * 12)                // total periods

	// Future value of the lump-sum principal
	fvPrincipal := input.Principal * math.Pow(1+r, n)

	// Future value of the annuity (monthly contributions)
	fvAnnuity := input.MonthlyContribution * ((math.Pow(1+r, n) - 1) / r)

	futureValue := fvPrincipal + fvAnnuity
	totalContributed := input.Principal + (input.MonthlyContribution * n)
	totalInterest := futureValue - totalContributed

	output := &domain.CompoundInterestOutput{
		FutureValue:      math.Round(futureValue*100) / 100,
		TotalContributed: math.Round(totalContributed*100) / 100,
		TotalInterest:    math.Round(totalInterest*100) / 100,
	}

	// ── 3. Cache result ─────────────────────────────────────────
	go func() {
		data, _ := json.Marshal(output)
		s.rdb.Set(context.Background(), cacheKey, data, redispkg.DefaultCacheTTL)
	}()

	// ── 4. Persist scenario for user history ────────────────────
	go s.persistScenario(userID, domain.ScenarioCompoundInterest, input, output)

	return output, nil
}

// ─────────────────────────────────────────────────────────────────
// DEBT PAYOFF SIMULATION
// ─────────────────────────────────────────────────────────────────
//
// Algorithm (amortisation loop):
//
//   Each month:
//     interest_charge = remaining_balance × (APR / 12 / 100)
//     principal_paid  = monthly_payment − interest_charge
//     remaining_balance -= principal_paid
//
//   If monthly_payment ≤ interest_charge the debt will NEVER be
//   paid off (IsFeasible = false).  We cap the loop at 600 months
//   (50 years) as a safety valve.
//
// This gives Gen-Z users a concrete timeline showing how extra
// payments dramatically shorten the payoff period.
// ─────────────────────────────────────────────────────────────────

// SimulateDebtPayoff runs the debt payoff timeline model.
func (s *SimulatorService) SimulateDebtPayoff(
	ctx context.Context,
	userID uuid.UUID,
	input domain.DebtPayoffInput,
) (*domain.DebtPayoffOutput, error) {

	// ── 1. Check Redis cache ────────────────────────────────────
	cacheKey := fmt.Sprintf("sim:dp:%s:%.2f:%.2f:%.2f",
		userID, input.TotalDebt, input.AnnualAPR, input.MonthlyPayment)

	if cached, err := s.rdb.Get(ctx, cacheKey).Bytes(); err == nil {
		var out domain.DebtPayoffOutput
		if json.Unmarshal(cached, &out) == nil {
			log.Printf("simulator: cache HIT for debt payoff [user=%s]", userID)
			return &out, nil
		}
	}

	// ── 2. Run the amortisation loop ────────────────────────────
	monthlyRate := input.AnnualAPR / 100.0 / 12.0
	balance := input.TotalDebt
	totalPaid := 0.0
	months := 0
	const maxMonths = 600 // safety cap (50 years)

	// Quick feasibility check: payment must exceed first month's interest.
	firstMonthInterest := balance * monthlyRate
	if input.MonthlyPayment <= firstMonthInterest {
		return &domain.DebtPayoffOutput{
			MonthsToPayoff: -1,
			TotalPaid:      0,
			TotalInterest:  0,
			IsFeasible:     false,
		}, nil
	}

	for balance > 0 && months < maxMonths {
		interest := balance * monthlyRate
		principal := input.MonthlyPayment - interest

		// In the final month the payment might be less than the
		// regular amount, so we cap it.
		if principal > balance {
			totalPaid += balance + interest
			balance = 0
		} else {
			balance -= principal
			totalPaid += input.MonthlyPayment
		}
		months++
	}

	totalInterest := totalPaid - input.TotalDebt

	output := &domain.DebtPayoffOutput{
		MonthsToPayoff: months,
		TotalPaid:      math.Round(totalPaid*100) / 100,
		TotalInterest:  math.Round(totalInterest*100) / 100,
		IsFeasible:     true,
	}

	// ── 3. Cache result ─────────────────────────────────────────
	go func() {
		data, _ := json.Marshal(output)
		s.rdb.Set(context.Background(), cacheKey, data, redispkg.DefaultCacheTTL)
	}()

	// ── 4. Persist scenario ─────────────────────────────────────
	go s.persistScenario(userID, domain.ScenarioDebtPayoff, input, output)

	return output, nil
}

// ─────────────────────────────────────────────────────────────────
// HELPERS
// ─────────────────────────────────────────────────────────────────

// persistScenario saves a DecisionScenario row so the user can
// review past simulations.  Runs in a background goroutine to
// keep API latency low.
func (s *SimulatorService) persistScenario(
	userID uuid.UUID,
	scenarioType domain.ScenarioType,
	input interface{},
	output interface{},
) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	inputJSON, _ := json.Marshal(input)
	outputJSON, _ := json.Marshal(output)

	scenario := &domain.DecisionScenario{
		UserID:     userID,
		Type:       scenarioType,
		Title:      fmt.Sprintf("%s simulation", scenarioType),
		InputJSON:  string(inputJSON),
		OutputJSON: string(outputJSON),
	}

	if err := s.scenarioRepo.Create(ctx, scenario); err != nil {
		log.Printf("simulator: persist scenario error: %v", err)
	}
}

// GetScenarioHistory returns paginated past simulations for a user.
func (s *SimulatorService) GetScenarioHistory(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int,
) ([]domain.DecisionScenario, error) {
	return s.scenarioRepo.ListByUserID(ctx, userID, limit, offset)
}

