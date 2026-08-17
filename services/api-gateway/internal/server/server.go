package server

import (
	"api-gateway/internal/config"
	identityv1 "api-gateway/internal/grpc/identity/v1"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func RunServer() {
	ctx := context.Background()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	slog.SetDefault(logger)

	config, err := config.NewConfig()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	connAuthClient, err := grpc.NewClient(config.AddressIdentityService, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(fmt.Sprintf("failed to run grcp client %v", err))
	}
	defer connAuthClient.Close()

	identityClient := identityv1.NewAuthServiceClient(connAuthClient)

	logger.Info("Create api-gateway server")
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	RegisterRoutes(ctx, r, config, identityClient)

	logger.Info("Run api-gateway server")
	Address := fmt.Sprintf(":%s", config.Port)
	err = http.ListenAndServe(Address, r)
	if err != nil {
		logger.Error("failed to run api-gateway server", "error", err)
		os.Exit(1)
	}
}
