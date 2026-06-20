package handler

import (
	"net/http"

	"github.com/Maibokdaimhai/ticketpop/internal/service"
	"github.com/gin-gonic/gin"
)

type TicketPoolHandler struct {
	ticketPoolService *service.TicketPoolService
}

func NewTicketPoolHandler(ticketPoolService *service.TicketPoolService) *TicketPoolHandler {
	return &TicketPoolHandler{ticketPoolService: ticketPoolService}
}

type CreateTicketPoolRequest struct {
	ZoneName      string  `json:"zone_name" binding:"required"`
	Price         float64 `json:"price" binding:"required"`
	TotalCapacity int     `json:"total_capacity" binding:"required"`
	DisplayName   string  `json:"display_name"`
}

func (h *TicketPoolHandler) CreateTicketPool(c *gin.Context){
	eventID := c.Param("id")

	var req CreateTicketPoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	result, err := h.ticketPoolService.CreateTicketPool(c.Request.Context(), service.CreateTicketPoolParams{
		ZoneName: req.ZoneName,
		Price: req.Price,
		TotalCapacity: req.TotalCapacity,
		DisplayName: req.DisplayName,
		EventID: eventID,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "Create Ticket Pool Success",
		"ticket_pool":   result,
	})
}


func (h *TicketPoolHandler) GetTicketPoolsByEventID(c *gin.Context) {
	id := c.Param("id")
	ticketPool, err := h.ticketPoolService.GetTicketPoolsByEventID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Get Ticket Pool By EventID Success",
		"ticket_pool":   ticketPool,
	})
}
