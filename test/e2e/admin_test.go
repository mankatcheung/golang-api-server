//go:build e2e

package e2e_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/golang-api-server/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminListUsers(t *testing.T) {
	truncateUsers(t)
	_, _, adminToken := seedAdmin(t)

	// Register a regular user so the list is non-empty.
	registerResp := postJSON(t, "/api/v1/auth/register", map[string]string{
		"email": "regular@example.com", "password": "password123", "name": "Regular User",
	})
	require.Equal(t, http.StatusCreated, registerResp.StatusCode)
	registerResp.Body.Close()

	t.Run("admin can list users", func(t *testing.T) {
		resp := authedGet(t, "/api/v1/admin/users", adminToken)
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusOK)

		var users []*model.User
		decodeJSON(t, resp.Body, &users)
		assert.GreaterOrEqual(t, len(users), 2)
	})

	t.Run("regular user gets 403", func(t *testing.T) {
		auth := mustLogin(t, "regular@example.com", "password123")
		resp := authedGet(t, "/api/v1/admin/users", auth.AccessToken)
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusForbidden)
	})

	t.Run("no authorization", func(t *testing.T) {
		resp := authedGet(t, "/api/v1/admin/users", "")
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusUnauthorized)
	})
}

func TestAdminDeleteUser(t *testing.T) {
	truncateUsers(t)
	_, _, adminToken := seedAdmin(t)

	// Register a target user.
	registerResp := postJSON(t, "/api/v1/auth/register", map[string]string{
		"email": "target@example.com", "password": "password123", "name": "Target User",
	})
	require.Equal(t, http.StatusCreated, registerResp.StatusCode)
	var targetAuth model.AuthResponse
	decodeJSON(t, registerResp.Body, &targetAuth)
	registerResp.Body.Close()

	// Get the target user's ID from the profile endpoint.
	profileResp := authedGet(t, "/api/v1/me", targetAuth.AccessToken)
	require.Equal(t, http.StatusOK, profileResp.StatusCode)
	var target model.User
	decodeJSON(t, profileResp.Body, &target)
	profileResp.Body.Close()

	t.Run("regular user gets 403", func(t *testing.T) {
		auth := mustLogin(t, "target@example.com", "password123")
		resp := authedDelete(t, fmt.Sprintf("/api/v1/admin/users/%d", target.ID), auth.AccessToken)
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusForbidden)
	})

	t.Run("non-existent user returns 404", func(t *testing.T) {
		resp := authedDelete(t, "/api/v1/admin/users/99999", adminToken)
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusNotFound)
	})

	t.Run("admin deletes user successfully", func(t *testing.T) {
		resp := authedDelete(t, fmt.Sprintf("/api/v1/admin/users/%d", target.ID), adminToken)
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusOK)
	})

	t.Run("deleted user no longer exists", func(t *testing.T) {
		resp := authedDelete(t, fmt.Sprintf("/api/v1/admin/users/%d", target.ID), adminToken)
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusNotFound)
	})

	t.Run("no authorization", func(t *testing.T) {
		resp := authedDelete(t, fmt.Sprintf("/api/v1/admin/users/%d", target.ID), "")
		defer resp.Body.Close()
		assertStatus(t, resp, http.StatusUnauthorized)
	})
}
