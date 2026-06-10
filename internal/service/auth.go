package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-api-server/internal/domain"
	"github.com/golang-api-server/internal/model"
	"github.com/golang-api-server/pkg/jwt"
	"github.com/golang-api-server/pkg/password"
)

var (
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrEmailTaken          = errors.New("email already taken")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
	ErrNotFound            = domain.ErrNotFound
)

// Compile-time check that authService implements AuthService.
var _ AuthService = (*authService)(nil)

type authService struct {
	userRepo      UserRepository
	jwtSecret     string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// AuthServiceDeps holds the dependencies required to create an AuthService.
type AuthServiceDeps struct {
	UserRepo      UserRepository
	JWTSecret     string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

// NewAuthService returns an AuthService backed by the given dependencies.
func NewAuthService(deps AuthServiceDeps) AuthService {
	return &authService{
		userRepo:      deps.UserRepo,
		jwtSecret:     deps.JWTSecret,
		accessExpiry:  deps.AccessExpiry,
		refreshExpiry: deps.RefreshExpiry,
	}
}

func (s *authService) Register(ctx context.Context, req *model.RegisterRequest) (*model.AuthResponse, error) {
	existing, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("check existing user: %w", err)
	}
	if existing != nil {
		return nil, ErrEmailTaken
	}

	hash, err := password.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &model.User{
		Email: req.Email,
		Name:  req.Name,
		Role:  model.RoleUser,
	}

	if err := s.userRepo.Create(ctx, user, hash); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return s.generateTokenPair(user)
}

func (s *authService) Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	if !password.Compare(user.Password, req.Password) {
		return nil, ErrInvalidCredentials
	}

	return s.generateTokenPair(user)
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*model.AuthResponse, error) {
	claims, err := jwt.ValidateToken(s.jwtSecret, refreshToken)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	return s.generateTokenPair(user)
}

func (s *authService) GetProfile(ctx context.Context, userID int64) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

func (s *authService) UpdateProfile(ctx context.Context, userID int64, req *model.UpdateProfileRequest) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return user, nil
}

func (s *authService) generateTokenPair(user *model.User) (*model.AuthResponse, error) {
	pair, err := jwt.GenerateTokenPair(
		s.jwtSecret,
		user.ID,
		user.Email,
		string(user.Role),
		s.accessExpiry,
		s.refreshExpiry,
	)
	if err != nil {
		return nil, fmt.Errorf("generate tokens: %w", err)
	}

	return &model.AuthResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresAt:    pair.ExpiresAt.Unix(),
	}, nil
}
