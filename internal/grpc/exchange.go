// Package grpc provides gRPC server implementations for the API services.
package grpc

import (
	"context"
	"errors"

	"github.com/golang-api-server/internal/model"
	"github.com/golang-api-server/internal/service"
	pb "github.com/golang-api-server/proto/exchange"
)

// ExchangeGRPCHandler implements the exchange gRPC service.
type ExchangeGRPCHandler struct {
	pb.UnimplementedExchangeServiceServer
	exchangeService service.ExchangeService
}

// NewExchangeGRPCHandler returns a new ExchangeGRPCHandler.
func NewExchangeGRPCHandler(exchangeService service.ExchangeService) *ExchangeGRPCHandler {
	return &ExchangeGRPCHandler{exchangeService: exchangeService}
}

// Convert converts an amount from one currency to another.
func (h *ExchangeGRPCHandler) Convert(ctx context.Context, req *pb.ConvertRequest) (*pb.ConvertResponse, error) {
	resp, err := h.exchangeService.Convert(ctx, &model.ConvertRequest{
		From: req.From,
		To:   req.To,
		Amt:  req.Amount,
	})
	if err != nil {
		if errors.Is(err, service.ErrUnsupportedCurrency) {
			return nil, errors.New("unsupported currency")
		}
		return nil, err
	}
	return &pb.ConvertResponse{
		From:      resp.From,
		To:        resp.To,
		Amount:    resp.Amount,
		Rate:      resp.Rate,
		Converted: resp.Converted,
	}, nil
}

// GetRates returns exchange rates for a base currency.
func (h *ExchangeGRPCHandler) GetRates(ctx context.Context, req *pb.RatesRequest) (*pb.RatesResponse, error) {
	rates, err := h.exchangeService.GetRates(ctx, req.Base)
	if err != nil {
		return nil, err
	}
	return &pb.RatesResponse{
		Base:  req.Base,
		Rates: rates,
	}, nil
}
