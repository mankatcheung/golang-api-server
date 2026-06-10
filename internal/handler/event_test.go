package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-api-server/internal/handler"
	"github.com/golang-api-server/internal/model"
	"github.com/golang-api-server/internal/service"
	"github.com/stretchr/testify/assert"
)

func setupEventRouter(eventSvc service.EventPublisher) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := handler.NewEventHandler(eventSvc)
	r.POST("/events/publish", h.PublishEvent)
	return r
}

func TestEventHandler_PublishEvent(t *testing.T) {
	tests := []struct {
		name       string
		mock       *mockEventPublisher
		body       interface{}
		wantStatus int
	}{
		{
			name: "success",
			mock: &mockEventPublisher{
				publishConversionFunc: func(_ context.Context, _ *model.ConversionEvent) error {
					return nil
				},
			},
			body: map[string]interface{}{
				"type": "conversion.completed",
				"payload": map[string]interface{}{
					"from":   "USD",
					"to":     "EUR",
					"amount": 100.0,
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid JSON",
			mock:       &mockEventPublisher{},
			body:       "not json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing type field",
			mock: &mockEventPublisher{},
			body: map[string]interface{}{
				"payload": map[string]interface{}{},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "publish error",
			mock: &mockEventPublisher{
				publishConversionFunc: func(_ context.Context, _ *model.ConversionEvent) error {
					return errors.New("kafka unavailable")
				},
			},
			body: map[string]interface{}{
				"type": "conversion.completed",
				"payload": map[string]interface{}{
					"from":   "USD",
					"to":     "EUR",
					"amount": 100.0,
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupEventRouter(tt.mock)

			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/events/publish", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
