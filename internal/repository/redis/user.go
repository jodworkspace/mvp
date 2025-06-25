package redis

import (
	"gitlab.com/jodworkspace/mvp/pkg/db"
)

type UserRepository struct {
	db.RedisClient
}
