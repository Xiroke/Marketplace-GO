package utils

import (
	"api-gateway/internal/errs"
	catalogv1 "api-gateway/internal/grpc/catalog/v1"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	ParseError = func(err error) *errs.AppError { return errs.Internal("failed to parse input data", err) }
)

func StringToNumeric(str string) (pgtype.Numeric, *errs.AppError) {
	var numeric pgtype.Numeric

	err := numeric.Scan(str)
	if err != nil {
		return pgtype.Numeric{Valid: false}, ParseError(err)
	}

	return numeric, nil
}

func StringToUUID(str string) (pgtype.UUID, *errs.AppError) {
	var uuid pgtype.UUID

	err := uuid.Scan(str)
	if err != nil {
		return pgtype.UUID{Valid: false}, ParseError(err)
	}

	return uuid, nil
}

func NumericToString(numeric pgtype.Numeric) (string, *errs.AppError) {
	str, err := numeric.Value()
	if err != nil {
		return "", ParseError(err)
	}

	return str.(string), nil
}

func TimestamptzToGRPCTimestamp(timestamp pgtype.Timestamptz) *timestamppb.Timestamp {
	return timestamppb.New(timestamp.Time)
}

func ProductStatusFromGRPCToString(status catalogv1.ProductStatus) (string, error) {
	switch status {
	case catalogv1.ProductStatus_PRODUCT_STATUS_DRAFT:
		return "draft", nil
	case catalogv1.ProductStatus_PRODUCT_STATUS_PUBLISHED:
		return "published", nil
	case catalogv1.ProductStatus_PRODUCT_STATUS_ARCHIVED:
		return "archived", nil

	case catalogv1.ProductStatus_PRODUCT_STATUS_UNSPECIFIED:
		return "", fmt.Errorf("product status is required")
	default:
		return "", fmt.Errorf("unknown product status %v", status)
	}
}
