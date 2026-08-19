package utils

import (
	"api-gateway/internal/types"
	"context"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func AddAuthorizationToCTX(ctx context.Context, logger *slog.Logger) (context.Context, error) {
	userID, ok := ctx.Value(types.UserIDKey).(string)
	if !ok {
		logger.Error("failed to parse user UUID", "userID", userID)
		return nil, status.Error(codes.Unauthenticated, "Unauthenticated")
	}

	md := metadata.New(map[string]string{
		"x-user-id": userID,
	})

	ctx = metadata.NewOutgoingContext(ctx, md)
	return ctx, nil
}
