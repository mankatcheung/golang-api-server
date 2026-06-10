// Package grpc provides gRPC server implementations for the API services.
package grpc

import (
	"context"
	"errors"
	"time"

	"github.com/golang-api-server/internal/domain"
	"github.com/golang-api-server/internal/model"
	"github.com/golang-api-server/internal/service"
	pb "github.com/golang-api-server/proto/auth"
)

// AuthGRPCHandler implements the auth gRPC service by delegating to AuthService.
type AuthGRPCHandler struct {
	pb.UnimplementedAuthServiceServer
	authService service.AuthService
}

// NewAuthGRPCHandler returns a new AuthGRPCHandler.
func NewAuthGRPCHandler(authService service.AuthService) *AuthGRPCHandler {
	return &AuthGRPCHandler{authService: authService}
}

// Register creates a new user and returns tokens.
func (h *AuthGRPCHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.AuthResponse, error) {
	resp, err := h.authService.Register(ctx, &model.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	})
	if err != nil {
		if errors.Is(err, service.ErrEmailTaken) {
			return nil, errors.New("email already taken")
		}
		return nil, err
	}
	return &pb.AuthResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    resp.ExpiresAt,
	}, nil
}

// Login authenticates a user and returns tokens.
func (h *AuthGRPCHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
	resp, err := h.authService.Login(ctx, &model.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}
	return &pb.AuthResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    resp.ExpiresAt,
	}, nil
}

// RefreshToken issues new tokens from a valid refresh token.
func (h *AuthGRPCHandler) RefreshToken(ctx context.Context, req *pb.RefreshRequest) (*pb.AuthResponse, error) {
	resp, err := h.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidRefreshToken) {
			return nil, errors.New("invalid or expired refresh token")
		}
		return nil, err
	}
	return &pb.AuthResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    resp.ExpiresAt,
	}, nil
}

// GetProfile returns the user profile for the given user ID.
func (h *AuthGRPCHandler) GetProfile(ctx context.Context, req *pb.ProfileRequest) (*pb.User, error) {
	user, err := h.authService.GetProfile(ctx, req.UserId)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &pb.User{
		Id:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Role:      string(user.Role),
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// Logout is a placeholder that returns a confirmation message.
func (h *AuthGRPCHandler) Logout(ctx context.Context, req *pb.MessageResponse) (*pb.MessageResponse, error) {
	return &pb.MessageResponse{Message: "logged out"}, nil
}
