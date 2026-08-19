package server

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"catalog/internal/config"
	"catalog/internal/db"
	pb "catalog/internal/grpc/catalog/v1"
	"catalog/internal/interceptors"
	"catalog/internal/services"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var methodsWithAuth = map[string]bool{
	"/catalog.v1.CatalogService/CreateCategory": true,
	"/catalog.v1.CatalogService/GetCategories":  false,
}

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
		logger.Error("failed to start net.Listen")
		os.Exit(1)
	}

	logger.Info("Create server")
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			interceptors.GetLoggingInterceptor(logger),
			interceptors.GetAuthUnaryInterceptor(methodsWithAuth),
		),
	}
	grpcServer := grpc.NewServer(opts...)

	productService := services.NewProductService(queries)
	categoryService := services.NewCategoryService(queries)

	pb.RegisterCatalogServiceServer(grpcServer, NewServer(dbpool, productService, categoryService))
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
