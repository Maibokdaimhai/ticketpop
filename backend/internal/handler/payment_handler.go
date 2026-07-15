package handler

import (
	"net/http"

	"github.com/Maibokdaimhai/ticketpop/internal/service"
	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService *service.PaymentService
}

func NewPaymentHandler(paymentService *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

type CreatePaymentRequest struct {
	BookingID string `json:"booking_id" binding:"required"`
}

func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	payment, err := h.paymentService.CreatePayment(c.Request.Context(), service.CreatePaymentParams{
		BookingID: req.BookingID,
		UserID:    userID,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "ชำระเงินสำเร็จ",
		"payment": payment,
	})
}

func (h *PaymentHandler) GetPaymentByBookingID(c *gin.Context) {
	bookingID := c.Param("booking_id")

	payment, err := h.paymentService.GetPaymentByBookingID(c.Request.Context(), bookingID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "ไม่พบข้อมูลการชำระเงิน",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ดึงข้อมูลการชำระเงินสำเร็จ",
		"payment": payment,
	})
}