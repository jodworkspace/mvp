package redis

import (
	"gitlab.com/gookie/mvp/pkg/db"
)

type UserRepository struct {
	db.RedisClient
}
