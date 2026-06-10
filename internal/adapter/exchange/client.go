// Package exchange implements the exchange rate client adapter.
package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const apiBase = "https://open.er-api.com/v6/latest"

const defaultHTTPTimeout = 10 * time.Second

// HTTPClient fetches exchange rates from an external API.
type HTTPClient struct {
	client *http.Client
}

// NewHTTPClient returns an HTTPClient. Pass 0 to use the default 10-second timeout.
func NewHTTPClient(timeout time.Duration) *HTTPClient {
	if timeout == 0 {
		timeout = defaultHTTPTimeout
	}
	return &HTTPClient{
		client: &http.Client{Timeout: timeout},
	}
}

// GetRates returns the exchange rates for the given base currency.
func (c *HTTPClient) GetRates(ctx context.Context, base string) (map[string]float64, error) {
	url := fmt.Sprintf("%s/%s", apiBase, strings.ToUpper(base))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch rates: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("exchange API returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Result string             `json:"result"`
		Base   string             `json:"base_code"`
		Rates  map[string]float64 `json:"rates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if result.Result != "success" {
		return nil, fmt.Errorf("exchange API error: %s", result.Result)
	}

	return result.Rates, nil
}
