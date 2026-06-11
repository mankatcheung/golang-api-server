package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-api-server/internal/domain"
	"github.com/golang-api-server/internal/model"
	"github.com/golang-api-server/pkg/bloom"
)

// Compile-time check that userServiceImpl implements UserService.
var _ UserService = (*userServiceImpl)(nil)

type userServiceImpl struct {
	userRepo UserRepository
	filter   *bloom.Filter
}

// NewUserService returns a UserService backed by the given repository and bloom filter.
func NewUserService(userRepo UserRepository, filter *bloom.Filter) UserService {
	return &userServiceImpl{userRepo: userRepo, filter: filter}
}

func (s *userServiceImpl) List(ctx context.Context, offset, limit int) ([]*model.User, error) {
	users, err := s.userRepo.List(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

func (s *userServiceImpl) Delete(ctx context.Context, id int64) error {
	err := s.userRepo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("delete user %d: %w", id, err)
	}
	return nil
}

func (s *userServiceImpl) CheckAvailability(ctx context.Context, email, username string) (bool, bool, error) {
	emailAvail := true
	if email != "" {
		if s.filter.Contains("email:" + email) {
			_, err := s.userRepo.GetByEmail(ctx, email)
			if err != nil && !errors.Is(err, domain.ErrNotFound) {
				return false, false, fmt.Errorf("check email availability: %w", err)
			}
			emailAvail = errors.Is(err, domain.ErrNotFound)
		}
	}

	usernameAvail := true
	if username != "" {
		if s.filter.Contains("username:" + username) {
			_, err := s.userRepo.GetByUsername(ctx, username)
			if err != nil && !errors.Is(err, domain.ErrNotFound) {
				return false, false, fmt.Errorf("check username availability: %w", err)
			}
			usernameAvail = errors.Is(err, domain.ErrNotFound)
		}
	}

	return emailAvail, usernameAvail, nil
}
