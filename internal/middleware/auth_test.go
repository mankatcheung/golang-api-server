package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-api-server/internal/middleware"
	"github.com/golang-api-server/pkg/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret-for-middleware"

func setupMiddlewareRouter(mw ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(mw...)
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"user_id": c.GetInt64("user_id"),
			"email":   c.GetString("email"),
			"role":    c.GetString("role"),
		})
	})
	r.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	return r
}

func generateToken(t *testing.T, userID int64, email, role string) string {
	t.Helper()
	pair, err := jwt.GenerateTokenPair(testSecret, userID, email, role, 15*time.Minute, 7*24*time.Hour)
	require.NoError(t, err)
	return pair.AccessToken
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	token := generateToken(t, 42, "user@test.com", "user")
	router := setupMiddlewareRouter(middleware.AuthMiddleware(testSecret))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(42), resp["user_id"])
	assert.Equal(t, "user@test.com", resp["email"])
	assert.Equal(t, "user", resp["role"])
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	router := setupMiddlewareRouter(middleware.AuthMiddleware(testSecret))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "authorization header required", resp["error"])
}

func TestAuthMiddleware_MalformedHeader(t *testing.T) {
	tests := []struct {
		name   string
		header string
	}{
		{"no bearer prefix", "Token abc123"},
		{"only bearer", "Bearer"},
		{"no space", "Bearerabc123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupMiddlewareRouter(middleware.AuthMiddleware(testSecret))

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", tt.header)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	router := setupMiddlewareRouter(middleware.AuthMiddleware(testSecret))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	pair, err := jwt.GenerateTokenPair(testSecret, 1, "user@test.com", "user", -1*time.Hour, 7*24*time.Hour)
	require.NoError(t, err)

	router := setupMiddlewareRouter(middleware.AuthMiddleware(testSecret))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_WrongSecret(t *testing.T) {
	token, err := jwt.GenerateTokenPair("wrong-secret", 1, "user@test.com", "user", 15*time.Minute, 7*24*time.Hour)
	require.NoError(t, err)

	router := setupMiddlewareRouter(middleware.AuthMiddleware(testSecret))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireRole_AllowedRole(t *testing.T) {
	token := generateToken(t, 1, "admin@test.com", "admin")
	router := setupMiddlewareRouter(
		middleware.AuthMiddleware(testSecret),
		middleware.RequireRole("admin"),
	)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRole_ForbiddenRole(t *testing.T) {
	token := generateToken(t, 1, "user@test.com", "user")
	router := setupMiddlewareRouter(
		middleware.AuthMiddleware(testSecret),
		middleware.RequireRole("admin"),
	)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireRole_NoRoleInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequireRole("admin"))
	r.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireRole_MultipleAllowedRoles(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		wantCode int
	}{
		{"admin allowed", "admin", http.StatusOK},
		{"editor allowed", "editor", http.StatusOK},
		{"user forbidden", "user", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := generateToken(t, 1, "test@test.com", tt.role)
			router := setupMiddlewareRouter(
				middleware.AuthMiddleware(testSecret),
				middleware.RequireRole("admin", "editor"),
			)

			req := httptest.NewRequest(http.MethodGet, "/admin", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}
