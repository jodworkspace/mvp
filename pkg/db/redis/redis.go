package redis

import (
	"context"
	"io"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Client is a subset of the go-redis Cmdable interface
type Client interface {
	Scan(ctx context.Context, cursor uint64, match string, count int64) *goredis.ScanCmd
	Keys(ctx context.Context, pattern string) *goredis.StringSliceCmd
	Exists(ctx context.Context, keys ...string) *goredis.IntCmd
	Get(ctx context.Context, key string) *goredis.StringCmd
	Set(ctx context.Context, key string, value any, expiration time.Duration) *goredis.StatusCmd
	Del(ctx context.Context, keys ...string) *goredis.IntCmd
	MGet(ctx context.Context, keys ...string) *goredis.SliceCmd
	MSet(ctx context.Context, values ...any) *goredis.StatusCmd
	io.Closer
}

type client struct {
	rdb *goredis.Client
}

func NewClient(config *goredis.Options) (Client, error) {
	rdb := goredis.NewClient(config)

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &client{rdb: rdb}, nil
}

func (r *client) Keys(ctx context.Context, pattern string) *goredis.StringSliceCmd {
	return r.rdb.Keys(ctx, pattern)
}

func (r *client) Scan(ctx context.Context, cursor uint64, match string, count int64) *goredis.ScanCmd {
	return r.rdb.Scan(ctx, cursor, match, count)
}

func (r *client) Exists(ctx context.Context, key ...string) *goredis.IntCmd {
	return r.rdb.Exists(ctx, key...)
}

func (r *client) Get(ctx context.Context, key string) *goredis.StringCmd {
	return r.rdb.Get(ctx, key)
}

func (r *client) Set(ctx context.Context, key string, value any, expiration time.Duration) *goredis.StatusCmd {
	return r.rdb.Set(ctx, key, value, expiration)
}

func (r *client) Del(ctx context.Context, keys ...string) *goredis.IntCmd {
	return r.rdb.Del(ctx, keys...)
}

func (r *client) MGet(ctx context.Context, keys ...string) *goredis.SliceCmd {
	return r.rdb.MGet(ctx, keys...)
}

func (r *client) MSet(ctx context.Context, pairs ...any) *goredis.StatusCmd {
	return r.rdb.MSet(ctx, pairs...)
}

func (r *client) Close() error {
	return r.rdb.Close()
}
