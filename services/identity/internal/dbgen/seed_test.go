package dbgen

import (
	"context"
	"errors"
	"identity/internal/errs"
	"identity/internal/token"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

func seedUser(userRepo UserRepository) (*User, error) {
	testUsername := "user"
	testEmail := "user@example.com"
	rawPassword := "P@ssw0rd"

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := userRepo.CreateUser(context.Background(), CreateUserParams{
		Username: testUsername,
		Email:    testEmail,
		Password: string(hashedBytes),
	})
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func seedTokens(ctx context.Context, refreshTokenRepo RefreshTokenRepository, JWT_SECRET []byte, user_id pgtype.UUID) (string, string, error) {
	refreshToken, err := token.GenerateRefreshToken(32)
	if err != nil {
	}

	_, err = refreshTokenRepo.CreateRefreshToken(ctx, CreateRefreshTokenParams{
		UserID: user_id,
		Token:  refreshToken,
		ExpiredAt: pgtype.Timestamptz{
			Time:  time.Now().Add(time.Hour * 24 * 7),
			Valid: true,
		},
	})
	if err != nil {
		if errs.IsPostgresErrorByMap(err, errs.UniqueViolation) {
			return "", "", errors.New("failed to create refresh token (unique token error), try again")
		}

		return "", "", errors.New("failed to create refresh token, try again")
	}

	accessToken, err := token.GenerateAccessToken(JWT_SECRET, user_id.String())
	if err != nil {
		return "", "", errors.New("failed to generate access token, try again")
	}

	return refreshToken, accessToken, nil
}
