package handlers

import (
	"net/http"
	"strconv"
	"time"

	"payment-gateway/internal/domain"
	"payment-gateway/internal/service"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type PaymentHandler struct {
	paymentService service.PaymentService
	logger         *logrus.Logger
}

func NewPaymentHandler(paymentService service.PaymentService, logger *logrus.Logger) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
		logger:         logger,
	}
}

// CreatePayment handles Ethiopian payment creation
// @Summary Create a new Ethiopian payment
// @Description Create a payment in ETB or USD with Ethiopian context
// @Tags payments
// @Accept json
// @Produce json
// @Param payment body domain.CreatePaymentRequest true "Payment details"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /payments [post]
func (h *PaymentHandler) CreatePayment(c echo.Context) error {
	var req domain.CreatePaymentRequest
	if err := c.Bind(&req); err != nil {
		h.logger.WithError(err).Error("Failed to bind request")
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Log Ethiopian payment attempt
	h.logger.WithFields(logrus.Fields{
		"reference":     req.Reference,
		"currency":      req.Currency,
		"amount":        req.Amount,
		"customer_name": req.CustomerName,
		"bank_code":     req.BankCode,
	}).Info("Ethiopian payment creation request")

	payment, err := h.paymentService.CreatePayment(c.Request().Context(), req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create payment")

		switch err {
		case domain.ErrInvalidInput:
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error":   "Invalid input data",
				"details": err.Error(),
			})
		case domain.ErrPaymentAlreadyExists:
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "Payment with this reference already exists",
			})
		case domain.ErrBusinessHours:
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Payments can only be processed during Ethiopian business hours (8:00 AM - 5:00 PM EAT)",
			})
		case domain.ErrAmountTooLarge:
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Amount exceeds Ethiopian regulatory limit (1,000,000 ETB)",
			})
		default:
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to create payment",
			})
		}
	}

	// Return Ethiopian response
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message":        "የክፍያ ሂደት ተጀምሯል (Payment process initiated)",
		"payment_id":     payment.ID,
		"status":         payment.Status,
		"reference":      payment.Reference,
		"created_at":     payment.CreatedAt.Format("2006-01-02 15:04:05 MST"),
		"ethiopian_time": payment.CreatedAt.Add(3 * time.Hour).Format("2006-01-02 15:04:05 EAT"),
	})
}

// GetPayment retrieves payment details
// @Summary Get payment details
// @Description Get details of a specific payment by ID
// @Tags payments
// @Produce json
// @Param id path string true "Payment ID"
// @Success 200 {object} domain.PaymentResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /payments/{id} [get]
func (h *PaymentHandler) GetPayment(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid payment ID format",
		})
	}

	payment, err := h.paymentService.GetPayment(c.Request().Context(), id)
	if err != nil {
		if err == domain.ErrPaymentNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Payment not found",
			})
		}
		h.logger.WithError(err).Error("Failed to get payment")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve payment",
		})
	}

	return c.JSON(http.StatusOK, payment.ToResponse())
}

// GetPaymentByReference retrieves payment by reference number
// @Summary Get payment by reference
// @Description Get payment details by Ethiopian reference number
// @Tags payments
// @Produce json
// @Param reference query string true "Payment Reference"
// @Success 200 {object} domain.PaymentResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /payments/by-reference [get]
func (h *PaymentHandler) GetPaymentByReference(c echo.Context) error {
	reference := c.QueryParam("reference")
	if reference == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Reference parameter is required",
		})
	}

	payment, err := h.paymentService.GetPaymentByReference(c.Request().Context(), reference)
	if err != nil {
		if err == domain.ErrPaymentNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Payment not found with reference: " + reference,
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve payment",
		})
	}

	return c.JSON(http.StatusOK, payment.ToResponse())
}

// ListPayments retrieves paginated list of payments
// @Summary List payments
// @Description Get paginated list of Ethiopian payments
// @Tags payments
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /payments [get]
func (h *PaymentHandler) ListPayments(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	payments, total, err := h.paymentService.ListPayments(c.Request().Context(), page, limit)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list payments")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to list payments",
		})
	}

	// Convert to responses
	responses := make([]domain.PaymentResponse, len(payments))
	for i, payment := range payments {
		responses[i] = payment.ToResponse()
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"payments": responses,
		"total":    total,
		"page":     page,
		"limit":    limit,
		"has_more": total > page*limit,
	})
}

// GetStatistics retrieves Ethiopian payment statistics
// @Summary Get payment statistics
// @Description Get statistics about Ethiopian payments
// @Tags statistics
// @Produce json
// @Success 200 {object} service.PaymentStatistics
// @Failure 500 {object} map[string]string
// @Router /statistics [get]
func (h *PaymentHandler) GetStatistics(c echo.Context) error {
	stats, err := h.paymentService.GetStatistics(c.Request().Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get statistics")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get statistics",
		})
	}

	return c.JSON(http.StatusOK, stats)
}

// HealthCheck handles health checks
// @Summary Health check
// @Description Check if the Ethiopian Payment Gateway is healthy
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (h *PaymentHandler) HealthCheck(c echo.Context) error {
	ethiopianTime := time.Now().Add(3 * time.Hour) // GMT+3

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":         "healthy",
		"service":        "Ethiopian Payment Gateway",
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
		"ethiopian_time": ethiopianTime.Format("2006-01-02 15:04:05 EAT"),
		"version":        "1.0.0",
	})
}

// EthiopianBankList returns list of Ethiopian banks
// @Summary Get Ethiopian banks
// @Description Get list of Ethiopian banks for payment processing
// @Tags banks
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /banks [get]
func (h *PaymentHandler) EthiopianBankList(c echo.Context) error {
	banks := []map[string]string{
		{
			"code":  "CBE",
			"name":  "Commercial Bank of Ethiopia",
			"swift": "CBETETAA",
		},
		{
			"code":  "AWASH",
			"name":  "Awash Bank",
			"swift": "AWINETAA",
		},
		{
			"code":  "DASHEN",
			"name":  "Dashen Bank",
			"swift": "DASHETAA",
		},
		{
			"code":  "ABYSSINIA",
			"name":  "Bank of Abyssinia",
			"swift": "ABYSETAA",
		},
		{
			"code":  "NIB",
			"name":  "Nib International Bank",
			"swift": "NIBIETAA",
		},
		{
			"code":  "UNITED",
			"name":  "United Bank",
			"swift": "UBNIETAA",
		},
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"banks":   banks,
		"message": "የኢትዮጵያ ባንኮች ዝርዝር (List of Ethiopian Banks)",
	})
}
