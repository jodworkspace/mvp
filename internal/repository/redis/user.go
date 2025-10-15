package redis

import (
	"gitlab.com/jodworkspace/mvp/pkg/db/redis"
)

type UserRepository struct {
	redisClient redis.Client
}
