package service

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"identity/internal/config"
	dbgen "identity/internal/dbgen"
	"identity/internal/errs"
	pb "identity/internal/grpc/v1"
	"identity/internal/interceptors"
	"identity/internal/token"
	"identity/internal/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserService struct {
	logger           *slog.Logger
	userRepo         dbgen.UserRepository
	refreshTokenRepo dbgen.RefreshTokenRepository
	config           *config.Config
}

func NewUserService(logger *slog.Logger, userRepo dbgen.UserRepository, refreshTokenRepo dbgen.RefreshTokenRepository, config *config.Config) *UserService {
	return &UserService{
		logger:           logger,
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		config:           config,
	}
}

func (u *UserService) Login(ctx context.Context, request *pb.LoginRequest) (*pb.LoginResponse, error) {
	if err := utils.CloseCanceledRequestByContext(ctx); err != nil {
		return nil, err
	}

	if strings.TrimSpace(request.Email) == "" ||
		request.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password must not be empty")
	}

	user, err := u.userRepo.GetUserByEmail(ctx, request.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
		}

		u.logger.Error("failed to get user for login",
			"error", err,
			"email", request.Email,
		)
		return nil, status.Errorf(codes.Internal, "failed to get user for login")
	}

	if !u.checkUserPassword(user.Password, request.Password) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid credentials")
	}

	refreshToken, accessToken, err := u.createUserTokens(ctx, user.ID, request.Email)
	if err != nil {
		return nil, err
	}

	u.logger.Info("user logined successfully",
		"user_id", user.ID.String(),
		"username", user.Username,
		"email", user.Email,
	)

	return &pb.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserId:       user.ID.String(),
	}, nil
}

func (u *UserService) Logout(ctx context.Context, request *pb.LogoutRequest) (*emptypb.Empty, error) {
	if err := utils.CloseCanceledRequestByContext(ctx); err != nil {
		return nil, err
	}

	userClaims, ok := ctx.Value(interceptors.UserKey).(token.UserClaims)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	var userUUID pgtype.UUID
	err := userUUID.Scan(userClaims.UserID)
	if err != nil {
		u.logger.Error("failed to parse user_id as UUID", "error", err)
		return nil, status.Error(codes.Internal, "invalid user ID format")
	}

	ok, err = u.refreshTokenRepo.ExistRefreshTokenByUser(ctx, dbgen.ExistRefreshTokenByUserParams{
		UserID: userUUID,
		Token:  request.RefreshToken,
	})
	if err != nil {
		u.logger.Error("failed to check user refresh token", "error", err)
		return nil, status.Error(codes.Internal, "failed to check user refresh token")
	}
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	err = u.refreshTokenRepo.DeleteRefreshToken(ctx, request.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to delete refresh token")
	}

	u.logger.Info("user logout successfully",
		"user_id", userClaims.ID,
	)

	return &emptypb.Empty{}, nil
}

func (u *UserService) Register(ctx context.Context, request *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if err := utils.CloseCanceledRequestByContext(ctx); err != nil {
		return nil, err
	}

	if strings.TrimSpace(request.Username) == "" ||
		strings.TrimSpace(request.Email) == "" ||
		request.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username, email and password must not be empty")
	}

	hashedPassword, err := u.hashUserPassword(request.Password)
	if err != nil {
		u.logger.Error("failed to hash password", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to hash password")
	}

	user, err := u.userRepo.CreateUser(ctx, dbgen.CreateUserParams{
		Username: request.Username,
		Email:    request.Email,
		Password: hashedPassword,
	})
	if err != nil {
		if errs.IsPostgresErrorByMap(err, errs.UniqueViolation) {
			var pgErr *pgconn.PgError

			if !errors.As(err, &pgErr) {
				u.logger.Error("failed to parse postgres error", "error", err)
				return nil, status.Errorf(codes.Internal, "failed to create user")
			}

			u.logger.Warn("registration attempt with duplicate value",
				"constraint", pgErr.ConstraintName,
				"username", request.Username,
				"email", request.Email,
			)

			field := u.getFieldNameFromConstraint(pgErr.ConstraintName)

			return nil, status.Errorf(codes.AlreadyExists, "%v must be unique", field)
		}

		u.logger.Error("failed to create user",
			"error", err,
			"username", request.Username,
			"email", request.Email,
		)
		return nil, status.Errorf(codes.Internal, "failed to create user")
	}

	refreshToken, accessToken, err := u.createUserTokens(ctx, user.ID, request.Email)
	if err != nil {
		return nil, err
	}

	u.logger.Info("user registered successfully",
		"user_id", user.ID.String(),
		"username", user.Username,
		"email", user.Email,
	)

	return &pb.RegisterResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserId:       user.ID.String(),
	}, nil
}

func (u *UserService) RefreshAccessToken(ctx context.Context, request *pb.RefreshAccessTokenRequest) (*pb.RefreshAccessTokenResponse, error) {
	if err := utils.CloseCanceledRequestByContext(ctx); err != nil {
		return nil, err
	}

	user, err := u.refreshTokenRepo.GetUserByRefreshToken(ctx, request.RefreshToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid refresh token")
	}

	accessToken, err := token.GenerateAccessToken(u.config.JWT_SECRET, user.ID.String())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate access token, try again")
	}

	u.logger.Info("user refresh access successfully",
		"user_id", user.ID.String(),
		"username", user.Username,
		"email", user.Email,
	)

	return &pb.RefreshAccessTokenResponse{AccessToken: accessToken}, nil
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

func (u *UserService) createUserTokens(ctx context.Context, user_id pgtype.UUID, email string) (string, string, error) {
	refreshToken, err := token.GenerateRefreshToken(32)
	if err != nil {
		u.logger.Error("failed to generate refresh token",
			"error", err,
			"email", email,
		)
		return "", "", status.Errorf(codes.Internal, "failed to generate refresh token, try again")
	}

	_, err = u.refreshTokenRepo.CreateRefreshToken(ctx, dbgen.CreateRefreshTokenParams{
		UserID: user_id,
		Token:  refreshToken,
		ExpiredAt: pgtype.Timestamptz{
			Time:  time.Now().Add(time.Hour * 24 * 7),
			Valid: true,
		},
	})
	if err != nil {
		if errs.IsPostgresErrorByMap(err, errs.UniqueViolation) {
			return "", "", status.Errorf(codes.Internal, "failed to create refresh token (unique token error), try again")
		}

		return "", "", status.Errorf(codes.Internal, "failed to create refresh token, try again")
	}

	accessToken, err := token.GenerateAccessToken(u.config.JWT_SECRET, user_id.String())
	if err != nil {
		return "", "", status.Errorf(codes.Internal, "failed to generate access token, try again")
	}

	return refreshToken, accessToken, nil
}
