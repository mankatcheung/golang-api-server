package handler_test

import (
	"context"

	"github.com/golang-api-server/internal/model"
)

type mockAuthService struct {
	registerFunc      func(ctx context.Context, req *model.RegisterRequest) (*model.AuthResponse, error)
	loginFunc         func(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error)
	refreshTokenFunc  func(ctx context.Context, refreshToken string) (*model.AuthResponse, error)
	getProfileFunc    func(ctx context.Context, userID int64) (*model.User, error)
	updateProfileFunc func(ctx context.Context, userID int64, req *model.UpdateProfileRequest) (*model.User, error)
}

func (m *mockAuthService) Register(ctx context.Context, req *model.RegisterRequest) (*model.AuthResponse, error) {
	if m.registerFunc != nil {
		return m.registerFunc(ctx, req)
	}
	return &model.AuthResponse{}, nil
}

func (m *mockAuthService) Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, req)
	}
	return &model.AuthResponse{}, nil
}

func (m *mockAuthService) RefreshToken(ctx context.Context, refreshToken string) (*model.AuthResponse, error) {
	if m.refreshTokenFunc != nil {
		return m.refreshTokenFunc(ctx, refreshToken)
	}
	return &model.AuthResponse{}, nil
}

func (m *mockAuthService) GetProfile(ctx context.Context, userID int64) (*model.User, error) {
	if m.getProfileFunc != nil {
		return m.getProfileFunc(ctx, userID)
	}
	return &model.User{}, nil
}

func (m *mockAuthService) UpdateProfile(ctx context.Context, userID int64, req *model.UpdateProfileRequest) (*model.User, error) {
	if m.updateProfileFunc != nil {
		return m.updateProfileFunc(ctx, userID, req)
	}
	return &model.User{}, nil
}

type mockUserService struct {
	listFunc                func(ctx context.Context, offset, limit int) ([]*model.User, error)
	deleteFunc              func(ctx context.Context, id int64) error
	checkAvailabilityFunc   func(ctx context.Context, email, username string) (bool, bool, error)
}

func (m *mockUserService) List(ctx context.Context, offset, limit int) ([]*model.User, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, offset, limit)
	}
	return []*model.User{}, nil
}

func (m *mockUserService) Delete(ctx context.Context, id int64) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockUserService) CheckAvailability(ctx context.Context, email, username string) (bool, bool, error) {
	if m.checkAvailabilityFunc != nil {
		return m.checkAvailabilityFunc(ctx, email, username)
	}
	return true, true, nil
}

type mockExchangeService struct {
	convertFunc  func(ctx context.Context, req *model.ConvertRequest) (*model.ConvertResponse, error)
	getRatesFunc func(ctx context.Context, base string) (map[string]float64, error)
}

func (m *mockExchangeService) Convert(ctx context.Context, req *model.ConvertRequest) (*model.ConvertResponse, error) {
	if m.convertFunc != nil {
		return m.convertFunc(ctx, req)
	}
	return &model.ConvertResponse{}, nil
}

func (m *mockExchangeService) GetRates(ctx context.Context, base string) (map[string]float64, error) {
	if m.getRatesFunc != nil {
		return m.getRatesFunc(ctx, base)
	}
	return map[string]float64{}, nil
}

type mockEventPublisher struct {
	publishConversionFunc func(ctx context.Context, event *model.ConversionEvent) error
}

func (m *mockEventPublisher) PublishConversion(ctx context.Context, event *model.ConversionEvent) error {
	if m.publishConversionFunc != nil {
		return m.publishConversionFunc(ctx, event)
	}
	return nil
}
