package server

import (
	"api-gateway/internal/api/auth"
	"api-gateway/internal/config"
	identityv1 "api-gateway/internal/grpc/identity/v1"
	"context"
	"net/http"

	_ "api-gateway/internal/swagger"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

func RegisterRoutes(ctx context.Context, r *chi.Mux, config *config.Config, identityClient identityv1.AuthServiceClient) {
    r.Post("/api/v1/auth/login", auth.LoginHandler(identityClient))

    r.Get("/swagger/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "./internal/swagger/swagger.yaml")
    })

    r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/swagger.yaml"),
	))
}
