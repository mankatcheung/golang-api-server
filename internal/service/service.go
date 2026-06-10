// Package service implements the business logic layer of the application.
package service

import (
	"context"

	"github.com/golang-api-server/internal/model"
)

// AuthService defines the interface for authentication operations.
type AuthService interface {
	Register(ctx context.Context, req *model.RegisterRequest) (*model.AuthResponse, error)
	Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*model.AuthResponse, error)
	GetProfile(ctx context.Context, userID int64) (*model.User, error)
	UpdateProfile(ctx context.Context, userID int64, req *model.UpdateProfileRequest) (*model.User, error)
}

// ExchangeService defines the interface for currency exchange operations.
type ExchangeService interface {
	Convert(ctx context.Context, req *model.ConvertRequest) (*model.ConvertResponse, error)
	GetRates(ctx context.Context, base string) (map[string]float64, error)
}

// EventPublisher defines the interface for publishing domain events.
type EventPublisher interface {
	PublishConversion(ctx context.Context, event *model.ConversionEvent) error
}
