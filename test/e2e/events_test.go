//go:build e2e

package e2e_test

import (
	"net/http"
	"testing"

	"github.com/golang-api-server/internal/model"
)

func TestPublishEvent(t *testing.T) {
	truncateUsers(t)
	registerResp := postJSON(t, "/api/v1/auth/register", map[string]string{
		"email": "events@example.com", "password": "password123", "name": "Events User",
	})
	var auth model.AuthResponse
	decodeJSON(t, registerResp.Body, &auth)
	registerResp.Body.Close()

	validPayload := map[string]interface{}{
		"type": "conversion",
		"payload": map[string]interface{}{
			"from":   "USD",
			"to":     "EUR",
			"amount": 50.0,
		},
	}

	t.Run("success", func(t *testing.T) {
		resp := authedPost(t, "/api/v1/events/publish", auth.AccessToken, validPayload)
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusOK)
	})

	t.Run("missing type field", func(t *testing.T) {
		resp := authedPost(t, "/api/v1/events/publish", auth.AccessToken, map[string]interface{}{
			"payload": map[string]interface{}{"from": "USD", "to": "EUR", "amount": 10.0},
		})
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("missing from in payload", func(t *testing.T) {
		resp := authedPost(t, "/api/v1/events/publish", auth.AccessToken, map[string]interface{}{
			"type":    "conversion",
			"payload": map[string]interface{}{"to": "EUR", "amount": 10.0},
		})
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("amount is not a number", func(t *testing.T) {
		resp := authedPost(t, "/api/v1/events/publish", auth.AccessToken, map[string]interface{}{
			"type":    "conversion",
			"payload": map[string]interface{}{"from": "USD", "to": "EUR", "amount": "not-a-number"},
		})
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("no authorization", func(t *testing.T) {
		resp := authedPost(t, "/api/v1/events/publish", "", validPayload)
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusUnauthorized)
	})
}
