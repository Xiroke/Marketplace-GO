package server

import (
	"catalog/internal/config"
	"catalog/internal/db"
	pb "catalog/internal/grpc/v1"
	"catalog/internal/interceptors"
	"catalog/internal/services"
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var methodsWithAuth = map[string]bool{
	"/identity.v1.CatalogService/CreateCategory":              true,
	"/identity.v1.CatalogService/CreateProduct":           true,
	"/identity.v1.CatalogService/GetProduct": true,
	"/identity.v1.CatalogService/GetProductsByCategories":             true,
	"/identity.v1.CatalogService/GetProductsByCreator":             true,
}

func RunServer() {
    config := config.NewConfig()

    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
    slog.SetDefault(logger)

    dbpool, err := pgxpool.New(context.Background(), config.DB.DATABASE_URL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

    queries := db.New(dbpool)

    port := 50051
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptors.LoggingInterceptor(logger)),
		grpc.UnaryInterceptor(interceptors.GetAuthUnaryInterceptor(methodsWithAuth, config, queries)),
	}

    grpcServer := grpc.NewServer(opts...)
    productService := services.NewProductService(queries)
    catategoryService := services.NewCategoryService(queries)
    pb.RegisterCatalogServiceServer(grpcServer, NewServer(dbpool, productService, catategoryService))
	reflection.Register(grpcServer)
	logger.Info("Run server")
	err = grpcServer.Serve(lis)
    if err != nil {
        panic("error to run server")
    }
}
