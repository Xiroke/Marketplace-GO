package dbgen

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepository interface {
	GetUser(ctx context.Context, id pgtype.UUID) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
}

type RefreshTokenRepository interface {
	GetUserByRefreshToken(ctx context.Context, token string) (User, error)
	ExistRefreshTokenByUser(ctx context.Context, arg ExistRefreshTokenByUserParams) (bool, error)
	CreateRefreshToken(ctx context.Context, arg CreateRefreshTokenParams) (RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, token string) error
}
