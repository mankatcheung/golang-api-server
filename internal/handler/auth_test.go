package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-api-server/internal/domain"
	"github.com/golang-api-server/internal/handler"
	"github.com/golang-api-server/internal/model"
	"github.com/golang-api-server/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAuthRouter(authSvc service.AuthService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := handler.NewAuthHandler(authSvc)
	r.POST("/api/v1/auth/register", h.Register)
	r.POST("/api/v1/auth/login", h.Login)
	r.POST("/api/v1/auth/refresh", h.RefreshToken)
	r.POST("/api/v1/auth/logout", h.Logout)
	r.GET("/api/v1/me", h.GetProfile)
	return r
}

func setupUserRouter(userSvc service.UserService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := handler.NewUserHandler(userSvc)
	r.GET("/admin/users", h.List)
	r.DELETE("/admin/users/:id", h.Delete)
	return r
}

func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name       string
		mock       *mockAuthService
		body       interface{}
		wantStatus int
	}{
		{
			name: "success",
			mock: &mockAuthService{
				registerFunc: func(_ context.Context, _ *model.RegisterRequest) (*model.AuthResponse, error) {
					return &model.AuthResponse{
						AccessToken:  "access-token",
						RefreshToken: "refresh-token",
						ExpiresAt:    time.Now().Add(15 * time.Minute).Unix(),
					}, nil
				},
			},
			body:       model.RegisterRequest{Email: "test@example.com", Password: "password123", Name: "Test", Username: "testuser"},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid JSON",
			mock:       &mockAuthService{},
			body:       "not json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing required fields",
			mock: &mockAuthService{},
			body: model.RegisterRequest{
				Email: "test@example.com",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "email already taken",
			mock: &mockAuthService{
				registerFunc: func(_ context.Context, _ *model.RegisterRequest) (*model.AuthResponse, error) {
					return nil, service.ErrEmailTaken
				},
			},
			body:       model.RegisterRequest{Email: "taken@example.com", Password: "password123", Name: "Test", Username: "takenuser"},
			wantStatus: http.StatusConflict,
		},
		{
			name: "server error",
			mock: &mockAuthService{
				registerFunc: func(_ context.Context, _ *model.RegisterRequest) (*model.AuthResponse, error) {
					return nil, errors.New("database error")
				},
			},
			body:       model.RegisterRequest{Email: "test@example.com", Password: "password123", Name: "Test", Username: "erruser"},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupAuthRouter(tt.mock)

			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name       string
		mock       *mockAuthService
		body       interface{}
		wantStatus int
	}{
		{
			name: "success",
			mock: &mockAuthService{
				loginFunc: func(_ context.Context, _ *model.LoginRequest) (*model.AuthResponse, error) {
					return &model.AuthResponse{
						AccessToken:  "access-token",
						RefreshToken: "refresh-token",
						ExpiresAt:    time.Now().Add(15 * time.Minute).Unix(),
					}, nil
				},
			},
			body:       model.LoginRequest{Email: "test@example.com", Password: "password123"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid JSON",
			mock:       &mockAuthService{},
			body:       "not json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid credentials",
			mock: &mockAuthService{
				loginFunc: func(_ context.Context, _ *model.LoginRequest) (*model.AuthResponse, error) {
					return nil, service.ErrInvalidCredentials
				},
			},
			body:       model.LoginRequest{Email: "wrong@example.com", Password: "wrong"},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "server error",
			mock: &mockAuthService{
				loginFunc: func(_ context.Context, _ *model.LoginRequest) (*model.AuthResponse, error) {
					return nil, errors.New("database error")
				},
			},
			body:       model.LoginRequest{Email: "test@example.com", Password: "password123"},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupAuthRouter(tt.mock)

			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	tests := []struct {
		name       string
		mock       *mockAuthService
		body       interface{}
		wantStatus int
	}{
		{
			name: "success",
			mock: &mockAuthService{
				refreshTokenFunc: func(_ context.Context, _ string) (*model.AuthResponse, error) {
					return &model.AuthResponse{
						AccessToken:  "new-access",
						RefreshToken: "new-refresh",
						ExpiresAt:    time.Now().Add(15 * time.Minute).Unix(),
					}, nil
				},
			},
			body:       model.RefreshRequest{RefreshToken: "valid-token"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid JSON",
			mock:       &mockAuthService{},
			body:       "not json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid refresh token",
			mock: &mockAuthService{
				refreshTokenFunc: func(_ context.Context, _ string) (*model.AuthResponse, error) {
					return nil, service.ErrInvalidRefreshToken
				},
			},
			body:       model.RefreshRequest{RefreshToken: "bad-token"},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "server error",
			mock: &mockAuthService{
				refreshTokenFunc: func(_ context.Context, _ string) (*model.AuthResponse, error) {
					return nil, errors.New("database error")
				},
			},
			body:       model.RefreshRequest{RefreshToken: "token"},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupAuthRouter(tt.mock)

			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupAuthRouter(&mockAuthService{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "logged out", resp["message"])
}

func TestAuthHandler_GetProfile(t *testing.T) {
	tests := []struct {
		name       string
		mock       *mockAuthService
		userID     interface{}
		wantStatus int
	}{
		{
			name: "success",
			mock: &mockAuthService{
				getProfileFunc: func(_ context.Context, _ int64) (*model.User, error) {
					return &model.User{ID: 1, Email: "test@example.com", Name: "Test"}, nil
				},
			},
			userID:     int64(1),
			wantStatus: http.StatusOK,
		},
		{
			name:       "no user_id in context",
			mock:       &mockAuthService{},
			userID:     nil,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "user not found",
			mock: &mockAuthService{
				getProfileFunc: func(_ context.Context, _ int64) (*model.User, error) {
					return nil, domain.ErrNotFound
				},
			},
			userID:     int64(999),
			wantStatus: http.StatusNotFound,
		},
		{
			name: "server error",
			mock: &mockAuthService{
				getProfileFunc: func(_ context.Context, _ int64) (*model.User, error) {
					return nil, errors.New("database error")
				},
			},
			userID:     int64(1),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			h := handler.NewAuthHandler(tt.mock)
			r.GET("/api/v1/me", func(c *gin.Context) {
				if tt.userID != nil {
					c.Set("user_id", tt.userID)
				}
				h.GetProfile(c)
			})

			req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestUserHandler_List(t *testing.T) {
	tests := []struct {
		name       string
		mock       *mockUserService
		wantStatus int
		wantCount  int
	}{
		{
			name: "success",
			mock: &mockUserService{
				listFunc: func(_ context.Context, _, _ int) ([]*model.User, error) {
					return []*model.User{
						{ID: 1, Email: "a@test.com"},
						{ID: 2, Email: "b@test.com"},
					}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantCount:  2,
		},
		{
			name: "server error",
			mock: &mockUserService{
				listFunc: func(_ context.Context, _, _ int) ([]*model.User, error) {
					return nil, errors.New("database error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupUserRouter(tt.mock)

			req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var users []*model.User
				err := json.Unmarshal(w.Body.Bytes(), &users)
				require.NoError(t, err)
				assert.Len(t, users, tt.wantCount)
			}
		})
	}
}

func TestUserHandler_Delete(t *testing.T) {
	tests := []struct {
		name       string
		mock       *mockUserService
		path       string
		wantStatus int
	}{
		{
			name: "success",
			mock: &mockUserService{
				deleteFunc: func(_ context.Context, _ int64) error {
					return nil
				},
			},
			path:       "/admin/users/1",
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid id",
			mock:       &mockUserService{},
			path:       "/admin/users/abc",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "user not found",
			mock: &mockUserService{
				deleteFunc: func(_ context.Context, _ int64) error {
					return domain.ErrNotFound
				},
			},
			path:       "/admin/users/999",
			wantStatus: http.StatusNotFound,
		},
		{
			name: "server error",
			mock: &mockUserService{
				deleteFunc: func(_ context.Context, _ int64) error {
					return errors.New("database error")
				},
			},
			path:       "/admin/users/1",
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupUserRouter(tt.mock)

			req := httptest.NewRequest(http.MethodDelete, tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestGetUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		val  interface{}
		want int64
	}{
		{"valid int64", int64(42), 42},
		{"not exists", nil, 0},
		{"wrong type", "not-int64", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			if tt.val != nil {
				c.Set("user_id", tt.val)
			}
			assert.Equal(t, tt.want, handler.GetUserID(c))
		})
	}
}

func TestGetUserRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		val  interface{}
		want string
	}{
		{"valid string", "admin", "admin"},
		{"not exists", nil, ""},
		{"wrong type", 123, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			if tt.val != nil {
				c.Set("role", tt.val)
			}
			assert.Equal(t, tt.want, handler.GetUserRole(c))
		})
	}
}
