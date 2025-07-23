package redis

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/sessions"
	goredis "github.com/redis/go-redis/v9"
	"net/http"
)

type Store struct {
	Client    *goredis.Client
	KeyPrefix string
}

func NewRedisStore(client *goredis.Client, keyPrefix string) *Store {
	return &Store{
		Client:    client,
		KeyPrefix: keyPrefix,
	}
}

func (s *Store) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

func (s *Store) New(r *http.Request, name string) (*sessions.Session, error) {
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
		if errors.Is(err, goredis.Nil) {
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

func (s *Store) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	return nil
}
