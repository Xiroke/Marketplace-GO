package server

import (
	"context"
	"log/slog"
	"net"
	"os"

	"catalog/internal/config"
	"catalog/internal/db"
	pb "catalog/internal/grpc/catalog/v1"
	"catalog/internal/interceptors"
	"catalog/internal/services"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func RunServer() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	slog.SetDefault(logger)

	config, err := config.NewConfig()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	dbpool, err := pgxpool.New(context.Background(), config.DB.DSN())
	if err != nil {
		logger.Error("Unable to create connection pool", "error", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	queries := db.New(dbpool)

	lis, err := net.Listen("tcp", config.App.Address())
	if err != nil {
		logger.Error("failed to start new.Listen")
		os.Exit(1)
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptors.LoggingInterceptor(logger)),
	}
	grpcServer := grpc.NewServer(opts...)

	productService := services.NewProductService(queries)
	catategoryService := services.NewCategoryService(queries)

	pb.RegisterCatalogServiceServer(grpcServer, NewServer(dbpool, productService, catategoryService))
	reflection.Register(grpcServer)

	logger.Info("Run server")
	err = grpcServer.Serve(lis)
	if err != nil {
		logger.Error("failed to run grpc server")
		os.Exit(1)
	}
}
