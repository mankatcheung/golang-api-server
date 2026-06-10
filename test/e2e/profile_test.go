//go:build e2e

package e2e_test

import (
	"net/http"
	"testing"

	"github.com/golang-api-server/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestGetProfile(t *testing.T) {
	truncateUsers(t)
	registerResp := postJSON(t, "/api/v1/auth/register", map[string]string{
		"email": "profile@example.com", "password": "password123", "name": "Profile User",
	})
	var auth model.AuthResponse
	decodeJSON(t, registerResp.Body, &auth)
	registerResp.Body.Close()

	t.Run("success", func(t *testing.T) {
		resp := authedGet(t, "/api/v1/me", auth.AccessToken)
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusOK)

		var user model.User
		decodeJSON(t, resp.Body, &user)
		assert.Equal(t, "profile@example.com", user.Email)
		assert.Equal(t, "Profile User", user.Name)
		assert.Equal(t, model.RoleUser, user.Role)
		assert.NotZero(t, user.ID)
	})

	t.Run("no authorization", func(t *testing.T) {
		resp := authedGet(t, "/api/v1/me", "")
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("invalid token", func(t *testing.T) {
		resp := authedGet(t, "/api/v1/me", "bad.token.here")
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusUnauthorized)
	})
}

func TestUpdateProfile(t *testing.T) {
	truncateUsers(t)
	registerResp := postJSON(t, "/api/v1/auth/register", map[string]string{
		"email": "update@example.com", "password": "password123", "name": "Old Name",
	})
	var auth model.AuthResponse
	decodeJSON(t, registerResp.Body, &auth)
	registerResp.Body.Close()

	t.Run("update name", func(t *testing.T) {
		resp := authedPut(t, "/api/v1/me", auth.AccessToken, map[string]string{
			"name": "New Name",
		})
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusOK)

		var user model.User
		decodeJSON(t, resp.Body, &user)
		assert.Equal(t, "New Name", user.Name)
	})

	t.Run("invalid email format", func(t *testing.T) {
		resp := authedPut(t, "/api/v1/me", auth.AccessToken, map[string]string{
			"email": "not-valid",
		})
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("no authorization", func(t *testing.T) {
		resp := authedPut(t, "/api/v1/me", "", map[string]string{"name": "X"})
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusUnauthorized)
	})
}
