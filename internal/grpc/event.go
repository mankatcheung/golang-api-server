// Package grpc provides gRPC server implementations for the API services.
package grpc

import (
	"context"
	"fmt"

	"github.com/golang-api-server/internal/model"
	"github.com/golang-api-server/internal/service"
	pb "github.com/golang-api-server/proto/event"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// EventGRPCHandler implements the event gRPC service.
type EventGRPCHandler struct {
	pb.UnimplementedEventServiceServer
	eventService service.EventPublisher
}

// NewEventGRPCHandler returns a new EventGRPCHandler.
func NewEventGRPCHandler(eventService service.EventPublisher) *EventGRPCHandler {
	return &EventGRPCHandler{eventService: eventService}
}

// PublishConversion publishes a currency conversion event to Kafka.
func (h *EventGRPCHandler) PublishConversion(ctx context.Context, req *pb.ConversionEvent) (*pb.PublishResponse, error) {
	err := h.eventService.PublishConversion(ctx, &model.ConversionEvent{
		From:      req.From,
		To:        req.To,
		Amount:    req.Amount,
		Rate:      req.Rate,
		Converted: req.Converted,
		UserID:    req.UserId,
	})
	if err != nil {
		return &pb.PublishResponse{
			Id:      uuid.New().String(),
			Success: false,
			Message: err.Error(),
		}, nil
	}
	return &pb.PublishResponse{
		Id:      uuid.New().String(),
		Success: true,
		Message: "conversion event published",
	}, nil
}

// PublishEvent publishes a generic event by extracting from/to/amount from the payload.
func (h *EventGRPCHandler) PublishEvent(ctx context.Context, req *pb.PublishRequest) (*pb.PublishResponse, error) {
	payload := req.GetPayload()
	from := payload["from"]
	to := payload["to"]
	amountStr := payload["amount"]
	if from == "" || to == "" || amountStr == "" {
		return nil, status.Error(codes.InvalidArgument, "payload must contain 'from', 'to', and 'amount'")
	}

	var amount float64
	if _, err := fmt.Sscanf(amountStr, "%f", &amount); err != nil {
		return nil, status.Error(codes.InvalidArgument, "'amount' must be a numeric string")
	}

	if err := h.eventService.PublishConversion(ctx, &model.ConversionEvent{
		From:   from,
		To:     to,
		Amount: amount,
	}); err != nil {
		return &pb.PublishResponse{Id: uuid.New().String(), Success: false, Message: err.Error()}, nil
	}
	return &pb.PublishResponse{Id: uuid.New().String(), Success: true, Message: "event published"}, nil
}
