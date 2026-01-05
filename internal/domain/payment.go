package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Ethiopian Currency types
type Currency string

const (
	CurrencyETB Currency = "ETB" // Ethiopian Birr
	CurrencyUSD Currency = "USD" // US Dollar
)

// Ethiopian Bank codes (for reference generation)
type EthiopianBank string

const (
	BankCBE       EthiopianBank = "CBE"       // Commercial Bank of Ethiopia
	BankAwash     EthiopianBank = "AWASH"     // Awash Bank
	BankDashen    EthiopianBank = "DASHEN"    // Dashen Bank
	BankAbyssinia EthiopianBank = "ABYSSINIA" // Bank of Abyssinia
	BankNib       EthiopianBank = "NIB"       // Nib International Bank
)

func (c Currency) IsValid() bool {
	return c == CurrencyETB || c == CurrencyUSD
}

func (c Currency) GetSymbol() string {
	switch c {
	case CurrencyETB:
		return "Br"
	case CurrencyUSD:
		return "$"
	default:
		return ""
	}
}

// Payment statuses
type PaymentStatus string

const (
	StatusPending PaymentStatus = "PENDING"
	StatusSuccess PaymentStatus = "SUCCESS"
	StatusFailed  PaymentStatus = "FAILED"
)

func (s PaymentStatus) IsTerminal() bool {
	return s == StatusSuccess || s == StatusFailed
}

// Payment represents an Ethiopian payment transaction
type Payment struct {
	ID           uuid.UUID     `json:"id"`
	Amount       float64       `json:"amount"`
	Currency     Currency      `json:"currency"`
	Reference    string        `json:"reference"`
	Status       PaymentStatus `json:"status"`
	Description  string        `json:"description,omitempty"`   // Ethiopian context: e.g., "Coffee export payment"
	CustomerName string        `json:"customer_name,omitempty"` // Ethiopian customer name
	BankCode     string        `json:"bank_code,omitempty"`     // Ethiopian bank code
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

// Ethiopian payment request with validation
type CreatePaymentRequest struct {
	Amount       float64  `json:"amount" validate:"required,gt=0"`
	Currency     Currency `json:"currency" validate:"required,oneof=ETB USD"`
	Reference    string   `json:"reference" validate:"required,min=5,max=50"`
	Description  string   `json:"description,omitempty" validate:"max=200"`
	CustomerName string   `json:"customer_name,omitempty" validate:"max=100"`
	BankCode     string   `json:"bank_code,omitempty" validate:"max=20"`
}

// Validate Ethiopian payment request
func (r *CreatePaymentRequest) Validate() error {
	if r.Amount <= 0 {
		return errors.New("amount must be greater than zero")
	}

	if !r.Currency.IsValid() {
		return errors.New("currency must be ETB or USD")
	}

	if len(r.Reference) < 5 {
		return errors.New("reference must be at least 5 characters")
	}

	if len(r.Reference) > 50 {
		return errors.New("reference is too long")
	}

	// Ethiopian business rule: For large ETB amounts, require description
	if r.Currency == CurrencyETB && r.Amount > 100000 && r.Description == "" {
		return errors.New("description is required for large ETB payments")
	}

	return nil
}

// Ethiopian payment response
type PaymentResponse struct {
	ID             uuid.UUID     `json:"id"`
	Amount         float64       `json:"amount"`
	Currency       Currency      `json:"currency"`
	CurrencySymbol string        `json:"currency_symbol"`
	Reference      string        `json:"reference"`
	Status         PaymentStatus `json:"status"`
	Description    string        `json:"description,omitempty"`
	CustomerName   string        `json:"customer_name,omitempty"`
	BankCode       string        `json:"bank_code,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
	CreatedAtET    string        `json:"created_at_et"` // Ethiopian time
}

// Convert to response with Ethiopian context
func (p *Payment) ToResponse() PaymentResponse {
	return PaymentResponse{
		ID:             p.ID,
		Amount:         p.Amount,
		Currency:       p.Currency,
		CurrencySymbol: p.Currency.GetSymbol(),
		Reference:      p.Reference,
		Status:         p.Status,
		Description:    p.Description,
		CustomerName:   p.CustomerName,
		BankCode:       p.BankCode,
		CreatedAt:      p.CreatedAt,
		CreatedAtET:    p.CreatedAt.Add(3 * time.Hour).Format(time.RFC3339), // GMT+3
	}
}

// Ethiopian errors
var (
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrInvalidInput         = errors.New("invalid input")
	ErrPaymentAlreadyExists = errors.New("payment with this reference already exists")
	ErrPaymentNotPending    = errors.New("payment is not in pending state")
	ErrAmountTooLarge       = errors.New("amount exceeds Ethiopian regulatory limit")
	ErrBusinessHours        = errors.New("payment outside Ethiopian business hours")
	ErrDatabase             = errors.New("database error")
)
