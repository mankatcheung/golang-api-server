package service_test

import (
	"context"

	"github.com/golang-api-server/internal/model"
)

type mockUserRepo struct {
	createFunc     func(ctx context.Context, user *model.User, passwordHash string) error
	getByIDFunc    func(ctx context.Context, id int64) (*model.User, error)
	getByEmailFunc func(ctx context.Context, email string) (*model.User, error)
	updateFunc     func(ctx context.Context, user *model.User) error
	deleteFunc     func(ctx context.Context, id int64) error
	listFunc       func(ctx context.Context, offset, limit int) ([]*model.User, error)
}

func (m *mockUserRepo) Create(ctx context.Context, user *model.User, passwordHash string) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, user, passwordHash)
	}
	return nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	if m.getByEmailFunc != nil {
		return m.getByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *mockUserRepo) Update(ctx context.Context, user *model.User) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, user)
	}
	return nil
}

func (m *mockUserRepo) Delete(ctx context.Context, id int64) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	return nil, nil
}

func (m *mockUserRepo) AllEmailsAndUsernames(_ context.Context) ([]string, []string, error) {
	return nil, nil, nil
}

func (m *mockUserRepo) List(ctx context.Context, offset, limit int) ([]*model.User, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, offset, limit)
	}
	return nil, nil
}

type mockProducer struct {
	publishFunc func(ctx context.Context, key, value interface{}) error
}

func (m *mockProducer) Publish(ctx context.Context, key, value interface{}) error {
	if m.publishFunc != nil {
		return m.publishFunc(ctx, key, value)
	}
	return nil
}

func (m *mockProducer) Close() error {
	return nil
}
