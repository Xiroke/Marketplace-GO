package utils

import (
	"api-gateway/internal/errs"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
    ParseError = func (err error) *errs.AppError {return errs.Internal("failed to parse input data", err)}
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

func NumericToString(numeric pgtype.Numeric) (string, *errs.AppError){
    str, err := numeric.Value()
    if err != nil {
        return "", ParseError(err)
    }

    return str.(string), nil
}

func TimestamptzToGRPCTimestamp(timestamp pgtype.Timestamptz) *timestamppb.Timestamp {
    return  timestamppb.New(timestamp.Time)
}
