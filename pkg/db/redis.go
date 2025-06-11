package db

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

// RedisClient is a subset of the redis.StringCmdable interface,
type RedisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
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

func (r *redisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	return r.client.Get(ctx, key)
}

func (r *redisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return r.client.Set(ctx, key, value, expiration)
}
