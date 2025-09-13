package redis

import (
	"context"

	"gitlab.com/jodworkspace/mvp/pkg/db/redis"
)

type BaseRepository struct {
	redisClient redis.Client
}

func (r *BaseRepository) Scan(ctx context.Context, pattern string) ([]string, error) {
	var cursor uint64 = 0
	foundKeys := make([]string, 0)

	for {
		resp := r.redisClient.Scan(ctx, cursor, pattern, 1000)
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
