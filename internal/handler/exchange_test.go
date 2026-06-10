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
	"github.com/stretchr/testify/require"
)

func setupExchangeRouter(exchangeSvc service.ExchangeService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := handler.NewExchangeHandler(exchangeSvc)
	r.POST("/exchange/convert", h.Convert)
	return r
}

func TestExchangeHandler_Convert(t *testing.T) {
	tests := []struct {
		name       string
		mock       *mockExchangeService
		body       interface{}
		wantStatus int
	}{
		{
			name: "success",
			mock: &mockExchangeService{
				convertFunc: func(_ context.Context, _ *model.ConvertRequest) (*model.ConvertResponse, error) {
					return &model.ConvertResponse{
						From:      "USD",
						To:        "EUR",
						Amount:    100,
						Rate:      0.92,
						Converted: 92,
					}, nil
				},
			},
			body:       model.ConvertRequest{From: "USD", To: "EUR", Amt: 100},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid JSON",
			mock:       &mockExchangeService{},
			body:       "not json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing required fields",
			mock: &mockExchangeService{},
			body: model.ConvertRequest{
				From: "USD",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "unsupported currency",
			mock: &mockExchangeService{
				convertFunc: func(_ context.Context, _ *model.ConvertRequest) (*model.ConvertResponse, error) {
					return nil, service.ErrUnsupportedCurrency
				},
			},
			body:       model.ConvertRequest{From: "USD", To: "XYZ", Amt: 100},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "API error",
			mock: &mockExchangeService{
				convertFunc: func(_ context.Context, _ *model.ConvertRequest) (*model.ConvertResponse, error) {
					return nil, errors.New("connection refused")
				},
			},
			body:       model.ConvertRequest{From: "USD", To: "EUR", Amt: 100},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupExchangeRouter(tt.mock)

			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/exchange/convert", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var resp model.ConvertResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Equal(t, "USD", resp.From)
				assert.Equal(t, "EUR", resp.To)
				assert.Equal(t, float64(100), resp.Amount)
				assert.Equal(t, float64(92), resp.Converted)
			}
		})
	}
}
