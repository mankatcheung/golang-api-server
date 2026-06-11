package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang-api-server/internal/domain"
	"github.com/golang-api-server/internal/model"
	"github.com/golang-api-server/internal/service"
	"github.com/golang-api-server/pkg/bloom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserService_List(t *testing.T) {
	tests := []struct {
		name    string
		repo    *mockUserRepo
		want    int
		wantErr bool
	}{
		{
			name: "success",
			repo: &mockUserRepo{
				listFunc: func(_ context.Context, offset, limit int) ([]*model.User, error) {
					return []*model.User{
						{ID: 1, Email: "a@test.com", Name: "A"},
						{ID: 2, Email: "b@test.com", Name: "B"},
					}, nil
				},
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "empty list",
			repo: &mockUserRepo{
				listFunc: func(_ context.Context, _, _ int) ([]*model.User, error) {
					return []*model.User{}, nil
				},
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "repo error",
			repo: &mockUserRepo{
				listFunc: func(_ context.Context, _, _ int) ([]*model.User, error) {
					return nil, errors.New("database error")
				},
			},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := service.NewUserService(tt.repo, bloom.New(100, 0.01))
			users, err := svc.List(context.Background(), 0, 100)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, users, tt.want)
		})
	}
}

func TestUserService_Delete(t *testing.T) {
	tests := []struct {
		name    string
		repo    *mockUserRepo
		id      int64
		wantErr error
	}{
		{
			name: "success",
			repo: &mockUserRepo{
				deleteFunc: func(_ context.Context, _ int64) error {
					return nil
				},
			},
			id:      1,
			wantErr: nil,
		},
		{
			name: "user not found",
			repo: &mockUserRepo{
				deleteFunc: func(_ context.Context, _ int64) error {
					return domain.ErrNotFound
				},
			},
			id:      999,
			wantErr: domain.ErrNotFound,
		},
		{
			name: "repo error",
			repo: &mockUserRepo{
				deleteFunc: func(_ context.Context, _ int64) error {
					return errors.New("database error")
				},
			},
			id:      1,
			wantErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := service.NewUserService(tt.repo, bloom.New(100, 0.01))
			err := svc.Delete(context.Background(), tt.id)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestUserService_CheckAvailability(t *testing.T) {
	const (
		takenEmail    = "taken@example.com"
		takenUsername = "takenuser"
	)

	tests := []struct {
		name              string
		repo              *mockUserRepo
		filterSetup       func(*bloom.Filter)
		email             string
		username          string
		wantEmailAvail    bool
		wantUsernameAvail bool
		wantErr           bool
	}{
		{
			name:              "email not in bloom, no DB call",
			repo:              &mockUserRepo{},
			filterSetup:       func(_ *bloom.Filter) {},
			email:             takenEmail,
			wantEmailAvail:    true,
			wantUsernameAvail: true,
		},
		{
			name: "email in bloom, DB confirms taken",
			repo: &mockUserRepo{
				getByEmailFunc: func(_ context.Context, _ string) (*model.User, error) {
					return &model.User{ID: 1, Email: takenEmail}, nil
				},
			},
			filterSetup: func(f *bloom.Filter) {
				f.Add("email:" + takenEmail)
			},
			email:             takenEmail,
			wantEmailAvail:    false,
			wantUsernameAvail: true,
		},
		{
			name: "email in bloom, DB says not found (false positive)",
			repo: &mockUserRepo{
				getByEmailFunc: func(_ context.Context, _ string) (*model.User, error) {
					return nil, domain.ErrNotFound
				},
			},
			filterSetup: func(f *bloom.Filter) {
				f.Add("email:" + takenEmail)
			},
			email:             takenEmail,
			wantEmailAvail:    true,
			wantUsernameAvail: true,
		},
		{
			name:              "username not in bloom, no DB call",
			repo:              &mockUserRepo{},
			filterSetup:       func(_ *bloom.Filter) {},
			username:          takenUsername,
			wantEmailAvail:    true,
			wantUsernameAvail: true,
		},
		{
			name: "username in bloom, DB confirms taken",
			repo: &mockUserRepo{
				getByUsernameFunc: func(_ context.Context, _ string) (*model.User, error) {
					return &model.User{ID: 1, Username: takenUsername}, nil
				},
			},
			filterSetup: func(f *bloom.Filter) {
				f.Add("username:" + takenUsername)
			},
			username:          takenUsername,
			wantEmailAvail:    true,
			wantUsernameAvail: false,
		},
		{
			name: "DB error on email bloom hit",
			repo: &mockUserRepo{
				getByEmailFunc: func(_ context.Context, _ string) (*model.User, error) {
					return nil, errors.New("db error")
				},
			},
			filterSetup: func(f *bloom.Filter) {
				f.Add("email:" + takenEmail)
			},
			email:   takenEmail,
			wantErr: true,
		},
		{
			name:              "empty params, both available",
			repo:              &mockUserRepo{},
			filterSetup:       func(_ *bloom.Filter) {},
			wantEmailAvail:    true,
			wantUsernameAvail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := bloom.New(100, 0.01)
			tt.filterSetup(f)
			svc := service.NewUserService(tt.repo, f)

			emailAvail, usernameAvail, err := svc.CheckAvailability(context.Background(), tt.email, tt.username)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantEmailAvail, emailAvail)
			assert.Equal(t, tt.wantUsernameAvail, usernameAvail)
		})
	}
}
