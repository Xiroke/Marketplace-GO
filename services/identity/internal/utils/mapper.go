package utils

import (
	"fmt"

	"identity/internal/errs"

	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ParseError = func(err error) *errs.AppError {
		return errs.Internal(fmt.Errorf("failed to parse input data: %w", err))
	}
)

func StringToUUID(str string) (pgtype.UUID, *errs.AppError) {
	var uuid pgtype.UUID

	err := uuid.Scan(str)
	if err != nil {
		return pgtype.UUID{Valid: false}, ParseError(err)
	}

	return uuid, nil
}
