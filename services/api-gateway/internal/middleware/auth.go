package middleware

import (
	"api-gateway/internal/types"
	"api-gateway/internal/utils"
	"api-gateway/internal/utils/responses"
	"context"
	"net/http"
	"strings"
)

func GetAuthMiddleware(JWTSecret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			authorization := r.Header.Get("Authorization")

			if len(authorization) < 1 {
				responses.ResponseWithError(w, "authorization header is missing", http.StatusUnauthorized)
				return
			}

			if !strings.HasPrefix(authorization, "Bearer ") {
				responses.ResponseWithError(w, "invalid authorization header format", http.StatusUnauthorized)
				return
			}

			accessToken := strings.TrimPrefix(authorization, "Bearer ")

			user, err := utils.DecodeAccessToken([]byte(JWTSecret), accessToken)
			if err != nil {
				responses.ResponseWithError(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx = context.WithValue(ctx, types.UserIDKey, user.UserID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

}
