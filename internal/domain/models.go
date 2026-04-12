// Package domain contains the core business entities for FinBud.
// These models are persistence-agnostic; GORM tags are added for
// convenience but the domain logic never depends on GORM directly.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// ─────────────────────────────────────────────────────────────
// User represents a Gen-Z user of the FinBud platform.
// ─────────────────────────────────────────────────────────────
type User struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Email     string    `gorm:"uniqueIndex;not null"                           json:"email"`
	Username  string    `gorm:"uniqueIndex;not null;size:50"                   json:"username"`
	FullName  string    `gorm:"not null;size:120"                              json:"full_name"`
	AgeGroup  string    `gorm:"size:10"                                        json:"age_group"` // e.g. "18-24"
	CreatedAt time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// One-to-one relationship
	FinancialProfile *FinancialProfile `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"financial_profile,omitempty"`
}

// ─────────────────────────────────────────────────────────────
// FinancialProfile stores a snapshot of a user's financial
// health: income, monthly expenses and outstanding debt.
// ─────────────────────────────────────────────────────────────
type FinancialProfile struct {
	ID              uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID          uuid.UUID `gorm:"type:uuid;uniqueIndex;not null"                 json:"user_id"`
	MonthlyIncome   float64   `gorm:"not null;default:0"                             json:"monthly_income"`
	MonthlyExpenses float64   `gorm:"not null;default:0"                             json:"monthly_expenses"`
	TotalDebt       float64   `gorm:"not null;default:0"                             json:"total_debt"`
	DebtAPR         float64   `gorm:"not null;default:0"                             json:"debt_apr"`         // Annual percentage rate on debt
	SavingsBalance  float64   `gorm:"not null;default:0"                             json:"savings_balance"`
	CreditScore     int       `gorm:"default:0"                                      json:"credit_score"`
	CreatedAt       time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

// MonthlySurplus returns how much money is left after expenses.
func (fp *FinancialProfile) MonthlySurplus() float64 {
	return fp.MonthlyIncome - fp.MonthlyExpenses
}

// ─────────────────────────────────────────────────────────────
// DecisionScenario represents a "what-if" simulation that a
// user runs through the Decision Engine.  For example:
//
//	"If I invest $200/month at 7% for 5 years, what do I get?"
//	"How many months to pay off $8,000 at 22% APR paying $300/mo?"
//
// ─────────────────────────────────────────────────────────────
type DecisionScenario struct {
	ID          uuid.UUID    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID      uuid.UUID    `gorm:"type:uuid;index;not null"                       json:"user_id"`
	Type        ScenarioType `gorm:"size:30;not null"                               json:"type"`
	Title       string       `gorm:"size:200;not null"                              json:"title"`
	InputJSON   string       `gorm:"type:jsonb;not null"                            json:"input_json"`  // raw scenario parameters
	OutputJSON  string       `gorm:"type:jsonb"                                     json:"output_json"` // simulation result
	CreatedAt   time.Time    `gorm:"autoCreateTime"                                 json:"created_at"`
}

// IngestedEvent stores profile update events consumed from Kafka.
// The unique event_id is used for idempotent ingestion.
type IngestedEvent struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	EventID        string    `gorm:"size:128;uniqueIndex;not null"                  json:"event_id"`
	EventType      string    `gorm:"size:64;index;not null"                         json:"event_type"`
	UserID         uuid.UUID `gorm:"type:uuid;index;not null"                       json:"user_id"`
	CorrelationID  string    `gorm:"size:128;index;not null"                        json:"correlation_id"`
	EventTimestamp time.Time `gorm:"not null"                                       json:"event_timestamp"`
	KafkaKey       string    `gorm:"size:128"                                       json:"kafka_key"`
	KafkaTopic     string    `gorm:"size:128;index"                                 json:"kafka_topic"`
	KafkaPartition int       `gorm:"not null"                                       json:"kafka_partition"`
	KafkaOffset    int64     `gorm:"not null"                                       json:"kafka_offset"`
	PayloadJSON    string    `gorm:"type:jsonb;not null"                            json:"payload_json"`
	ProcessedAt    time.Time `gorm:"not null"                                       json:"processed_at"`
	CreatedAt      time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
}

// ScenarioType enumerates the kinds of simulations the engine supports.
type ScenarioType string

const (
	ScenarioCompoundInterest ScenarioType = "COMPOUND_INTEREST"
	ScenarioDebtPayoff       ScenarioType = "DEBT_PAYOFF"
)

// ─────────────────────────────────────────────────────────────
// DTOs (Data Transfer Objects) used at the API boundary.
// ─────────────────────────────────────────────────────────────

// CompoundInterestInput captures the parameters for a compound
// interest simulation.
type CompoundInterestInput struct {
	Principal          float64 `json:"principal"            binding:"required,gt=0"`   // initial amount
	MonthlyContribution float64 `json:"monthly_contribution" binding:"gte=0"`          // added every month
	AnnualRatePercent  float64 `json:"annual_rate_percent"  binding:"required,gt=0"`   // e.g. 7 for 7%
	Years              int     `json:"years"                binding:"required,gt=0"`   // investment horizon
}

// CompoundInterestOutput holds the simulation result.
type CompoundInterestOutput struct {
	FutureValue      float64 `json:"future_value"`
	TotalContributed float64 `json:"total_contributed"`
	TotalInterest    float64 `json:"total_interest"`
}

// DebtPayoffInput captures the parameters for a debt payoff timeline.
type DebtPayoffInput struct {
	TotalDebt      float64 `json:"total_debt"       binding:"required,gt=0"`
	AnnualAPR      float64 `json:"annual_apr"        binding:"required,gt=0"` // e.g. 22 for 22%
	MonthlyPayment float64 `json:"monthly_payment"   binding:"required,gt=0"`
}

// DebtPayoffOutput holds the payoff simulation result.
type DebtPayoffOutput struct {
	MonthsToPayoff int     `json:"months_to_payoff"`
	TotalPaid      float64 `json:"total_paid"`
	TotalInterest  float64 `json:"total_interest"`
	IsFeasible     bool    `json:"is_feasible"` // false if payment < interest
}

// UpdateProfileRequest is the payload accepted by PUT /profile.
type UpdateProfileRequest struct {
	MonthlyIncome   *float64 `json:"monthly_income"`
	MonthlyExpenses *float64 `json:"monthly_expenses"`
	TotalDebt       *float64 `json:"total_debt"`
	DebtAPR         *float64 `json:"debt_apr"`
	SavingsBalance  *float64 `json:"savings_balance"`
	CreditScore     *int     `json:"credit_score"`
}

