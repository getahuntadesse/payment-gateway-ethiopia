package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"payment-gateway/internal/api"
	"payment-gateway/internal/config"
	"payment-gateway/internal/messaging"
	"payment-gateway/internal/repository"
	"payment-gateway/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger with Ethiopian context
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05 EAT",
	})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)

	// Ethiopian time (GMT+3)
	ethiopianTime := time.Now().Add(3 * time.Hour)

	logger.Info("Starting Ethiopian Payment Gateway API...")
	logger.Info("የኢትዮጵያ ክፍያ ግብይት መተግበሪያ እየተጀመረ ነው...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration: ", err)
	}

	// Database connection
	dbDSN := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	// Connect to database
	dbPool, err := pgxpool.New(context.Background(), dbDSN)
	if err != nil {
		logger.Fatal("Failed to connect to database: ", err)
	}
	defer dbPool.Close()

	// Test database connection
	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := dbPool.Ping(pingCtx); err != nil {
		logger.Fatal("Database ping failed: ", err)
	}

	logger.Info("Connected to PostgreSQL database successfully")

	// RabbitMQ connection
	rabbitConfig := messaging.RabbitMQConfig{
		URL:           cfg.RabbitMQ.URL,
		QueueName:     cfg.RabbitMQ.QueueName,
		Exchange:      cfg.RabbitMQ.Exchange,
		ConsumerTag:   cfg.RabbitMQ.ConsumerTag,
		PrefetchCount: cfg.RabbitMQ.PrefetchCount,
	}

	rabbitClient, err := messaging.NewRabbitMQClient(rabbitConfig, logger)
	if err != nil {
		logger.Fatal("Failed to connect to RabbitMQ: ", err)
	}
	defer rabbitClient.Close()

	logger.Info("Connected to RabbitMQ successfully")

	// Initialize dependencies
	paymentRepo := repository.NewPaymentRepository(dbPool, logger)
	publisher := messaging.NewPaymentPublisher(rabbitClient, logger)
	paymentService := service.NewPaymentService(paymentRepo, publisher, logger)

	// Create and start server
	server := api.NewServer(cfg, paymentService, logger)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.Start(); err != nil {
			logger.Fatal("Server failed to start: ", err)
		}
	}()

	// Update Ethiopian time for final log
	ethiopianTime = time.Now().Add(3 * time.Hour)
	logger.WithFields(logrus.Fields{
		"port":           cfg.Server.Port,
		"environment":    cfg.App.Environment,
		"ethiopian_time": ethiopianTime.Format("15:04:05"),
	}).Info("Ethiopian Payment Gateway API is running")

	<-quit
	logger.Info("Shutting down Ethiopian Payment Gateway API...")

	// For now, just wait a moment for graceful shutdown
	// In a real implementation, you would use the context to shutdown the server
	time.Sleep(2 * time.Second)

	logger.Info("Ethiopian Payment Gateway API stopped successfully")
}
