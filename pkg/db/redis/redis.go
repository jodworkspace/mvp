package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	goredis "github.com/redis/go-redis/v9"
)

type client struct {
	rdb *goredis.Client
}

func NewClient(ctx context.Context, addr string, opts ...Option) (Client, error) {
	options := &goredis.Options{
		Addr: addr,
	}

	for _, opt := range opts {
		opt(options)
	}

	rdb := goredis.NewClient(options)

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &client{rdb: rdb}, nil
}

func (c *client) Instrument() error {
	var errs error

	if err := redisotel.InstrumentMetrics(c.rdb); err != nil {
		errs = errors.Join(errs, err)
	}

	if err := redisotel.InstrumentTracing(c.rdb); err != nil {
		errs = errors.Join(errs, err)
	}

	return errs
}

func (c *client) Keys(ctx context.Context, pattern string) *goredis.StringSliceCmd {
	return c.rdb.Keys(ctx, pattern)
}

func (c *client) Scan(ctx context.Context, cursor uint64, match string, count int64) *goredis.ScanCmd {
	return c.rdb.Scan(ctx, cursor, match, count)
}

func (c *client) Exists(ctx context.Context, key ...string) *goredis.IntCmd {
	return c.rdb.Exists(ctx, key...)
}

func (c *client) Get(ctx context.Context, key string) *goredis.StringCmd {
	return c.rdb.Get(ctx, key)
}

func (c *client) Set(ctx context.Context, key string, value any, expiration time.Duration) *goredis.StatusCmd {
	return c.rdb.Set(ctx, key, value, expiration)
}

func (c *client) Del(ctx context.Context, keys ...string) *goredis.IntCmd {
	return c.rdb.Del(ctx, keys...)
}

func (c *client) MGet(ctx context.Context, keys ...string) *goredis.SliceCmd {
	return c.rdb.MGet(ctx, keys...)
}

func (c *client) MSet(ctx context.Context, pairs ...any) *goredis.StatusCmd {
	return c.rdb.MSet(ctx, pairs...)
}

func (c *client) Close() error {
	return c.rdb.Close()
}
