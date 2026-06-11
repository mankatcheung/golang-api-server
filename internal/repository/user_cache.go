package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-api-server/internal/cache"
	"github.com/golang-api-server/internal/model"
	"github.com/golang-api-server/internal/service"
)

// CachingUserRepository wraps a UserRepository with Redis caching.
type CachingUserRepository struct {
	underlying service.UserRepository
	cache      cache.Cache
	ttl        time.Duration
}

// NewCachingUserRepository returns a UserRepository with cache-aside reads and write-through invalidation.
func NewCachingUserRepository(underlying service.UserRepository, c cache.Cache, ttl time.Duration) service.UserRepository {
	return &CachingUserRepository{
		underlying: underlying,
		cache:      c,
		ttl:        ttl,
	}
}

func (r *CachingUserRepository) Create(ctx context.Context, user *model.User, passwordHash string) error {
	if err := r.underlying.Create(ctx, user, passwordHash); err != nil {
		return err
	}
	// Cache after zeroing the password — the hash must not be stored in Redis,
	// and the json:"-" tag would silently drop it anyway (causing auth failures
	// on the first login after register if we cached here).
	r.cacheUser(ctx, user)
	return nil
}

func (r *CachingUserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	key := userIDCacheKey(id)
	var cached model.User
	if err := r.cache.Get(ctx, key, &cached); err == nil {
		return &cached, nil
	}

	user, err := r.underlying.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	r.cacheUser(ctx, user)
	return user, nil
}

func (r *CachingUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	key := userEmailCacheKey(email)
	var cached model.User
	if err := r.cache.Get(ctx, key, &cached); err == nil {
		return &cached, nil
	}

	user, err := r.underlying.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	r.cacheUser(ctx, user)
	return user, nil
}

func (r *CachingUserRepository) Update(ctx context.Context, user *model.User) error {
	if err := r.underlying.Update(ctx, user); err != nil {
		return err
	}
	r.cacheUser(ctx, user)
	return nil
}

// cacheUser writes the user to all cache keys with the password field zeroed.
// The password hash must never be stored in Redis: the model's json:"-" tag would
// silently drop it on serialisation, causing auth failures on the next cache hit.
func (r *CachingUserRepository) cacheUser(ctx context.Context, user *model.User) {
	safe := *user
	safe.Password = ""
	cacheSet(ctx, r.cache, userIDCacheKey(safe.ID), &safe, r.ttl)
	cacheSet(ctx, r.cache, userEmailCacheKey(safe.Email), &safe, r.ttl)
	if safe.Username != "" {
		cacheSet(ctx, r.cache, userUsernameCacheKey(safe.Username), &safe, r.ttl)
	}
}

func (r *CachingUserRepository) Delete(ctx context.Context, id int64) error {
	// Fetch email before deletion so we can invalidate the email cache key.
	user, err := r.underlying.GetByID(ctx, id)
	if err == nil {
		cacheDel(ctx, r.cache, userEmailCacheKey(user.Email))
	}
	if err := r.underlying.Delete(ctx, id); err != nil {
		return err
	}
	cacheDel(ctx, r.cache, userIDCacheKey(id))
	return nil
}

func (r *CachingUserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	key := userUsernameCacheKey(username)
	var cached model.User
	if err := r.cache.Get(ctx, key, &cached); err == nil {
		return &cached, nil
	}

	user, err := r.underlying.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	r.cacheUser(ctx, user)
	return user, nil
}

func (r *CachingUserRepository) AllEmailsAndUsernames(ctx context.Context) ([]string, []string, error) {
	return r.underlying.AllEmailsAndUsernames(ctx)
}

func (r *CachingUserRepository) List(ctx context.Context, offset, limit int) ([]*model.User, error) {
	return r.underlying.List(ctx, offset, limit)
}

func userIDCacheKey(id int64) string {
	return fmt.Sprintf("user:id:%d", id)
}

func userEmailCacheKey(email string) string {
	return fmt.Sprintf("user:email:%s", email)
}

func userUsernameCacheKey(username string) string {
	return fmt.Sprintf("user:username:%s", username)
}

func cacheSet(ctx context.Context, c cache.Cache, key string, value interface{}, ttl time.Duration) {
	if err := c.Set(ctx, key, value, ttl); err != nil {
		slog.Debug("cache set failed", "key", key, "error", err)
	}
}

func cacheDel(ctx context.Context, c cache.Cache, keys ...string) {
	if err := c.Delete(ctx, keys...); err != nil {
		slog.Debug("cache delete failed", "keys", keys, "error", err)
	}
}
