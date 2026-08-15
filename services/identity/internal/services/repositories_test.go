package services

import (
	"context"
	"testing"
	"time"

	"identity/internal/db"
	"identity/internal/tests/testutils"
	"identity/internal/token"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestUserRepository_CreateUser(t *testing.T) {
	testutils.CleanDB(t, testDBPool)
	queries := db.New(testDBPool)

	testUsername := "user"
	testEmail := "user@example.com"
	rawPassword := "P@ssw0rd"

	var userRepo UserRepository = queries

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	require.NoError(t, err, "failed to hash password for test")
	res, err := userRepo.CreateUser(context.Background(), db.CreateUserParams{
		Username: testUsername,
		Email:    testEmail,
		Password: string(hashedBytes),
	})

	require.NoError(t, err, "failed to create user")
	require.NotNil(t, res)
	require.Equal(t, res.Username, testUsername)
	require.Equal(t, res.Email, testEmail)
	require.Equal(t, res.Password, string(hashedBytes))
}

func TestUserRepository_GetUser(t *testing.T) {
	testutils.CleanDB(t, testDBPool)
	queries := db.New(testDBPool)

	var userRepo UserRepository = queries
	user, err := seedUser(userRepo)
	require.NoError(t, err, "failed to seed user")

	res, err := userRepo.GetUser(context.Background(), user.ID)
	require.NoError(t, err, "failed to get user")
	require.NotNil(t, res)
	require.Equal(t, res.Username, user.Username)
	require.Equal(t, res.Email, user.Email)
	require.Equal(t, res.Password, user.Password)
}

func TestUserRepository_GetUserByEmail(t *testing.T) {
	testutils.CleanDB(t, testDBPool)
	queries := db.New(testDBPool)

	var userRepo UserRepository = queries
	user, err := seedUser(userRepo)
	require.NoError(t, err, "failed to seed user")

	res, err := userRepo.GetUserByEmail(context.Background(), user.Email)
	require.NoError(t, err, "failed to get user")
	require.NotNil(t, res)
	require.Equal(t, res.Username, user.Username)
	require.Equal(t, res.Email, user.Email)
	require.Equal(t, res.Password, user.Password)
}

func TestRefreshTokenRepository_GetUserByEmail(t *testing.T) {
	testutils.CleanDB(t, testDBPool)
	queries := db.New(testDBPool)

	var userRepo UserRepository = queries
	var refreshTokenRepo RefreshTokenRepository = queries
	user, err := seedUser(userRepo)
	require.NoError(t, err, "failed to seed user")
	refreshToken, _, err := seedTokens(context.Background(), refreshTokenRepo, []byte("secret"), user.ID)
	require.NoError(t, err, "failed to seed tokens")

	res, err := refreshTokenRepo.GetUserByRefreshToken(context.Background(), refreshToken)
	require.NoError(t, err, "failed to get user by email")
	require.NotNil(t, res)
	require.Equal(t, res.Username, user.Username)
	require.Equal(t, res.Email, user.Email)
	require.Equal(t, res.Password, user.Password)
}

func TestRefreshTokenRepository_ExistRefreshTokenByUser(t *testing.T) {
	testutils.CleanDB(t, testDBPool)
	queries := db.New(testDBPool)

	var userRepo UserRepository = queries
	var refreshTokenRepo RefreshTokenRepository = queries
	user, err := seedUser(userRepo)
	require.NoError(t, err, "failed to seed user")
	refreshToken, _, err := seedTokens(context.Background(), refreshTokenRepo, []byte("secret"), user.ID)
	require.NoError(t, err, "failed to seed tokens")

	res, err := refreshTokenRepo.ExistRefreshTokenByUser(context.Background(), db.ExistRefreshTokenByUserParams{
		UserID: user.ID,
		Token:  refreshToken,
	})
	require.NoError(t, err, "failed check exist refresh token")
	require.NotNil(t, res)
	require.Equal(t, res, true)
}

func TestRefreshTokenRepository_CreateRefreshToken(t *testing.T) {
	testutils.CleanDB(t, testDBPool)
	queries := db.New(testDBPool)

	var userRepo UserRepository = queries
	var refreshTokenRepo RefreshTokenRepository = queries
	user, err := seedUser(userRepo)
	require.NoError(t, err, "failed to seed user")

	refreshToken, err := token.GenerateRefreshToken(32)
	require.NoError(t, err, "failed to seed user")

	res, err := refreshTokenRepo.CreateRefreshToken(context.Background(), db.CreateRefreshTokenParams{
		Token:  refreshToken,
		UserID: user.ID,
		ExpiredAt: pgtype.Timestamptz{
			Time:  time.Now().Add(time.Hour * 24 * 7),
			Valid: true,
		},
	})
	require.NoError(t, err, "failed create refresh token")
	require.NotNil(t, res)
	require.Equal(t, res.Token, refreshToken)
}

func TestRefreshTokenRepository_DeleteRefreshToken(t *testing.T) {
	testutils.CleanDB(t, testDBPool)
	queries := db.New(testDBPool)

	var userRepo UserRepository = queries
	var refreshTokenRepo RefreshTokenRepository = queries
	user, err := seedUser(userRepo)
	require.NoError(t, err, "failed to seed user")
	refreshToken, _, err := seedTokens(context.Background(), refreshTokenRepo, []byte("secret"), user.ID)
	require.NoError(t, err, "failed to seed tokens")

	err = refreshTokenRepo.DeleteRefreshToken(context.Background(), refreshToken)
	require.NoError(t, err, "failed delete refresh token")
	ok, err := refreshTokenRepo.ExistRefreshTokenByUser(context.Background(), db.ExistRefreshTokenByUserParams{
		UserID: user.ID,
		Token:  refreshToken,
	})
	require.NoError(t, err, "failed to get user")
	require.NotNil(t, ok)
	require.Equal(t, ok, false)
}
