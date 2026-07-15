package handler

import (
	"net/http"

	"github.com/Maibokdaimhai/ticketpop/internal/service"
	"github.com/gin-gonic/gin"
)

type BookingHandler struct {
	bookingService *service.BookingService
}

func NewBookingHandler(bookingService *service.BookingService) *BookingHandler {
	return &BookingHandler{bookingService: bookingService}
}

type CreateBookingRequest struct {
	TicketPoolID string `json:"ticket_pool_id" binding:"required"`
	Quantity     int    `json:"quantity"       binding:"required,min=1,max=10"`
}

func (h *BookingHandler) CreateBooking(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	var req CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	result, err := h.bookingService.CreateBooking(c.Request.Context(), service.CreateBookingParams{
		UserID:       userID,
		TicketPoolID: req.TicketPoolID,
		Quantity:     req.Quantity,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "จองตั๋วสำเร็จ",
		"data":    result,
	})
}

func (h *BookingHandler) GetBookings(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	bookings, err := h.bookingService.GetBookingsByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "ดึงข้อมูล booking สำเร็จ",
		"bookings": bookings,
	})
}

func (h *BookingHandler) GetBookingByID(c *gin.Context) {
	id := c.Param("id")

	booking, err := h.bookingService.GetBookingByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "ไม่พบ booking",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ดึงข้อมูล booking สำเร็จ",
		"booking": booking,
	})
}
