package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"

	"identity/internal/config"
	"identity/internal/db"
	pb "identity/internal/grpc/identity/v1"
	"identity/internal/interceptors"
	"identity/internal/services"
)

var _ pb.AuthServiceServer = (*server)(nil)

type server struct {
	pb.UnsafeAuthServiceServer
	db          *pgxpool.Pool
	authService *services.UserService
}

func newServer(db *pgxpool.Pool, authService *services.UserService) *server {
	return &server{db: db, authService: authService}
}

func (s *server) Login(ctx context.Context, request *pb.LoginRequest) (*pb.LoginResponse, error) {
	if strings.TrimSpace(request.Email) == "" ||
		request.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password must not be empty")
	}

	return s.authService.Login(ctx, request)
}

func (s *server) Logout(ctx context.Context, request *pb.LogoutRequest) (*emptypb.Empty, error) {
	return s.authService.Logout(ctx, request)
}

func (s *server) Register(ctx context.Context, request *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if strings.TrimSpace(request.Username) == "" ||
		strings.TrimSpace(request.Email) == "" ||
		request.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username, email and password must not be empty")
	}
	return s.authService.Register(ctx, request)
}

func (s *server) RefreshAccessToken(ctx context.Context, request *pb.RefreshAccessTokenRequest) (*pb.RefreshAccessTokenResponse, error) {
	return s.authService.RefreshAccessToken(ctx, request)
}

func (s *server) GetUserByAccess(ctx context.Context, request *pb.GetUserByAccessRequest) (*pb.GetUserByAccessResponse, error) {
	return s.authService.GetUserByAccess(ctx, request)
}

var methodsWithAuth = map[string]bool{
	"/identity.v1.AuthService/Login":              false,
	"/identity.v1.AuthService/Register":           false,
	"/identity.v1.AuthService/RefreshAccessToken": false,
	"/identity.v1.AuthService/Logout":             true,
	"/identity.v1.AuthService/GetUserByAccess":    true,
}

func StartServer() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	slog.SetDefault(logger)

	config, err := config.NewConfig()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	dbpool, err := pgxpool.New(context.Background(), config.DB.DSN())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	queries := db.New(dbpool)

	lis, err := net.Listen("tcp", config.App.Address())
	if err != nil {
		logger.Error("failed to start net.Listen", "error", err)
		os.Exit(1)
	}
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptors.LoggingInterceptor(logger)),
		grpc.UnaryInterceptor(interceptors.GetAuthUnaryInterceptor(methodsWithAuth, config, queries)),
	}

	logger.Info("Create server")
	grpcServer := grpc.NewServer(opts...)

	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	authService := services.NewUserService(logger, queries, queries, config)
	pb.RegisterAuthServiceServer(grpcServer, newServer(dbpool, authService))
	reflection.Register(grpcServer)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownComplete := make(chan struct{})

	go func() {
		<-ctx.Done()
		logger.Info("Get stop signal")

		stopped := make(chan struct{})
		go func() {
			grpcServer.GracefulStop()
			close(stopped)
		}()

		select {
		case <-time.After(10 * time.Second):
			logger.Info("Timeout GracefulStop, immediately stopping...")
			grpcServer.Stop()
		case <-stopped:
			logger.Info("Successfully stopping server")
		}

		close(shutdownComplete)
	}()

	logger.Info("Run server")

	err = grpcServer.Serve(lis)
	if err != nil {
		logger.Error("failed to run grpc server")
		os.Exit(1)
	}

	<-shutdownComplete
	logger.Info("Server process exited cleanly")
}
