package service

import (
	"context"
	"time"

	"github.com/golang-api-server/internal/model"
	"github.com/google/uuid"
)

// MessageProducer defines the port for publishing messages.
type MessageProducer interface {
	Publish(ctx context.Context, key, value interface{}) error
	Close() error
}

// Compile-time check that eventServiceImpl implements EventPublisher.
var _ EventPublisher = (*eventServiceImpl)(nil)

// eventServiceImpl publishes events to a message broker.
type eventServiceImpl struct {
	producer MessageProducer
}

// NewEventService returns an EventPublisher backed by the given producer.
func NewEventService(producer MessageProducer) EventPublisher {
	return &eventServiceImpl{producer: producer}
}

// PublishConversion wraps a conversion result in an Event and publishes it.
func (s *eventServiceImpl) PublishConversion(ctx context.Context, event *model.ConversionEvent) error {
	evt := model.Event{
		ID:        uuid.New().String(),
		Type:      "conversion.completed",
		Payload: map[string]interface{}{
			"from":      event.From,
			"to":        event.To,
			"amount":    event.Amount,
			"rate":      event.Rate,
			"converted": event.Converted,
			"user_id":   event.UserID,
		},
		Timestamp: time.Now(),
	}

	return s.producer.Publish(ctx, evt.ID, evt)
}
