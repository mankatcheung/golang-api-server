package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/golang-api-server/internal/model"
)

var ErrUnsupportedCurrency = errors.New("unsupported currency")

// ExchangeRateClient defines the port for fetching exchange rates.
type ExchangeRateClient interface {
	GetRates(ctx context.Context, base string) (map[string]float64, error)
}

// Compile-time check that exchangeServiceImpl implements ExchangeService.
var _ ExchangeService = (*exchangeServiceImpl)(nil)

type exchangeServiceImpl struct {
	client ExchangeRateClient
}

// NewExchangeService returns an ExchangeService backed by the given client.
func NewExchangeService(client ExchangeRateClient) ExchangeService {
	return &exchangeServiceImpl{client: client}
}

func (s *exchangeServiceImpl) GetRates(ctx context.Context, base string) (map[string]float64, error) {
	return s.client.GetRates(ctx, strings.ToUpper(base))
}

func (s *exchangeServiceImpl) Convert(ctx context.Context, req *model.ConvertRequest) (*model.ConvertResponse, error) {
	from := strings.ToUpper(req.From)
	to := strings.ToUpper(req.To)

	rates, err := s.client.GetRates(ctx, from)
	if err != nil {
		return nil, fmt.Errorf("get rates for %s: %w", from, err)
	}

	rate, ok := rates[to]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedCurrency, to)
	}

	return &model.ConvertResponse{
		From:      from,
		To:        to,
		Amount:    req.Amt,
		Rate:      rate,
		Converted: req.Amt * rate,
	}, nil
}
