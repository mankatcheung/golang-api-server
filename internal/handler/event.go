// Package handler implements HTTP request handlers for the API endpoints.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-api-server/internal/model"
	"github.com/golang-api-server/internal/service"
)

// EventHandler handles event publishing HTTP requests.
type EventHandler struct {
	eventService service.EventPublisher
}

// NewEventHandler returns a new EventHandler.
func NewEventHandler(eventService service.EventPublisher) *EventHandler {
	return &EventHandler{eventService: eventService}
}

func (h *EventHandler) PublishEvent(c *gin.Context) {
	var req struct {
		Type    string                 `json:"type" binding:"required"`
		Payload map[string]interface{} `json:"payload" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	from, ok := req.Payload["from"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payload 'from' must be a string"})
		return
	}
	to, ok := req.Payload["to"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payload 'to' must be a string"})
		return
	}
	amount, ok := req.Payload["amount"].(float64)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payload 'amount' must be a number"})
		return
	}

	if err := h.eventService.PublishConversion(c.Request.Context(), &model.ConversionEvent{
		From:   from,
		To:     to,
		Amount: amount,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "event published"})
}
