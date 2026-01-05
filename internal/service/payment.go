package service

import (
	"context"
	"math/rand"
	"time"

	"payment-gateway/internal/domain"
	"payment-gateway/internal/messaging"
	"payment-gateway/internal/repository"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type PaymentService interface {
	CreatePayment(ctx context.Context, req domain.CreatePaymentRequest) (*domain.Payment, error)
	GetPayment(ctx context.Context, id uuid.UUID) (*domain.Payment, error)
	GetPaymentByReference(ctx context.Context, reference string) (*domain.Payment, error)
	ListPayments(ctx context.Context, page, limit int) ([]*domain.Payment, int, error)
	ProcessPayment(ctx context.Context, id uuid.UUID) error
	GetStatistics(ctx context.Context) (*PaymentStatistics, error)
}

type paymentService struct {
	repo      repository.PaymentRepository
	publisher messaging.PaymentPublisher
	logger    *logrus.Logger
}

// Ethiopian Payment Statistics
type PaymentStatistics struct {
	TotalPayments      int     `json:"total_payments"`
	TotalAmountETB     float64 `json:"total_amount_etb"`
	TotalAmountUSD     float64 `json:"total_amount_usd"`
	SuccessfulPayments int     `json:"successful_payments"`
	FailedPayments     int     `json:"failed_payments"`
	PendingPayments    int     `json:"pending_payments"`
	AverageAmountETB   float64 `json:"average_amount_etb"`
	AverageAmountUSD   float64 `json:"average_amount_usd"`
}

func NewPaymentService(repo repository.PaymentRepository, publisher messaging.PaymentPublisher, logger *logrus.Logger) PaymentService {
	return &paymentService{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
	}
}

func (s *paymentService) CreatePayment(ctx context.Context, req domain.CreatePaymentRequest) (*domain.Payment, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check if payment with same reference already exists
	existing, err := s.repo.GetByReference(ctx, req.Reference)
	if err != nil && err != domain.ErrPaymentNotFound {
		s.logger.WithError(err).Error("Failed to check existing payment")
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrPaymentAlreadyExists
	}

	// Create payment
	now := time.Now().UTC()
	payment := &domain.Payment{
		ID:           uuid.New(),
		Amount:       req.Amount,
		Currency:     req.Currency,
		Reference:    req.Reference,
		Status:       domain.StatusPending,
		Description:  req.Description,
		CustomerName: req.CustomerName,
		BankCode:     req.BankCode,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Save to database
	if err := s.repo.Create(ctx, payment); err != nil {
		s.logger.WithError(err).Error("Failed to create payment")
		return nil, err
	}

	// Publish message for async processing
	if err := s.publisher.PublishPaymentCreated(ctx, payment.ID); err != nil {
		s.logger.WithError(err).Error("Failed to publish payment message")
		// We still return success, as payment is created in database
	}

	s.logger.WithFields(logrus.Fields{
		"payment_id":    payment.ID,
		"reference":     payment.Reference,
		"amount":        payment.Amount,
		"currency":      payment.Currency,
		"customer_name": payment.CustomerName,
		"bank_code":     payment.BankCode,
	}).Info("Ethiopian payment created successfully")

	return payment, nil
}

func (s *paymentService) GetPayment(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	payment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("payment_id", id).Error("Failed to get payment")
		return nil, err
	}

	s.logger.WithField("payment_id", id).Debug("Payment retrieved")
	return payment, nil
}

func (s *paymentService) GetPaymentByReference(ctx context.Context, reference string) (*domain.Payment, error) {
	payment, err := s.repo.GetByReference(ctx, reference)
	if err != nil {
		s.logger.WithError(err).WithField("reference", reference).Error("Failed to get payment by reference")
		return nil, err
	}

	return payment, nil
}

func (s *paymentService) ListPayments(ctx context.Context, page, limit int) ([]*domain.Payment, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Get paginated payments
	payments, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list payments")
		return nil, 0, err
	}

	// For demo purposes, return length as total
	total := len(payments)

	return payments, total, nil
}

func (s *paymentService) ProcessPayment(ctx context.Context, id uuid.UUID) error {
	s.logger.WithField("payment_id", id).Info("Starting payment processing")

	// Idempotent processing - only process if pending
	// Simulate external payment processing
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(500)+100))

	// Simulate different payment processors based on bank code
	var successRate float64
	payment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Different success rates for different Ethiopian banks
	switch payment.BankCode {
	case "CBE":
		successRate = 0.95 // Commercial Bank of Ethiopia - high success rate
	case "AWASH":
		successRate = 0.90 // Awash Bank
	case "DASHEN":
		successRate = 0.88 // Dashen Bank
	case "ABYSSINIA":
		successRate = 0.92 // Bank of Abyssinia
	default:
		successRate = 0.85 // Default for other banks
	}

	// Random result based on success rate
	var newStatus domain.PaymentStatus
	if rand.Float64() < successRate {
		newStatus = domain.StatusSuccess
		s.logger.WithField("payment_id", id).Info("Payment processing successful")
	} else {
		newStatus = domain.StatusFailed
		s.logger.WithField("payment_id", id).Warn("Payment processing failed")
	}

	// Update status atomically if still pending
	updated, err := s.repo.UpdateStatusIfPending(ctx, id, newStatus)
	if err != nil {
		s.logger.WithError(err).WithField("payment_id", id).Error("Failed to update payment status")
		return err
	}

	if !updated {
		s.logger.WithField("payment_id", id).Info("Payment already processed, skipping")
		return nil // Idempotent - no error if already processed
	}

	s.logger.WithFields(logrus.Fields{
		"payment_id": id,
		"status":     newStatus,
		"bank_code":  payment.BankCode,
		"amount":     payment.Amount,
		"currency":   payment.Currency,
	}).Info("Ethiopian payment processed successfully")

	return nil
}

func (s *paymentService) GetStatistics(ctx context.Context) (*PaymentStatistics, error) {
	// Get all payments (limited for demo)
	payments, err := s.repo.List(ctx, 1000, 0)
	if err != nil {
		return nil, err
	}

	stats := &PaymentStatistics{
		TotalPayments: len(payments),
	}

	var totalETB, totalUSD float64
	var etbCount, usdCount int

	for _, payment := range payments {
		switch payment.Status {
		case domain.StatusSuccess:
			stats.SuccessfulPayments++
		case domain.StatusFailed:
			stats.FailedPayments++
		case domain.StatusPending:
			stats.PendingPayments++
		}

		if payment.Currency == domain.CurrencyETB {
			totalETB += payment.Amount
			etbCount++
		} else if payment.Currency == domain.CurrencyUSD {
			totalUSD += payment.Amount
			usdCount++
		}
	}

	stats.TotalAmountETB = totalETB
	stats.TotalAmountUSD = totalUSD

	if etbCount > 0 {
		stats.AverageAmountETB = totalETB / float64(etbCount)
	}
	if usdCount > 0 {
		stats.AverageAmountUSD = totalUSD / float64(usdCount)
	}

	return stats, nil
}
