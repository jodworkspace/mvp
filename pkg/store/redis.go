package store

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/sessions"
	"github.com/redis/go-redis/v9"
	"net/http"
)

type RedisStore struct {
	Client    *redis.Client
	KeyPrefix string
}

func NewRedisStore(client *redis.Client, keyPrefix string) *RedisStore {
	return &RedisStore{
		Client:    client,
		KeyPrefix: keyPrefix,
	}
}

func (s *RedisStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

func (s *RedisStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(s, name)
	session.Options = &sessions.Options{
		Path: "/",
	}
	session.IsNew = true

	cookie, err := r.Cookie(name)
	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			return nil, err
		}
		return session, nil
	}

	data, err := s.Client.Get(r.Context(), s.KeyPrefix+cookie.Value).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return session, nil
		}
		return nil, err
	}

	err = json.Unmarshal([]byte(data), &session.Values)
	if err != nil {
		return session, err
	}

	session.IsNew = false
	return session, nil
}

func (s *RedisStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	return nil
}
