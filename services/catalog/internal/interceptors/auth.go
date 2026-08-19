package interceptors

import (
	"context"

	"catalog/internal/errs"
	"catalog/internal/types"
	"catalog/internal/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func GetAuthUnaryInterceptor(methodsWithAuth map[string]bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if !methodsWithAuth[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errs.Unauthenticated()
		}

		userID := md["x-user-id"]
		if len(userID) < 1 {
			return nil, errs.Unauthenticated()
		}

		userUUID, appErr := utils.StringToUUID(userID[0])
		if appErr != nil {
			return nil, errs.Unauthenticated()
		}

		ctx = context.WithValue(ctx, types.UserIDKey, &userUUID)
		return handler(ctx, req)
	}
}
