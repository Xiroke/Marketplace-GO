package service

import (
	"context"
	"identity/internal/config"
	"identity/internal/dbgen"
	pb "identity/internal/grpc/v1"
	"identity/internal/interceptors"
	"identity/internal/service/mocks"
	"identity/internal/token"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestUserService_Login_Success(t *testing.T) {
	cfg := &config.Config{
		JWT_SECRET: []byte("secret"),
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))
	slog.SetDefault(logger)

	userMockRepo := mocks.NewMockUserRepository(t)
	refreshTokenMockRepo := mocks.NewMockRefreshTokenRepository(t)

	testEmail := "user@example.com"
	rawPassword := "raw_password"
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	require.NoError(t, err, "failed to hash password for test")

	expectedUser := dbgen.User{
		ID:       pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
		Username: "user",
		Email:    testEmail,
		Password: string(hashedBytes),
	}

	userMockRepo.EXPECT().
		GetUserByEmail(mock.Anything, testEmail).
		Return(expectedUser, nil).
		Once()

	refreshToken, err := token.GenerateRefreshToken(32)
	require.NoError(t, err, "failed to generate refresh token")
	refreshTokenMockRepo.EXPECT().
		CreateRefreshToken(mock.Anything, mock.AnythingOfType("dbgen.CreateRefreshTokenParams")).
		Return(dbgen.RefreshToken{
			Token:  refreshToken,
			UserID: expectedUser.ID,
			ExpiredAt: pgtype.Timestamptz{
				Time:  time.Now().Add(time.Hour * 24 * 7),
				Valid: true,
			},
		}, nil).
		Once()

	service := NewUserService(logger, userMockRepo, refreshTokenMockRepo, cfg)

	req := &pb.LoginRequest{
		Email:    testEmail,
		Password: rawPassword,
	}

	ctx := context.Background()
	res, err := service.Login(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, res)
	require.NotEmpty(t, res.AccessToken, "access token should not be empty")
	require.NotEmpty(t, res.RefreshToken, "refresh token should not be empty")
}

func TestUserService_Register_Success(t *testing.T) {
	cfg := &config.Config{
		JWT_SECRET: []byte("secret"),
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))
	slog.SetDefault(logger)

	userMockRepo := mocks.NewMockUserRepository(t)
	refreshTokenMockRepo := mocks.NewMockRefreshTokenRepository(t)

	testEmail := "user@example.com"
	rawPassword := "raw_password"
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	require.NoError(t, err, "failed to hash password for test")

	expectedUser := dbgen.User{
		ID:       pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
		Username: "user",
		Email:    testEmail,
		Password: string(hashedBytes),
	}

	userMockRepo.EXPECT().
		CreateUser(mock.Anything, mock.AnythingOfType("dbgen.CreateUserParams")).
		Return(expectedUser, nil).
		Once()

	refreshToken, err := token.GenerateRefreshToken(32)
	require.NoError(t, err, "failed to generate refresh token")
	refreshTokenMockRepo.EXPECT().
		CreateRefreshToken(mock.Anything, mock.AnythingOfType("dbgen.CreateRefreshTokenParams")).
		Return(dbgen.RefreshToken{
			Token:  refreshToken,
			UserID: expectedUser.ID,
			ExpiredAt: pgtype.Timestamptz{
				Time:  time.Now().Add(time.Hour * 24 * 7),
				Valid: true,
			},
		}, nil).
		Once()

	service := NewUserService(logger, userMockRepo, refreshTokenMockRepo, cfg)

	req := &pb.RegisterRequest{
		Username: "user",
		Email:    testEmail,
		Password: rawPassword,
	}

	ctx := context.Background()
	res, err := service.Register(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, res)
	require.NotEmpty(t, res.AccessToken, "access token should not be empty")
	require.NotEmpty(t, res.RefreshToken, "refresh token should not be empty")
}

func TestUserService_Logout_Success(t *testing.T) {
	cfg := &config.Config{
		JWT_SECRET: []byte("secret"),
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))
	slog.SetDefault(logger)

	userMockRepo := mocks.NewMockUserRepository(t)
	refreshTokenMockRepo := mocks.NewMockRefreshTokenRepository(t)

	testEmail := "user@example.com"
	rawPassword := "raw_password"
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	require.NoError(t, err, "failed to hash password for test")

	expectedUser := dbgen.User{
		ID:       pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
		Username: "user",
		Email:    testEmail,
		Password: string(hashedBytes),
	}
	refreshToken, err := token.GenerateRefreshToken(32)
	require.NoError(t, err, "failed to generate refresh token")

	refreshTokenMockRepo.EXPECT().
		ExistRefreshTokenByUser(mock.Anything, mock.AnythingOfType("dbgen.ExistRefreshTokenByUserParams")).
		Return(true, nil).
		Once()

	refreshTokenMockRepo.EXPECT().
		DeleteRefreshToken(mock.Anything, refreshToken).
		Return(nil).
		Once()

	service := NewUserService(logger, userMockRepo, refreshTokenMockRepo, cfg)

	req := &pb.LogoutRequest{
		RefreshToken: refreshToken,
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, interceptors.UserKey, token.UserClaims{UserID: expectedUser.ID.String()})
	res, err := service.Logout(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestUserService_RefreshAccessToken_Success(t *testing.T) {
	cfg := &config.Config{
		JWT_SECRET: []byte("secret"),
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))
	slog.SetDefault(logger)

	userMockRepo := mocks.NewMockUserRepository(t)
	refreshTokenMockRepo := mocks.NewMockRefreshTokenRepository(t)

	testEmail := "user@example.com"
	rawPassword := "raw_password"
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	require.NoError(t, err, "failed to hash password for test")

	expectedUser := dbgen.User{
		ID:       pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
		Username: "user",
		Email:    testEmail,
		Password: string(hashedBytes),
	}

	refreshToken, err := token.GenerateRefreshToken(32)
	refreshTokenMockRepo.EXPECT().
		GetUserByRefreshToken(mock.Anything, refreshToken).
		Return(expectedUser, nil).
		Once()

	service := NewUserService(logger, userMockRepo, refreshTokenMockRepo, cfg)

	req := &pb.RefreshAccessTokenRequest{
		RefreshToken: refreshToken,
	}

	ctx := context.Background()
	res, err := service.RefreshAccessToken(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, res)
	require.NotEmpty(t, res.AccessToken, "access token should not be empty")
}
