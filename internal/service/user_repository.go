package service

import (
	"context"

	"github.com/golang-api-server/internal/model"
)

// UserRepository defines the persistence operations the auth service needs.
type UserRepository interface {
	Create(ctx context.Context, user *model.User, passwordHash string) error
	GetByID(ctx context.Context, id int64) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, offset, limit int) ([]*model.User, error)
	AllEmailsAndUsernames(ctx context.Context) (emails, usernames []string, err error)
}

// UserService defines user management operations exposed to adapters.
type UserService interface {
	List(ctx context.Context, offset, limit int) ([]*model.User, error)
	Delete(ctx context.Context, id int64) error
	CheckAvailability(ctx context.Context, email, username string) (emailAvail, usernameAvail bool, err error)
}
