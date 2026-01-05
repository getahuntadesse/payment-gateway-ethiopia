package repository

import (
	"context"
	"errors"
	"time"

	"payment-gateway/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *domain.Payment) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error)
	GetByReference(ctx context.Context, reference string) (*domain.Payment, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.PaymentStatus) error
	UpdateStatusIfPending(ctx context.Context, id uuid.UUID, status domain.PaymentStatus) (bool, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Payment, error)
	Count(ctx context.Context) (int, error)
}

type paymentRepository struct {
	db     *pgxpool.Pool
	logger *logrus.Logger
}

func NewPaymentRepository(db *pgxpool.Pool, logger *logrus.Logger) PaymentRepository {
	return &paymentRepository{db: db, logger: logger}
}

func (r *paymentRepository) Create(ctx context.Context, payment *domain.Payment) error {
	query := `
		INSERT INTO payments (id, amount, currency, reference, status, description, customer_name, bank_code, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (reference) DO NOTHING
		RETURNING id
	`

	err := r.db.QueryRow(ctx, query,
		payment.ID,
		payment.Amount,
		payment.Currency,
		payment.Reference,
		payment.Status,
		payment.Description,
		payment.CustomerName,
		payment.BankCode,
		payment.CreatedAt,
		payment.UpdatedAt,
	).Scan(&payment.ID)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrPaymentAlreadyExists
	}

	if err != nil {
		r.logger.WithError(err).Error("Failed to create payment")
		return domain.ErrDatabase
	}

	r.logger.WithFields(logrus.Fields{
		"payment_id": payment.ID,
		"reference":  payment.Reference,
		"currency":   payment.Currency,
		"amount":     payment.Amount,
	}).Info("Payment created successfully")

	return nil
}

func (r *paymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	query := `
		SELECT id, amount, currency, reference, status, description, customer_name, bank_code, created_at, updated_at
		FROM payments
		WHERE id = $1
	`

	var payment domain.Payment
	err := r.db.QueryRow(ctx, query, id).Scan(
		&payment.ID,
		&payment.Amount,
		&payment.Currency,
		&payment.Reference,
		&payment.Status,
		&payment.Description,
		&payment.CustomerName,
		&payment.BankCode,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrPaymentNotFound
	}

	if err != nil {
		r.logger.WithError(err).Error("Failed to get payment by ID")
		return nil, domain.ErrDatabase
	}

	return &payment, nil
}

func (r *paymentRepository) GetByReference(ctx context.Context, reference string) (*domain.Payment, error) {
	query := `
		SELECT id, amount, currency, reference, status, description, customer_name, bank_code, created_at, updated_at
		FROM payments
		WHERE reference = $1
	`

	var payment domain.Payment
	err := r.db.QueryRow(ctx, query, reference).Scan(
		&payment.ID,
		&payment.Amount,
		&payment.Currency,
		&payment.Reference,
		&payment.Status,
		&payment.Description,
		&payment.CustomerName,
		&payment.BankCode,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrPaymentNotFound
	}

	if err != nil {
		r.logger.WithError(err).Error("Failed to get payment by reference")
		return nil, domain.ErrDatabase
	}

	return &payment, nil
}

func (r *paymentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.PaymentStatus) error {
	query := `
		UPDATE payments
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.Exec(ctx, query, status, time.Now().UTC(), id)
	if err != nil {
		r.logger.WithError(err).Error("Failed to update payment status")
		return domain.ErrDatabase
	}

	if result.RowsAffected() == 0 {
		return domain.ErrPaymentNotFound
	}

	return nil
}

// Idempotent update - only updates if status is PENDING
func (r *paymentRepository) UpdateStatusIfPending(ctx context.Context, id uuid.UUID, newStatus domain.PaymentStatus) (bool, error) {
	// Start transaction for atomic update
	tx, err := r.db.Begin(ctx)
	if err != nil {
		r.logger.WithError(err).Error("Failed to begin transaction")
		return false, domain.ErrDatabase
	}
	defer tx.Rollback(ctx)

	// Lock the row for update
	var currentStatus domain.PaymentStatus
	err = tx.QueryRow(ctx,
		"SELECT status FROM payments WHERE id = $1 FOR UPDATE",
		id,
	).Scan(&currentStatus)

	if errors.Is(err, pgx.ErrNoRows) {
		return false, domain.ErrPaymentNotFound
	}
	if err != nil {
		r.logger.WithError(err).Error("Failed to lock payment row")
		return false, domain.ErrDatabase
	}

	// Check if it's still pending
	if currentStatus != domain.StatusPending {
		r.logger.WithFields(logrus.Fields{
			"payment_id":     id,
			"current_status": currentStatus,
		}).Info("Payment already processed, skipping")
		return false, nil
	}

	// Update status
	_, err = tx.Exec(ctx,
		"UPDATE payments SET status = $1, updated_at = $2 WHERE id = $3",
		newStatus, time.Now().UTC(), id,
	)
	if err != nil {
		r.logger.WithError(err).Error("Failed to update payment status in transaction")
		return false, domain.ErrDatabase
	}

	if err = tx.Commit(ctx); err != nil {
		r.logger.WithError(err).Error("Failed to commit transaction")
		return false, domain.ErrDatabase
	}

	return true, nil
}

func (r *paymentRepository) List(ctx context.Context, limit, offset int) ([]*domain.Payment, error) {
	query := `
		SELECT id, amount, currency, reference, status, description, customer_name, bank_code, created_at, updated_at
		FROM payments
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		r.logger.WithError(err).Error("Failed to list payments")
		return nil, domain.ErrDatabase
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		var payment domain.Payment
		err := rows.Scan(
			&payment.ID,
			&payment.Amount,
			&payment.Currency,
			&payment.Reference,
			&payment.Status,
			&payment.Description,
			&payment.CustomerName,
			&payment.BankCode,
			&payment.CreatedAt,
			&payment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		payments = append(payments, &payment)
	}

	return payments, nil
}

func (r *paymentRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM payments`

	var count int
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		r.logger.WithError(err).Error("Failed to count payments")
		return 0, domain.ErrDatabase
	}

	return count, nil
}
