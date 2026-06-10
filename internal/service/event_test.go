package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang-api-server/internal/model"
	"github.com/golang-api-server/internal/service"
	"github.com/stretchr/testify/require"
)

func TestEventService_PublishConversion(t *testing.T) {
	tests := []struct {
		name     string
		producer *mockProducer
		event    *model.ConversionEvent
		wantErr  bool
	}{
		{
			name: "success",
			producer: &mockProducer{
				publishFunc: func(_ context.Context, key, value interface{}) error {
					return nil
				},
			},
			event: &model.ConversionEvent{
				From:      "USD",
				To:        "EUR",
				Amount:    100,
				Rate:      0.92,
				Converted: 92,
				UserID:    1,
			},
			wantErr: false,
		},
		{
			name: "publish error",
			producer: &mockProducer{
				publishFunc: func(_ context.Context, _, _ interface{}) error {
					return errors.New("kafka unavailable")
				},
			},
			event: &model.ConversionEvent{
				From:      "USD",
				To:        "EUR",
				Amount:    100,
				Rate:      0.92,
				Converted: 92,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := service.NewEventService(tt.producer)
			err := svc.PublishConversion(context.Background(), tt.event)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
