package server

import (
	"api-gateway/internal/api/auth"
	"api-gateway/internal/api/categories"
	"api-gateway/internal/api/products"
	"api-gateway/internal/config"
	catalogv1 "api-gateway/internal/grpc/catalog/v1"
	identityv1 "api-gateway/internal/grpc/identity/v1"
	"api-gateway/internal/middleware"
	"context"
	"log/slog"
	"net/http"

	_ "api-gateway/internal/swagger"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

func RegisterRoutes(
	ctx context.Context,
	r *chi.Mux,
	config *config.Config,
	logger *slog.Logger,
	identityClient identityv1.AuthServiceClient,
	catalogClient catalogv1.CatalogServiceClient,
) {
	r.Post("/api/v1/auth/login", auth.LoginHandler(logger, identityClient))
	r.Post("/api/v1/auth/register", auth.RegisterHandler(logger, identityClient))

	r.Get("/api/v1/auth/categories", categories.GetCategoriesHandler(logger, identityClient, catalogClient))

	r.Group(func(r chi.Router) {
		r.Use(middleware.GetAuthMiddleware(config.JWTSecret))

		r.Post("/api/v1/auth/categories", categories.CreateCategoryHandler(logger, identityClient, catalogClient))
		r.Post("/api/v1/auth/products", products.CreateProductHandler(logger, identityClient, catalogClient))
	})

	r.Get("/swagger/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./internal/swagger/swagger.yaml")
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/swagger.yaml"),
	))
}
