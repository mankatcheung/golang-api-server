package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-api-server/internal/domain"
	"github.com/golang-api-server/internal/model"
	"github.com/golang-api-server/internal/service"
	"github.com/golang-api-server/pkg/password"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestAuthService(repo service.UserRepository) service.AuthService {
	return service.NewAuthService(service.AuthServiceDeps{
		UserRepo:      repo,
		JWTSecret:     "test-secret-key-for-jwt-signing",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	})
}

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name    string
		repo    *mockUserRepo
		req     *model.RegisterRequest
		wantErr error
	}{
		{
			name: "success",
			repo: &mockUserRepo{
				getByEmailFunc: func(_ context.Context, _ string) (*model.User, error) {
					return nil, domain.ErrNotFound
				},
				createFunc: func(_ context.Context, user *model.User, _ string) error {
					user.ID = 1
					return nil
				},
			},
			req: &model.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Name:     "Test User",
			},
			wantErr: nil,
		},
		{
			name: "email already taken",
			repo: &mockUserRepo{
				getByEmailFunc: func(_ context.Context, _ string) (*model.User, error) {
					return &model.User{ID: 1, Email: "taken@example.com"}, nil
				},
			},
			req: &model.RegisterRequest{
				Email:    "taken@example.com",
				Password: "password123",
				Name:     "Test User",
			},
			wantErr: service.ErrEmailTaken,
		},
		{
			name: "repo error on get by email",
			repo: &mockUserRepo{
				getByEmailFunc: func(_ context.Context, _ string) (*model.User, error) {
					return nil, errors.New("database connection failed")
				},
			},
			req: &model.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Name:     "Test User",
			},
			wantErr: errors.New("check existing user"),
		},
		{
			name: "repo error on create",
			repo: &mockUserRepo{
				getByEmailFunc: func(_ context.Context, _ string) (*model.User, error) {
					return nil, domain.ErrNotFound
				},
				createFunc: func(_ context.Context, _ *model.User, _ string) error {
					return errors.New("database insert failed")
				},
			},
			req: &model.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Name:     "Test User",
			},
			wantErr: errors.New("create user"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestAuthService(tt.repo)
			resp, err := svc.Register(context.Background(), tt.req)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.NotEmpty(t, resp.AccessToken)
			assert.NotEmpty(t, resp.RefreshToken)
			assert.Greater(t, resp.ExpiresAt, int64(0))
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	validHash, err := password.Hash("correctpassword")
	require.NoError(t, err)

	tests := []struct {
		name    string
		repo    *mockUserRepo
		req     *model.LoginRequest
		wantErr error
	}{
		{
			name: "success",
			repo: &mockUserRepo{
				getByEmailFunc: func(_ context.Context, _ string) (*model.User, error) {
					return &model.User{
						ID:       1,
						Email:    "test@example.com",
						Password: validHash,
						Name:     "Test User",
						Role:     model.RoleUser,
					}, nil
				},
			},
			req: &model.LoginRequest{
				Email:    "test@example.com",
				Password: "correctpassword",
			},
			wantErr: nil,
		},
		{
			name: "user not found",
			repo: &mockUserRepo{
				getByEmailFunc: func(_ context.Context, _ string) (*model.User, error) {
					return nil, domain.ErrNotFound
				},
			},
			req: &model.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			wantErr: service.ErrInvalidCredentials,
		},
		{
			name: "repo error",
			repo: &mockUserRepo{
				getByEmailFunc: func(_ context.Context, _ string) (*model.User, error) {
					return nil, errors.New("database error")
				},
			},
			req: &model.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr: errors.New("get user by email"),
		},
		{
			name: "wrong password",
			repo: &mockUserRepo{
				getByEmailFunc: func(_ context.Context, _ string) (*model.User, error) {
					return &model.User{
						ID:       1,
						Email:    "test@example.com",
						Password: validHash,
						Name:     "Test User",
						Role:     model.RoleUser,
					}, nil
				},
			},
			req: &model.LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			wantErr: service.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestAuthService(tt.repo)
			resp, err := svc.Login(context.Background(), tt.req)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.NotEmpty(t, resp.AccessToken)
			assert.NotEmpty(t, resp.RefreshToken)
		})
	}
}

func TestAuthService_GetProfile(t *testing.T) {
	tests := []struct {
		name    string
		repo    *mockUserRepo
		userID  int64
		wantErr error
	}{
		{
			name: "success",
			repo: &mockUserRepo{
				getByIDFunc: func(_ context.Context, id int64) (*model.User, error) {
					return &model.User{ID: id, Email: "test@example.com", Name: "Test"}, nil
				},
			},
			userID:  1,
			wantErr: nil,
		},
		{
			name: "user not found",
			repo: &mockUserRepo{
				getByIDFunc: func(_ context.Context, _ int64) (*model.User, error) {
					return nil, domain.ErrNotFound
				},
			},
			userID:  999,
			wantErr: domain.ErrNotFound,
		},
		{
			name: "repo error",
			repo: &mockUserRepo{
				getByIDFunc: func(_ context.Context, _ int64) (*model.User, error) {
					return nil, errors.New("database error")
				},
			},
			userID:  1,
			wantErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestAuthService(tt.repo)
			user, err := svc.GetProfile(context.Background(), tt.userID)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Nil(t, user)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, user)
			assert.Equal(t, tt.userID, user.ID)
		})
	}
}
