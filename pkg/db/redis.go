package db

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisClient interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) ([]byte, error)
	Del(ctx context.Context, key string) error
}

type redisClient struct {
	client *redis.Client
}

func NewRedisClient(options *redis.Options) RedisClient {
	return &redisClient{
		client: redis.NewClient(options),
	}
}

func NewRedisFailoverClient(options *redis.FailoverOptions) RedisClient {
	return &redisClient{
		client: redis.NewFailoverClient(options),
	}
}

func (r *redisClient) Get(ctx context.Context, key string) ([]byte, error) {
	res := r.client.Get(ctx, key)
	if err := res.Err(); err != nil {
		return nil, err
	}

	return res.Bytes()
}

func (r *redisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) ([]byte, error) {
	res := r.client.Set(ctx, key, value, expiration)
	if err := res.Err(); err != nil {
		return nil, err
	}

	// Return "OK" as a byte slice
	return res.Bytes()
}

func (r *redisClient) Del(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}
