//go:build e2e

package e2e_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/golang-api-server/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	truncateUsers(t)

	t.Run("success", func(t *testing.T) {
		resp := postJSON(t, "/api/v1/auth/register", map[string]string{
			"email": "user@example.com", "password": "password123", "name": "Test User",
		})
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusCreated)

		var auth model.AuthResponse
		decodeJSON(t, resp.Body, &auth)
		assert.NotEmpty(t, auth.AccessToken)
		assert.NotEmpty(t, auth.RefreshToken)
		assert.NotZero(t, auth.ExpiresAt)
	})

	t.Run("duplicate email", func(t *testing.T) {
		resp := postJSON(t, "/api/v1/auth/register", map[string]string{
			"email": "user@example.com", "password": "password123", "name": "Test User",
		})
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusConflict)
	})

	t.Run("missing name", func(t *testing.T) {
		resp := postJSON(t, "/api/v1/auth/register", map[string]string{
			"email": "other@example.com", "password": "password123",
		})
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("invalid email format", func(t *testing.T) {
		resp := postJSON(t, "/api/v1/auth/register", map[string]string{
			"email": "not-an-email", "password": "password123", "name": "Test",
		})
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("password too short", func(t *testing.T) {
		resp := postJSON(t, "/api/v1/auth/register", map[string]string{
			"email": "short@example.com", "password": "abc", "name": "Test",
		})
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("empty body", func(t *testing.T) {
		resp := postJSON(t, "/api/v1/auth/register", map[string]string{})
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusBadRequest)
	})
}

func TestLogin(t *testing.T) {
	truncateUsers(t)
	// Seed a user via register.
	registerResp := postJSON(t, "/api/v1/auth/register", map[string]string{
		"email": "login@example.com", "password": "password123", "name": "Login User",
	})
	require.Equal(t, http.StatusCreated, registerResp.StatusCode)
	io.Copy(io.Discard, registerResp.Body) //nolint:errcheck
	registerResp.Body.Close()

	t.Run("success", func(t *testing.T) {
		resp := postJSON(t, "/api/v1/auth/login", map[string]string{
			"email": "login@example.com", "password": "password123",
		})
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusOK)

		var auth model.AuthResponse
		decodeJSON(t, resp.Body, &auth)
		assert.NotEmpty(t, auth.AccessToken)
		assert.NotEmpty(t, auth.RefreshToken)
	})

	t.Run("wrong password", func(t *testing.T) {
		resp := postJSON(t, "/api/v1/auth/login", map[string]string{
			"email": "login@example.com", "password": "wrongpass",
		})
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("unknown email", func(t *testing.T) {
		resp := postJSON(t, "/api/v1/auth/login", map[string]string{
			"email": "nobody@example.com", "password": "password123",
		})
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("missing fields", func(t *testing.T) {
		resp := postJSON(t, "/api/v1/auth/login", map[string]string{"email": "login@example.com"})
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusBadRequest)
	})
}

func TestRefreshToken(t *testing.T) {
	truncateUsers(t)
	registerResp := postJSON(t, "/api/v1/auth/register", map[string]string{
		"email": "refresh@example.com", "password": "password123", "name": "Refresh User",
	})
	require.Equal(t, http.StatusCreated, registerResp.StatusCode)
	var initial model.AuthResponse
	decodeJSON(t, registerResp.Body, &initial)
	registerResp.Body.Close()

	t.Run("success", func(t *testing.T) {
		resp := postJSON(t, "/api/v1/auth/refresh", map[string]string{
			"refresh_token": initial.RefreshToken,
		})
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusOK)

		var auth model.AuthResponse
		decodeJSON(t, resp.Body, &auth)
		assert.NotEmpty(t, auth.AccessToken)
	})

	t.Run("invalid token", func(t *testing.T) {
		resp := postJSON(t, "/api/v1/auth/refresh", map[string]string{
			"refresh_token": "this.is.not.a.valid.jwt",
		})
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusUnauthorized)
	})
}

func TestLogout(t *testing.T) {
	truncateUsers(t)
	registerResp := postJSON(t, "/api/v1/auth/register", map[string]string{
		"email": "logout@example.com", "password": "password123", "name": "Logout User",
	})
	require.Equal(t, http.StatusCreated, registerResp.StatusCode)
	var auth model.AuthResponse
	decodeJSON(t, registerResp.Body, &auth)
	registerResp.Body.Close()

	t.Run("success with valid token", func(t *testing.T) {
		resp := authedPost(t, "/api/v1/auth/logout", auth.AccessToken, nil)
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusOK)
	})

	t.Run("no authorization header", func(t *testing.T) {
		resp := postJSON(t, "/api/v1/auth/logout", nil)
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusUnauthorized)
	})
}
