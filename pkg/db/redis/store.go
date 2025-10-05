package redis

import (
	"bytes"
	"encoding/gob"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	goredis "github.com/redis/go-redis/v9"
)

// Store implements gorilla/sessions Store interface
type Store struct {
	redisClient Client
	keyPrefix   string
	options     *sessions.Options
}

func NewStore(client Client, keyPrefix string, opts *sessions.Options) *Store {
	return &Store{
		redisClient: client,
		keyPrefix:   keyPrefix,
		options:     opts,
	}
}

func (s *Store) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

func (s *Store) New(r *http.Request, name string) (*sessions.Session, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			return nil, err
		}
		// if the cookie does not exist, return a new session
		return s.newSession(name), nil
	}

	key := s.keyPrefix + cookie.Value
	cmd := s.redisClient.Get(r.Context(), key)
	data, err := cmd.Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			// return a new session if the key does not exist
			return s.newSession(name), nil
		}
		return nil, err
	}

	session := sessions.NewSession(s, name)
	err = gob.NewDecoder(bytes.NewReader(data)).Decode(&session.Values)
	if err != nil {
		return nil, err
	}

	session.ID = cookie.Value
	session.IsNew = false
	return session, nil
}

func (s *Store) newSession(name string) *sessions.Session {
	session := sessions.NewSession(s, name)
	session.ID = uuid.NewString()
	session.IsNew = true

	session.Options = &sessions.Options{
		Path:     s.options.Path,
		MaxAge:   s.options.MaxAge,
		HttpOnly: s.options.HttpOnly,
		Secure:   s.options.Secure,
	}

	return session
}

func (s *Store) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	if session.ID == "" {
		session.ID = uuid.NewString()
		session.IsNew = true
	}

	data := &bytes.Buffer{}
	err := gob.NewEncoder(data).Encode(session.Values)
	if err != nil {
		return err
	}

	key := s.keyPrefix + session.ID
	exp := time.Duration(session.Options.MaxAge) * time.Second

	if exp < 0 {
		cmd := s.redisClient.Del(r.Context(), key)
		if err = cmd.Err(); err != nil {
			return err
		}
	} else {
		cmd := s.redisClient.Set(r.Context(), key, data.Bytes(), exp)
		if err = cmd.Err(); err != nil {
			return err
		}
	}

	http.SetCookie(w, &http.Cookie{
		Value:    session.ID,
		Name:     session.Name(),
		Path:     session.Options.Path,
		MaxAge:   session.Options.MaxAge,
		HttpOnly: session.Options.HttpOnly,
	})

	return nil
}
