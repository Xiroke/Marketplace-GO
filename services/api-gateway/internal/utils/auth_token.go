package utils

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func DecodeAccessToken(secret []byte, accessToken string) (*UserClaims, error) {
	claims := &UserClaims{}

	token, err := jwt.ParseWithClaims(accessToken, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
