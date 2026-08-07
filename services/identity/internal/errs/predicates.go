package errs

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

type PostgresErrMapKey string

var (
	UniqueViolation PostgresErrMapKey = "unique_violation"
)

var PostgresErrMap = map[PostgresErrMapKey]string{
	UniqueViolation: "23505",
}

func IsPostgresErrorByMap(err error, key PostgresErrMapKey) bool {
	var pgErr *pgconn.PgError

	if !errors.As(err, &pgErr) {
		return false
	}

	code, exists := PostgresErrMap[key]
	if !exists {
		return false
	}

	return pgErr.Code == code
}
