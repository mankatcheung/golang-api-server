package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-api-server/internal/domain"
	"github.com/golang-api-server/internal/model"
)

// Compile-time check that userServiceImpl implements UserService.
var _ UserService = (*userServiceImpl)(nil)

type userServiceImpl struct {
	userRepo UserRepository
}

// NewUserService returns a UserService backed by the given repository.
func NewUserService(userRepo UserRepository) UserService {
	return &userServiceImpl{userRepo: userRepo}
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
