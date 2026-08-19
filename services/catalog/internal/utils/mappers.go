package utils

import (
	"fmt"

	"catalog/internal/db"
	"catalog/internal/errs"
	catalogv1 "catalog/internal/grpc/catalog/v1"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	ParseError = func(err error) *errs.AppError {
		return errs.Internal(fmt.Errorf("failed to parse input data: %w", err))
	}
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

func ProductStatusFromDBToGRPC(status db.ProductStatus) catalogv1.ProductStatus {
	switch status {
	case db.ProductStatusDraft:
		return catalogv1.ProductStatus_PRODUCT_STATUS_DRAFT
	case db.ProductStatusPublished:
		return catalogv1.ProductStatus_PRODUCT_STATUS_PUBLISHED
	case db.ProductStatusArchived:
		return catalogv1.ProductStatus_PRODUCT_STATUS_ARCHIVED
	default:
		return catalogv1.ProductStatus_PRODUCT_STATUS_UNSPECIFIED
	}
}

func ProductStatusFromGRPCToDB(status catalogv1.ProductStatus) (db.ProductStatus, error) {
	switch status {
	case catalogv1.ProductStatus_PRODUCT_STATUS_DRAFT:
		return db.ProductStatusDraft, nil
	case catalogv1.ProductStatus_PRODUCT_STATUS_PUBLISHED:
		return db.ProductStatusPublished, nil
	case catalogv1.ProductStatus_PRODUCT_STATUS_ARCHIVED:
		return db.ProductStatusArchived, nil

	case catalogv1.ProductStatus_PRODUCT_STATUS_UNSPECIFIED:
		return db.ProductStatusDraft, fmt.Errorf("product status is required")
	default:
		return db.ProductStatusDraft, fmt.Errorf("unknown product status %v", status)
	}
}
