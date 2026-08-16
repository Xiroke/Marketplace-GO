package server

import (
	"api-gateway/internal/config"
	identityv1 "api-gateway/internal/grpc/v1"
	"context"
	"encoding/json"
	"net/http"
	"time"

	_ "api-gateway/internal/swagger"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
	"google.golang.org/grpc"
)

type LoginRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

func RegisterRoutes(ctx context.Context, r *chi.Mux, config *config.Config) {
    connAuthClient, err := grpc.NewClient(config.ADDRESS_AUTH_SERVICE)
    if err != nil {
        panic("failed to run grcp client")
    }
    defer connAuthClient.Close()


    identityClient := identityv1.NewAuthServiceClient(connAuthClient)

    r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
        defer r.Body.Close()

        ctx, _ := context.WithTimeout(r.Context(), time.Second*5)

        var req LoginRequest
        err := json.NewDecoder(r.Body).Decode(&req)
        if err != nil {
            http.Error(w, "Invalid JSON data", http.StatusBadRequest)
            return
        }

        res, err := identityClient.Login(ctx, &identityv1.LoginRequest{
            Email: req.Email,
            Password: req.Password,
        })
        if err != nil {
            http.Error(w, "Failed to login", http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusOK)
    })

    r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:1323/swagger/doc.json"), //The url pointing to API definition
	))
}
