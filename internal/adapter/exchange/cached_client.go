package exchange

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/golang-api-server/internal/cache"
)

// CachedHTTPClient wraps HTTPClient with Redis caching for exchange rates.
type CachedHTTPClient struct {
	underlying *HTTPClient
	cache      cache.Cache
	ttl        time.Duration
}

// NewCachedHTTPClient returns a CachedHTTPClient backed by the given cache.
func NewCachedHTTPClient(underlying *HTTPClient, c cache.Cache, ttl time.Duration) *CachedHTTPClient {
	return &CachedHTTPClient{
		underlying: underlying,
		cache:      c,
		ttl:        ttl,
	}
}

// GetRates returns cached exchange rates, falling back to the underlying HTTP client.
func (c *CachedHTTPClient) GetRates(ctx context.Context, base string) (map[string]float64, error) {
	key := fmt.Sprintf("exchange:rates:%s", strings.ToUpper(base))

	var rates map[string]float64
	if err := c.cache.Get(ctx, key, &rates); err == nil {
		return rates, nil
	}

	rates, err := c.underlying.GetRates(ctx, base)
	if err != nil {
		return nil, err
	}

	if err := c.cache.Set(ctx, key, rates, c.ttl); err != nil {
		slog.Debug("cache set failed", "key", key, "error", err)
	}
	return rates, nil
}
