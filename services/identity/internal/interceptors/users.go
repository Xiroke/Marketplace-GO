package interceptors

import (
	"context"
	"strings"

	"identity/internal/config"
	"identity/internal/db"
	"identity/internal/token"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ctxKey string

const UserKey ctxKey = "user"

func GetAuthUnaryInterceptor(methodsWithAuth map[string]bool, config *config.Config, queries *db.Queries) func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if !methodsWithAuth[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "missing metadata")
		}

		authorization := md["authorization"]
		if len(authorization) < 1 {
			return nil, status.Errorf(codes.Unauthenticated, "invalid authorization token")
		}

		accessToken := strings.TrimPrefix(authorization[0], "Bearer ")

		user, err := token.DecodeAccessToken(config.JWTSecret, accessToken)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		newCtx := context.WithValue(ctx, UserKey, user)
		return handler(newCtx, req)
	}
}
