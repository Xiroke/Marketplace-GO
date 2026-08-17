package auth

import (
	identityv1 "api-gateway/internal/grpc/identity/v1"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"secret123"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token" binding:"required" example:"string"`
	RefreshToken string `json:"refresh_token" binding:"required" example:"string"`
	UserId       string `json:"user_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"Failed to login"`
}

func respondWithError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

// @Summary      Login
// @Description  login the user
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Email"
// @Success      200  {object}  LoginResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /auth/login [post]
func LoginHandler(identityClient identityv1.AuthServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()

		var req LoginRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			respondWithError(w, "Invalid JSON data", http.StatusBadRequest)
			return
		}

		res, err := identityClient.Login(ctx, &identityv1.LoginRequest{
			Email:    req.Email,
			Password: req.Password,
		})
		if err != nil {
			respondWithError(w, "Failed to login", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}
