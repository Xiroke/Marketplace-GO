package interceptors

import (
	"context"
	"errors"
	"log/slog"

	"catalog/internal/errs"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func GetLoggingInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		resp, err = handler(ctx, req)
		if err != nil {
			var appError *errs.AppError
			if errors.As(err, &appError) {
				if appError.Code == codes.Internal {
					logger.Error("internal error",
						"code", appError.Code,
						"msg", appError.Msg,
						"cause", appError.Err,
					)
				}
			} else {
				logger.Error("unexpected error",
					"error", err,
				)
			}
		}
		return resp, err
	}
}
