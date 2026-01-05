package api

import (
	"net/http"
	"strconv"
	"time"

	"payment-gateway/internal/api/handlers"
	"payment-gateway/internal/config"
	"payment-gateway/internal/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

type Server struct {
	e      *echo.Echo
	logger *logrus.Logger
	cfg    *config.Config
}

func NewServer(cfg *config.Config, paymentService service.PaymentService, logger *logrus.Logger) *Server {
	e := echo.New()

	// Hide banner
	e.HideBanner = true

	// Middleware
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
	}))

	// Request logging middleware
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `${time_rfc3339} | ${status} | ${latency_human} | ${remote_ip} | ${method} ${uri}` + "\n",
		Output: logger.Writer(),
	}))

	// Create handlers
	paymentHandler := handlers.NewPaymentHandler(paymentService, logger)

	// Routes
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message":     "Ethiopian Payment Gateway API",
			"version":     "1.0.0",
			"description": "Ethiopian Payment Gateway",
			"docs":        "/api/v1/docs",
		})
	})

	// API v1 routes
	v1 := e.Group("/api/v1")
	{
		// Health check
		v1.GET("/health", paymentHandler.HealthCheck)

		// Ethiopian banks
		v1.GET("/banks", paymentHandler.EthiopianBankList)

		// Payment routes
		payments := v1.Group("/payments")
		{
			payments.POST("", paymentHandler.CreatePayment)
			payments.GET("", paymentHandler.ListPayments)
			payments.GET("/by-reference", paymentHandler.GetPaymentByReference)
			payments.GET("/:id", paymentHandler.GetPayment)
		}

		// Statistics
		v1.GET("/statistics", paymentHandler.GetStatistics)

		// Documentation
		v1.GET("/docs", func(c echo.Context) error {
			docs := `
Ethiopian Payment Gateway API Documentation

Available Endpoints:
GET  /api/v1/health            - Service health check
GET  /api/v1/banks             - List of Ethiopian banks
POST /api/v1/payments          - Create new payment
GET  /api/v1/payments          - List all payments (paginated)
GET  /api/v1/payments/:id      - Get payment by ID
GET  /api/v1/payments/by-reference - Get payment by reference
GET  /api/v1/statistics        - Get payment statistics

Sample Ethiopian Payment Request:
{
  "amount": 1500.75,
  "currency": "ETB",
  "reference": "CBE-2023-001234",
  "description": "Payment for Addis Ababa coffee export",
  "customer_name": "Test Customer",
  "bank_code": "CBE"
}

Currencies: ETB (Ethiopian Birr) or USD
Business Hours: 8:00 AM - 5:00 PM Ethiopian Time (GMT+3)
			`
			return c.String(http.StatusOK, docs)
		})
	}

	return &Server{
		e:      e,
		logger: logger,
		cfg:    cfg,
	}
}

func (s *Server) Start() error {
	addr := ":" + strconv.Itoa(s.cfg.Server.Port)

	// Ethiopian time (GMT+3)
	ethiopianTime := time.Now().Add(3 * time.Hour)

	s.logger.WithFields(logrus.Fields{
		"port":           s.cfg.Server.Port,
		"environment":    s.cfg.App.Environment,
		"name":           s.cfg.App.Name,
		"ethiopian_time": ethiopianTime.Format("15:04:05"),
	}).Info("Starting Ethiopian Payment Gateway API")

	return s.e.Start(addr)
}
