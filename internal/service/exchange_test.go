package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang-api-server/internal/model"
	"github.com/golang-api-server/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockExchangeClient struct {
	rates map[string]float64
	err   error
}

func (m *mockExchangeClient) GetRates(_ context.Context, _ string) (map[string]float64, error) {
	return m.rates, m.err
}

func TestExchangeService_Convert(t *testing.T) {
	client := &mockExchangeClient{
		rates: map[string]float64{
			"EUR": 0.92,
			"GBP": 0.79,
			"JPY": 149.5,
		},
	}
	svc := service.NewExchangeService(client)

	tests := []struct {
		name      string
		req       *model.ConvertRequest
		wantRate  float64
		wantConv  float64
		wantErr   bool
		errIs     error
	}{
		{
			name:     "USD to EUR",
			req:      &model.ConvertRequest{From: "USD", To: "EUR", Amt: 100},
			wantRate: 0.92,
			wantConv: 92,
		},
		{
			name:     "usd to jpy lowercase",
			req:      &model.ConvertRequest{From: "usd", To: "jpy", Amt: 50},
			wantRate: 149.5,
			wantConv: 7475,
		},
		{
			name:    "unsupported currency",
			req:     &model.ConvertRequest{From: "USD", To: "XYZ", Amt: 100},
			wantErr: true,
			errIs:   service.ErrUnsupportedCurrency,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.Convert(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errIs != nil {
					assert.True(t, errors.Is(err, tt.errIs))
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantRate, resp.Rate)
			assert.Equal(t, tt.wantConv, resp.Converted)
			assert.Equal(t, tt.req.Amt, resp.Amount)
		})
	}
}

func TestExchangeService_Convert_APIError(t *testing.T) {
	client := &mockExchangeClient{
		err: errors.New("connection refused"),
	}
	svc := service.NewExchangeService(client)

	_, err := svc.Convert(context.Background(), &model.ConvertRequest{
		From: "USD",
		To:   "EUR",
		Amt:  100,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}
