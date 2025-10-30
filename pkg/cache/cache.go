package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"solecode/pkg/config"

	"github.com/redis/go-redis/v9"
)

//go:generate mockery --name CacheItf --output mocks --filename cache_mock.go --outpkg mocks
type CacheItf interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}, expiration time.Duration) error
	Delete(key string) error
	GetJSON(key string, v any) error
	SetJSON(key string, v any, expiration time.Duration) error
}

type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

var (
	ErrCacheUnavailable = errors.New("cache service unavailable")
	ErrCacheConnection  = errors.New("cache connection failed")
	ErrCacheOperation   = errors.New("cache operation failed")
	ErrInvalidJSON      = errors.New("invalid JSON data")
	ErrKeyNotFound      = errors.New("key not found")
)

func NewRedisCache(cfg *config.RedisConfig) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx := context.Background()

	// Test connection
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCacheConnection, err)
	}

	return &RedisCache{
		client: client,
		ctx:    ctx,
	}, nil
}

func (r *RedisCache) Get(key string) (interface{}, error) {
	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil // Key doesn't exist, not an error
		}
		return nil, fmt.Errorf("%w: %v", ErrCacheOperation, err)
	}
	return val, nil
}

func (r *RedisCache) Set(key string, value interface{}, expiration time.Duration) error {
	err := r.client.Set(r.ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCacheOperation, err)
	}
	return nil
}

func (r *RedisCache) Delete(key string) error {
	err := r.client.Del(r.ctx, key).Err()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCacheOperation, err)
	}
	return nil
}

func (r *RedisCache) GetJSON(key string, v interface{}) error {
	val, err := r.Get(key)
	if err != nil {
		return err
	}
	if val == nil {
		return nil
	}

	// Convert to string if it's a string
	var jsonStr string
	switch val := val.(type) {
	case string:
		jsonStr = val
	case []byte:
		jsonStr = string(val)
	default:
		return fmt.Errorf("%w: unexpected type %T", ErrInvalidJSON, val)
	}

	if err := json.Unmarshal([]byte(jsonStr), v); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}
	return nil
}

func (r *RedisCache) SetJSON(key string, v interface{}, expiration time.Duration) error {
	val, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	return r.Set(key, string(val), expiration)
}

func (r *RedisCache) Close() error {
	if err := r.client.Close(); err != nil {
		return fmt.Errorf("%w: %v", ErrCacheOperation, err)
	}
	return nil
}

// IsCacheUnavailable checks if the error is due to cache being unavailable
func IsCacheUnavailable(err error) bool {
	return errors.Is(err, ErrCacheUnavailable) ||
		errors.Is(err, ErrCacheConnection) ||
		errors.Is(err, redis.ErrClosed)
}

// IsKeyNotFound checks if the error is due to key not found
func IsKeyNotFound(err error) bool {
	return errors.Is(err, ErrKeyNotFound) ||
		errors.Is(err, redis.Nil)
}
