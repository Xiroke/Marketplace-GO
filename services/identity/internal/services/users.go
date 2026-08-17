package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"identity/internal/config"
	"identity/internal/db"
	"identity/internal/errs"
	pb "identity/internal/grpc/identity/v1"
	"identity/internal/interceptors"
	"identity/internal/token"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"
)

//mockery:generate: true
type UserRepository interface {
	GetUser(ctx context.Context, id pgtype.UUID) (db.User, error)
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
}

//mockery:generate: true
type RefreshTokenRepository interface {
	GetUserByRefreshToken(ctx context.Context, token string) (db.User, error)
	ExistRefreshTokenByUser(ctx context.Context, arg db.ExistRefreshTokenByUserParams) (bool, error)
	CreateRefreshToken(ctx context.Context, arg db.CreateRefreshTokenParams) (db.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, token string) error
}

type UserService struct {
	logger           *slog.Logger
	userRepo         UserRepository
	refreshTokenRepo RefreshTokenRepository
	config           *config.Config
}

func NewUserService(logger *slog.Logger, userRepo UserRepository, refreshTokenRepo RefreshTokenRepository, config *config.Config) *UserService {
	return &UserService{
		logger:           logger,
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		config:           config,
	}
}

func (u *UserService) Login(ctx context.Context, request *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, err := u.userRepo.GetUserByEmail(ctx, request.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &errs.AppError{Code: codes.Unauthenticated, Msg: "invalid credentials"}
		}

		return nil, errs.Internal("failed to get user for login", err)
	}

	if !u.checkUserPassword(user.Password, request.Password) {
		return nil, &errs.AppError{Code: codes.Unauthenticated, Msg: "invalid credentials"}
	}

	refreshToken, accessToken, err := u.createUserTokens(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	u.logger.Debug("user logined successfully",
		"user_id", user.ID.String(),
	)

	return &pb.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserId:       user.ID.String(),
	}, nil
}

func (u *UserService) Logout(ctx context.Context, request *pb.LogoutRequest) (*emptypb.Empty, error) {
	userClaims, ok := ctx.Value(interceptors.UserKey).(token.UserClaims)
	if !ok {
		return nil, &errs.AppError{Code: codes.Unauthenticated, Msg: "unauthenticated"}
	}

	var userUUID pgtype.UUID
	err := userUUID.Scan(userClaims.UserID)
	if err != nil {
		return nil, errs.Internal("invalid user ID format", err)
	}

	ok, err = u.refreshTokenRepo.ExistRefreshTokenByUser(ctx, db.ExistRefreshTokenByUserParams{
		UserID: userUUID,
		Token:  request.RefreshToken,
	})
	if err != nil {
		return nil, errs.Internal("failed to check user refresh token", err)
	}
	if !ok {
		return nil, &errs.AppError{Code: codes.Unauthenticated, Msg: "invalid token"}
	}

	err = u.refreshTokenRepo.DeleteRefreshToken(ctx, request.RefreshToken)
	if err != nil {
		return nil, errs.Internal("failed to delete refresh token", err)
	}

	u.logger.Debug("user logout successfully",
		"user_id", userClaims.ID,
	)

	return &emptypb.Empty{}, nil
}

func (u *UserService) Register(ctx context.Context, request *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	hashedPassword, err := u.hashUserPassword(request.Password)
	if err != nil {
		return nil, errs.Internal("failed to hash password", err)
	}

	user, err := u.userRepo.CreateUser(ctx, db.CreateUserParams{
		Username: request.Username,
		Email:    request.Email,
		Password: hashedPassword,
	})
	if err != nil {
		if errs.IsPostgresErrorByMap(err, errs.UniqueViolation) {
			var pgErr *pgconn.PgError

			if !errors.As(err, &pgErr) {
				return nil, errs.Internal("failed to create user", err)
			}

			u.logger.Warn("registration attempt with duplicate value",
				"constraint", pgErr.ConstraintName,
				"username", request.Username,
				"email", request.Email,
			)

			field := u.getFieldNameFromConstraint(pgErr.ConstraintName)

			return nil, &errs.AppError{Code: codes.AlreadyExists, Msg: fmt.Sprintf("%v must be unique", field)}
		}

		return nil, errs.Internal("failed to create user", err)
	}

	refreshToken, accessToken, err := u.createUserTokens(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	u.logger.Debug("user registered successfully",
		"user_id", user.ID.String(),
	)

	return &pb.RegisterResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserId:       user.ID.String(),
	}, nil
}

func (u *UserService) RefreshAccessToken(ctx context.Context, request *pb.RefreshAccessTokenRequest) (*pb.RefreshAccessTokenResponse, error) {
	user, err := u.refreshTokenRepo.GetUserByRefreshToken(ctx, request.RefreshToken)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.Internal("invalid refresh token", err)
		}

		return nil, errs.Internal("internal server error", err)
	}

	accessToken, err := token.GenerateAccessToken([]byte(u.config.JWTSecret), user.ID.String())
	if err != nil {
		return nil, errs.Internal("failed to generate access token, try again", err)
	}

	u.logger.Debug("user refresh access successfully",
		"user_id", user.ID.String(),
	)

	return &pb.RefreshAccessTokenResponse{AccessToken: accessToken}, nil
}

func (u *UserService) GetUserByAccess(ctx context.Context, request *pb.GetUserByAccessRequest) (*pb.GetUserByAccessResponse, error) {
	userData, err := token.DecodeAccessToken([]byte(u.config.JWTSecret), request.AccessToken)
	if err != nil {
		return nil, errs.Internal("failed to decode access token", err)
	}

	u.logger.Debug("user refresh access successfully",
		"user_id", userData.UserID,
	)

	return &pb.GetUserByAccessResponse{UserId: userData.UserID}, nil
}

func (u *UserService) hashUserPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (u *UserService) checkUserPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func (u *UserService) getFieldNameFromConstraint(constraint string) string {
	switch {
	case strings.Contains(constraint, "email"):
		return "email"
	case strings.Contains(constraint, "username"):
		return "username"
	}

	return "undefined"
}

func (u *UserService) createUserTokens(ctx context.Context, user_id pgtype.UUID) (string, string, error) {
	refreshToken, err := token.GenerateRefreshToken(32)
	if err != nil {
		return "", "", errs.Internal("failed to generate refresh token, try again", err)
	}

	_, err = u.refreshTokenRepo.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID: user_id,
		Token:  refreshToken,
		ExpiredAt: pgtype.Timestamptz{
			Time:  time.Now().Add(time.Hour * 24 * 7),
			Valid: true,
		},
	})
	if err != nil {
		if errs.IsPostgresErrorByMap(err, errs.UniqueViolation) {
			return "", "", errs.Internal("failed to create refresh token (unique token error), try again", err)
		}

		return "", "", errs.Internal("failed to create refresh token, try again", err)
	}

	accessToken, err := token.GenerateAccessToken([]byte(u.config.JWTSecret), user_id.String())
	if err != nil {
		return "", "", errs.Internal("failed to generate access token, try again", err)
	}

	return refreshToken, accessToken, nil
}
