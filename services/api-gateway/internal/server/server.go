package server

import (
	"api-gateway/internal/config"
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func RunServer() {
    ctx := context.Background()

    config := config.NewConfig()

    r := chi.NewRouter()
    r.Use(middleware.Logger)

    RegisterRoutes(ctx, r, config)

    http.ListenAndServe(":3000", r)
}
