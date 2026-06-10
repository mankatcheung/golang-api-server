// Package handler implements HTTP request handlers for the API endpoints.
package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-api-server/internal/model"
	"github.com/golang-api-server/internal/service"
)

// ExchangeHandler handles currency exchange HTTP requests.
type ExchangeHandler struct {
	exchangeService service.ExchangeService
}

// NewExchangeHandler returns a new ExchangeHandler.
func NewExchangeHandler(exchangeService service.ExchangeService) *ExchangeHandler {
	return &ExchangeHandler{exchangeService: exchangeService}
}

func (h *ExchangeHandler) Convert(c *gin.Context) {
	var req model.ConvertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	resp, err := h.exchangeService.Convert(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrUnsupportedCurrency) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch exchange rates"})
		return
	}

	c.JSON(http.StatusOK, resp)
}
