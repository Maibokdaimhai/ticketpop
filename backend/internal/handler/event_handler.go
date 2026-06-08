package handler

import (
	"net/http"
	"time"

	"github.com/Maibokdaimhai/ticketpop/internal/service"
	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	eventService *service.EventService
}

func NewEventHandler(eventService *service.EventService) *EventHandler {
	return &EventHandler{eventService: eventService}
}

type CreateEventRequest struct {
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description" binding:"required"`
	Place       string    `json:"place" binding:"required"`
	ShowTime    time.Time `json:"show_time" binding:"required"`
	ImgURL      string    `json:"img_url"`
	OrganizerID string    `json:"organizer_id" binding:"required"`
	CategoryID  *string    `json:"category_id"`
}

func (h *EventHandler) CreateEvent(c *gin.Context) {
	var req CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	result, err := h.eventService.CreateEvent(c.Request.Context(), service.CreateEventParams{
		Title:       req.Title,
		Description: req.Description,
		Place:       req.Place,
		ShowTime:    req.ShowTime,
		ImgURL:      req.ImgURL,
		OrganizerID: req.OrganizerID,
		CategoryID:  req.CategoryID,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Create Event Success",
		"event":   result,
	})
}

func (h *EventHandler) GetEvents(c *gin.Context) {
	events, err := h.eventService.GetEvents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Get Events Success",
		"events":  events,
	})
}

func (h *EventHandler) GetEventByID(c *gin.Context) {
	id := c.Param("id")
	event, err := h.eventService.GetEventByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Get Event By ID Success",
		"event":   event,
	})
}
