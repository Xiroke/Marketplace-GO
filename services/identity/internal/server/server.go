package server

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	emptypb "google.golang.org/protobuf/types/known/emptypb"

	"identity/internal/config"
	dbgen "identity/internal/db"
	pb "identity/internal/grpc/v1"
	"identity/internal/interceptors"
	"identity/internal/service"
)

var _ pb.AuthServiceServer = (*server)(nil)

type server struct {
	pb.UnsafeAuthServiceServer
	db          *pgxpool.Pool
	authService *service.UserService
}

func newServer(db *pgxpool.Pool, authService *service.UserService) *server {
	return &server{db: db, authService: authService}
}

func (s *server) Login(ctx context.Context, request *pb.LoginRequest) (*pb.LoginResponse, error) {
	return s.authService.Login(ctx, request)
}

func (s *server) Logout(ctx context.Context, request *pb.LogoutRequest) (*emptypb.Empty, error) {
	return s.authService.Logout(ctx, request)
}

func (s *server) Register(ctx context.Context, request *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	return s.authService.Register(ctx, request)
}

func (s *server) RefreshAccessToken(ctx context.Context, request *pb.RefreshAccessTokenRequest) (*pb.RefreshAccessTokenResponse, error) {
	return s.authService.RefreshAccessToken(ctx, request)
}

var methodsWithAuth = map[string]bool{
	"/identity.v1.AuthService/Login":              false,
	"/identity.v1.AuthService/Register":           false,
	"/identity.v1.AuthService/RefreshAccessToken": false,
	"/identity.v1.AuthService/Logout":             true,
}

func StartServer() {
	config := config.NewConfig()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))
	slog.SetDefault(logger)

	dbpool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	queries := dbgen.New(dbpool)

	port := 8000
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptors.GetAuthUnaryInterceptor(methodsWithAuth, queries)),
	}

	grpcServer := grpc.NewServer(opts...)
	authService := service.NewUserService(logger, queries, config)
	pb.RegisterAuthServiceServer(grpcServer, newServer(dbpool, authService))
	grpcServer.Serve(lis)
}
