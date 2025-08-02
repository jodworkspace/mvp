package redis

import (
	"context"
	goredis "github.com/redis/go-redis/v9"
	"time"
)

type Client interface {
	Close() error
	Scan(ctx context.Context, match string) ([]string, error)
	Keys(ctx context.Context, pattern string) ([]string, error)
	Exists(ctx context.Context, key ...string) (int64, error)
	Get(ctx context.Context, key string) (string, error)

	// Set sets the value of a key in Redis with an expiration time in seconds.
	Set(ctx context.Context, key string, value any, expiration int) (string, error)

	Del(ctx context.Context, key string) error
	MDel(ctx context.Context, keys ...string) (int64, error)
	MGet(ctx context.Context, keys ...string) ([]any, error)
	MSet(ctx context.Context, pairs ...any) (string, error)
}

type redisClient struct {
	client *goredis.Client
}

func NewClient(opts *goredis.Options) (Client, error) {
	c := goredis.NewClient(opts)

	if err := c.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &redisClient{client: c}, nil
}

func NewFailoverClient(opts *goredis.FailoverOptions) (Client, error) {
	c := goredis.NewFailoverClient(opts)

	if err := c.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &redisClient{client: c}, nil
}

func (r *redisClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	return r.client.Keys(ctx, pattern).Result()
}

func (r *redisClient) Scan(ctx context.Context, pattern string) ([]string, error) {
	var cursor uint64 = 0
	foundKeys := make([]string, 0)

	for {
		resp := r.client.Scan(ctx, cursor, pattern, 1000)
		if resp.Err() != nil {
			return nil, resp.Err()
		}

		keys, nextCursor, err := resp.Result()
		if err != nil {
			return nil, err
		}

		cursor = nextCursor
		foundKeys = append(foundKeys, keys...)

		if cursor == 0 {
			break
		}
	}

	return foundKeys, nil
}

func (r *redisClient) Exists(ctx context.Context, key ...string) (int64, error) {
	resp := r.client.Exists(ctx, key...)
	if err := resp.Err(); err != nil {
		return 0, err
	}

	return resp.Val(), nil
}

func (r *redisClient) Get(ctx context.Context, key string) (string, error) {
	resp := r.client.Get(ctx, key)
	if err := resp.Err(); err != nil {
		return "", err
	}

	return resp.Val(), nil
}

// Set sets the value of a key in Redis with an expiration time in seconds.
func (r *redisClient) Set(ctx context.Context, key string, value any, expiration int) (string, error) {
	exp := time.Duration(expiration) * time.Second
	resp := r.client.Set(ctx, key, value, exp)
	if err := resp.Err(); err != nil {
		return "", err
	}

	return resp.Val(), nil
}

func (r *redisClient) Del(ctx context.Context, key string) error {
	resp := r.client.Del(ctx, key)
	if err := resp.Err(); err != nil {
		return err
	}

	return nil
}

func (r *redisClient) MGet(ctx context.Context, keys ...string) ([]any, error) {
	resp := r.client.MGet(ctx, keys...)
	if err := resp.Err(); err != nil {
		return nil, err
	}

	return resp.Val(), nil
}

func (r *redisClient) MSet(ctx context.Context, pairs ...any) (string, error) {
	resp := r.client.MSet(ctx, pairs...)
	if err := resp.Err(); err != nil {
		return "", err
	}

	return resp.Val(), nil
}

func (r *redisClient) MDel(ctx context.Context, keys ...string) (int64, error) {
	resp := r.client.Del(ctx, keys...)
	if err := resp.Err(); err != nil {
		return 0, err
	}

	return resp.Val(), nil
}

func (r *redisClient) Close() error {
	return r.client.Close()
}
