package dbgen

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

//mockery:generate: true
type UserRepository interface {
	GetUser(ctx context.Context, id pgtype.UUID) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
}

//mockery:generate: true
type RefreshTokenRepository interface {
	GetUserByRefreshToken(ctx context.Context, token string) (User, error)
	ExistRefreshTokenByUser(ctx context.Context, arg ExistRefreshTokenByUserParams) (bool, error)
	CreateRefreshToken(ctx context.Context, arg CreateRefreshTokenParams) (RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, token string) error
}
