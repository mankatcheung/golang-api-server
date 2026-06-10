//go:build e2e

package e2e_test

import (
	"net/http"
	"testing"

	"github.com/golang-api-server/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestExchangeConvert(t *testing.T) {
	truncateUsers(t)
	registerResp := postJSON(t, "/api/v1/auth/register", map[string]string{
		"email": "exchange@example.com", "password": "password123", "name": "Exchange User",
	})
	var auth model.AuthResponse
	decodeJSON(t, registerResp.Body, &auth)
	registerResp.Body.Close()

	t.Run("valid USD to EUR", func(t *testing.T) {
		resp := authedPost(t, "/api/v1/exchange/convert", auth.AccessToken, map[string]interface{}{
			"from": "USD", "to": "EUR", "amount": 100.0,
		})
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusOK)

		var result model.ConvertResponse
		decodeJSON(t, resp.Body, &result)
		assert.Equal(t, "USD", result.From)
		assert.Equal(t, "EUR", result.To)
		assert.Equal(t, 100.0, result.Amount)
		assert.Greater(t, result.Rate, 0.0)
		assert.Greater(t, result.Converted, 0.0)
	})

	t.Run("unsupported currency", func(t *testing.T) {
		resp := authedPost(t, "/api/v1/exchange/convert", auth.AccessToken, map[string]interface{}{
			"from": "USD", "to": "XXX", "amount": 100.0,
		})
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("missing from field", func(t *testing.T) {
		resp := authedPost(t, "/api/v1/exchange/convert", auth.AccessToken, map[string]interface{}{
			"to": "EUR", "amount": 100.0,
		})
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("zero amount", func(t *testing.T) {
		resp := authedPost(t, "/api/v1/exchange/convert", auth.AccessToken, map[string]interface{}{
			"from": "USD", "to": "EUR", "amount": 0,
		})
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("no authorization", func(t *testing.T) {
		resp := authedPost(t, "/api/v1/exchange/convert", "", map[string]interface{}{
			"from": "USD", "to": "EUR", "amount": 100.0,
		})
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusUnauthorized)
	})
}
