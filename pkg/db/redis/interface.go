package redis

import (
	"context"
	"io"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Client is a subset of the go-redis Cmdable interface
type Client interface {
	SubsetCmdable
	Instrument
}

type SubsetCmdable interface {
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

type Instrument interface {
	Instrument() error
}
