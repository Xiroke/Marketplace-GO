package service

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	dbgen "identity/internal/db"
	"identity/internal/errs"
	pb "identity/internal/grpc/v1"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserService struct {
	logger  *slog.Logger
	queries *dbgen.Queries
}

func NewUserService(logger *slog.Logger, queries *dbgen.Queries) *UserService {
	return &UserService{logger: logger, queries: queries}
}

func (u *UserService) Login(ctx context.Context, request *pb.LoginRequest) (*pb.LoginResponse, error) {
	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "request canceled")
	default:
	}

	if strings.TrimSpace(request.Email) == "" ||
		request.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password must not be empty")
	}

	user, err := u.queries.GetUserByEmail(ctx, request.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Errorf(codes.InvalidArgument, "invalid credentials")
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

	u.logger.Info("user logined successfully",
		"user_id", user.ID.String(),
		"username", user.Username,
		"email", user.Email,
	)

	return &pb.LoginResponse{
		AccessToken:  "",
		RefreshToken: "",
		UserId:       user.ID.String(),
	}, nil
}

func (u *UserService) Logout(ctx context.Context, request *pb.LogoutRequest) (*emptypb.Empty, error) {
	panic("Not implemented")
}

func (u *UserService) Register(ctx context.Context, request *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "request canceled")
	default:
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

	user, err := u.queries.CreateUser(ctx, dbgen.CreateUserParams{
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

	u.logger.Info("user registered successfully",
		"user_id", user.ID.String(),
		"username", user.Username,
		"email", user.Email,
	)

	return &pb.RegisterResponse{
		AccessToken:  "",
		RefreshToken: "",
		UserId:       user.ID.String(),
	}, nil
}

func (u *UserService) RefreshAccessToken(ctx context.Context, request *pb.RefreshAccessTokenRequest) (*pb.RefreshAccessTokenResponse, error) {
	panic("Not implemented")
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
