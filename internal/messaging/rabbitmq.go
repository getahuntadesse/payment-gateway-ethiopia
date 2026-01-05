package messaging

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type RabbitMQConfig struct {
	URL           string
	QueueName     string
	Exchange      string
	ConsumerTag   string
	PrefetchCount int
}

type RabbitMQClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
	logger  *logrus.Logger
	Config  RabbitMQConfig // Changed to exported (uppercase)
}

func NewRabbitMQClient(config RabbitMQConfig, logger *logrus.Logger) (*RabbitMQClient, error) {
	conn, err := amqp.Dial(config.URL)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Set QoS for fair dispatch
	err = channel.Qos(
		config.PrefetchCount, // prefetch count
		0,                    // prefetch size
		false,                // global
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	// Declare exchange
	err = channel.ExchangeDeclare(
		config.Exchange,
		"direct",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	// Declare queue with DLQ (Dead Letter Queue) for failed messages
	queue, err := channel.QueueDeclare(
		config.QueueName,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		amqp.Table{
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": config.QueueName + "_dlq",
		},
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	// Bind queue to exchange
	err = channel.QueueBind(
		queue.Name,
		"payment.created",
		config.Exchange,
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	// Declare DLQ
	_, err = channel.QueueDeclare(
		config.QueueName+"_dlq",
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	logger.Info("Connected to RabbitMQ successfully")

	return &RabbitMQClient{
		conn:    conn,
		channel: channel,
		queue:   queue,
		logger:  logger,
		Config:  config, // Changed to uppercase
	}, nil
}

func (c *RabbitMQClient) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *RabbitMQClient) Consume() (<-chan amqp.Delivery, error) {
	return c.channel.Consume(
		c.queue.Name,
		c.Config.ConsumerTag, // Use uppercase Config
		false,                // auto-ack
		false,                // exclusive
		false,                // no-local
		false,                // no-wait
		nil,                  // args
	)
}

type PaymentPublisher interface {
	PublishPaymentCreated(ctx context.Context, paymentID uuid.UUID) error
}

type paymentPublisher struct {
	client *RabbitMQClient
	logger *logrus.Logger
}

func NewPaymentPublisher(client *RabbitMQClient, logger *logrus.Logger) PaymentPublisher {
	return &paymentPublisher{
		client: client,
		logger: logger,
	}
}

func (p *paymentPublisher) PublishPaymentCreated(ctx context.Context, paymentID uuid.UUID) error {
	message := PaymentMessage{
		PaymentID: paymentID,
		Type:      "payment.created",
		Timestamp: time.Now().UTC(),
	}

	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = p.client.channel.PublishWithContext(
		ctx,
		p.client.Config.Exchange, // Use uppercase Config
		"payment.created",
		true,  // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			MessageId:    uuid.New().String(),
			Timestamp:    time.Now().UTC(),
			Headers: amqp.Table{
				"retry_count": 0,
			},
		},
	)

	if err != nil {
		p.logger.WithError(err).Error("Failed to publish payment message")
		return err
	}

	p.logger.WithField("payment_id", paymentID).Debug("Payment message published to RabbitMQ")
	return nil
}

type PaymentMessage struct {
	PaymentID uuid.UUID `json:"payment_id"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}
