package utils

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type Claims struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
}

func GenerateJWT(expiry time.Duration, secret string) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry * time.Second)),
		},
		Email: "",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func VerifyJWT(token string, secret string) (*Claims, error) {
	tokenInfo, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := tokenInfo.Claims.(*Claims); ok && tokenInfo.Valid {
		return claims, nil
	}
	return nil, err
}
