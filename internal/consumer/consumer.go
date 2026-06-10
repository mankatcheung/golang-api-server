// Package consumer provides Kafka consumer lifecycle management.
package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/golang-api-server/internal/model"
	"github.com/golang-api-server/pkg/kafka"
)

// MessageConsumer defines the interface for consuming messages.
type MessageConsumer interface {
	Close() error
}

// StartConversionConsumer launches a background goroutine that processes conversion events.
func StartConversionConsumer(brokers []string, topic, groupID string) MessageConsumer {
	c := kafka.NewConsumer(brokers, topic, groupID)

	go func() {
		if err := c.Consume(context.Background(), handleConversionEvent); err != nil {
			slog.Error("conversion consumer stopped", "error", err)
		}
	}()

	return c
}

func handleConversionEvent(ctx context.Context, key, value []byte) error {
	var event model.Event
	if err := json.Unmarshal(value, &event); err != nil {
		return err
	}

	slog.Info("processing conversion event", "id", event.ID, "type", event.Type)
	return nil
}
