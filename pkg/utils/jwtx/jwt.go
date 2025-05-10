package jwtx

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"log"
	"time"
)

type Claims struct {
	jwt.RegisteredClaims
}

type Option func(claims *Claims)

func WithIssuer(issuer string) Option {
	return func(claims *Claims) {
		claims.Issuer = issuer
	}
}

func WithSubject(subject string) Option {
	return func(claims *Claims) {
		claims.Subject = subject
	}
}

func WithAudience(audience string) Option {
	return func(claims *Claims) {
		claims.Audience = []string{audience}
	}
}

func GenerateToken(expiry time.Duration, secret []byte, opts ...Option) string {
	now := time.Now()
	claims := Claims{
		jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
		},
	}

	for _, opt := range opts {
		opt(&claims)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(secret)
	if err != nil {
		log.Fatal(err)
	}

	return signedToken
}

func ParseToken(tokenString string, secret []byte, issuer ...string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		return secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if ok && token.Valid {
		return claims, nil
	}

	if len(issuer) > 0 && claims.Issuer != issuer[0] {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return nil, jwt.ErrTokenInvalidClaims
}
