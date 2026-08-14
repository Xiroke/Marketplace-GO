package interceptors

import (
	"catalog/internal/errs"
	"context"
	"errors"
	"log/slog"

	"google.golang.org/grpc"
)

func LoggingInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
        resp, err = handler(ctx, req)
        if err != nil {
            var appError *errs.AppError
             if (errors.As(err, &appError)) {
                logger.Error("application error",
                    "code", appError.Code,
                    "msg", appError.Msg,
                    "cause", appError.Err,
                )
             } else {
                logger.Error("unexpected error",
                    "error", err,
                )
             }
        }
        return resp, err
    }
}
