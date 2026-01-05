package worker

import (
	"context"
	"encoding/json"
	"time"

	"payment-gateway/internal/domain"
	"payment-gateway/internal/messaging"
	"payment-gateway/internal/service"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type PaymentProcessor struct {
	paymentService service.PaymentService
	rabbitMQ       *messaging.RabbitMQClient
	logger         *logrus.Logger
	workerCount    int
}

func NewPaymentProcessor(
	paymentService service.PaymentService,
	rabbitMQ *messaging.RabbitMQClient,
	logger *logrus.Logger,
	workerCount int,
) *PaymentProcessor {
	return &PaymentProcessor{
		paymentService: paymentService,
		rabbitMQ:       rabbitMQ,
		logger:         logger,
		workerCount:    workerCount,
	}
}

func (p *PaymentProcessor) Start(ctx context.Context) error {
	deliveries, err := p.rabbitMQ.Consume()
	if err != nil {
		return err
	}

	// Start multiple workers for concurrency
	for i := 0; i < p.workerCount; i++ {
		go p.worker(ctx, deliveries, i)
	}

	p.logger.WithFields(logrus.Fields{
		"worker_count": p.workerCount,
		"queue":        p.rabbitMQ.Config.QueueName,
	}).Info("Ethiopian Payment Processor started with workers")

	return nil
}

func (p *PaymentProcessor) worker(ctx context.Context, deliveries <-chan amqp.Delivery, workerID int) {
	logger := p.logger.WithFields(logrus.Fields{
		"worker_id": workerID,
		"component": "payment_worker",
	})

	logger.Info("Payment worker started")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Worker stopped by context")
			return
		case delivery, ok := <-deliveries:
			if !ok {
				logger.Warn("Delivery channel closed")
				return
			}

			// Process message with retry logic
			if err := p.processMessageWithRetry(ctx, delivery); err != nil {
				logger.WithError(err).Error("Failed to process message after retries")

				// Don't requeue, send to DLQ
				delivery.Nack(false, false)
			} else {
				// Acknowledge successful processing
				delivery.Ack(false)
			}
		}
	}
}

func (p *PaymentProcessor) processMessageWithRetry(ctx context.Context, delivery amqp.Delivery) error {
	var msg messaging.PaymentMessage
	if err := json.Unmarshal(delivery.Body, &msg); err != nil {
		p.logger.WithError(err).Error("Failed to unmarshal message")
		return err
	}

	logger := p.logger.WithFields(logrus.Fields{
		"payment_id": msg.PaymentID,
		"message_id": delivery.MessageId,
		"timestamp":  msg.Timestamp.Format(time.RFC3339),
	})

	logger.Info("Processing Ethiopian payment message")

	// Process payment with timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Check retry count from headers
	retryCount := 0
	if retryHeader, ok := delivery.Headers["retry_count"].(int32); ok {
		retryCount = int(retryHeader)
	}

	if retryCount > 3 {
		logger.Warn("Max retries exceeded, sending to DLQ")
		return domain.ErrPaymentNotPending
	}

	// Process the payment
	if err := p.paymentService.ProcessPayment(ctx, msg.PaymentID); err != nil {
		// If payment is not pending (already processed), we consider it success
		if err == domain.ErrPaymentNotPending {
			logger.Info("Payment already processed, acknowledging message")
			return nil
		}

		// For other errors, log and retry
		logger.WithError(err).Error("Failed to process payment")

		// Increment retry count
		delivery.Headers["retry_count"] = retryCount + 1

		// Reject with requeue (for retry)
		return err
	}

	logger.Info("Payment processed successfully")
	return nil
}
