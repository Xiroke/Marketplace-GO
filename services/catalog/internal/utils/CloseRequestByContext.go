package utils

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func CloseCanceledRequestByContext(ctx context.Context) error {
	if ctx.Err() != nil {
		return status.Error(codes.Canceled, "request canceled")
	}
	return nil
}
