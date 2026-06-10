package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-api-server/internal/database"
	"github.com/golang-api-server/internal/domain"
	"github.com/golang-api-server/internal/model"
	"github.com/golang-api-server/internal/service"
	"gorm.io/gorm"
)

// Compile-time check that UserRepositoryImpl implements service.UserRepository.
var _ service.UserRepository = (*UserRepositoryImpl)(nil)

type UserRepositoryImpl struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) service.UserRepository {
	return &UserRepositoryImpl{db: db}
}

func (r *UserRepositoryImpl) Create(ctx context.Context, user *model.User, passwordHash string) error {
	db := database.TXFromContext(ctx, r.db)
	user.Password = passwordHash
	result := db.WithContext(ctx).Create(user)
	if result.Error != nil {
		if isUniqueViolation(result.Error) {
			return domain.ErrConflict
		}
		return fmt.Errorf("create user: %w", result.Error)
	}
	return nil
}

func (r *UserRepositoryImpl) GetByID(ctx context.Context, id int64) (*model.User, error) {
	db := database.TXFromContext(ctx, r.db)
	user := &model.User{}
	result := db.WithContext(ctx).First(user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get user by id %d: %w", id, result.Error)
	}
	return user, nil
}

func (r *UserRepositoryImpl) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	db := database.TXFromContext(ctx, r.db)
	user := &model.User{}
	result := db.WithContext(ctx).Where("email = ?", email).First(user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get user by email %s: %w", email, result.Error)
	}
	return user, nil
}

func (r *UserRepositoryImpl) Update(ctx context.Context, user *model.User) error {
	db := database.TXFromContext(ctx, r.db)
	result := db.WithContext(ctx).Model(user).Updates(map[string]interface{}{
		"name":  user.Name,
		"email": user.Email,
	})
	if result.Error != nil {
		return fmt.Errorf("update user %d: %w", user.ID, result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *UserRepositoryImpl) Delete(ctx context.Context, id int64) error {
	db := database.TXFromContext(ctx, r.db)
	result := db.WithContext(ctx).Delete(&model.User{}, id)
	if result.Error != nil {
		return fmt.Errorf("delete user %d: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *UserRepositoryImpl) List(ctx context.Context, offset, limit int) ([]*model.User, error) {
	db := database.TXFromContext(ctx, r.db)
	var users []*model.User
	result := db.WithContext(ctx).Offset(offset).Limit(limit).Order("id").Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("list users: %w", result.Error)
	}
	return users, nil
}

func isUniqueViolation(err error) bool {
	return fmt.Sprintf("%v", err) == "duplicate key value violates unique constraint"
}
