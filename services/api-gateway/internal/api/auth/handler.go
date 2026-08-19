package auth

import (
	"api-gateway/internal/api/base"
	identityv1 "api-gateway/internal/grpc/identity/v1"
	"api-gateway/internal/utils"
	_ "api-gateway/internal/utils/responses"
	"context"
	"log/slog"
	"net/http"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"secret123"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token" example:"string"`
	RefreshToken string `json:"refresh_token" example:"string"`
	UserId       string `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// @Summary      Login
// @Description  login the user
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "body"
// @Success      200  {object}  LoginResponse
// @Failure      401  {object}  responses.ErrorResponse
// @Failure      404  {object}  responses.ErrorResponse
// @Failure      500  {object}  responses.ErrorResponse
// @Router       /auth/login [post]
func LoginHandler(logger *slog.Logger, identityClient identityv1.AuthServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		base.HandleRequest(w, r, logger, func(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
			data, err := identityClient.Login(ctx, &identityv1.LoginRequest{
				Email:    req.Email,
				Password: req.Password,
			})

			if err != nil {
				return nil, err
			}

			return &LoginResponse{
				AccessToken:  data.AccessToken,
				RefreshToken: data.RefreshToken,
				UserId:       data.UserId,
			}, nil
		})
	}
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"secret123"`
	Username string `json:"username" binding:"required" example:"user"`
}

type RegisterResponse struct {
	AccessToken  string `json:"access_token" example:"string"`
	RefreshToken string `json:"refresh_token" example:"string"`
	UserId       string `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// @Summary      register
// @Description  create the user account
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "body"
// @Success      200  {object}  RegisterResponse
// @Failure      401  {object}  responses.ErrorResponse
// @Failure      500  {object}  responses.ErrorResponse
// @Router       /auth/register [post]
func RegisterHandler(logger *slog.Logger, identityClient identityv1.AuthServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		base.HandleRequest(w, r, logger, func(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
			ctx, err := utils.AddAuthorizationToCTX(ctx, logger)
			if err != nil {
				return nil, err
			}

			data, err := identityClient.Register(ctx, &identityv1.RegisterRequest{
				Email:    req.Email,
				Password: req.Password,
				Username: req.Username,
			})
			if err != nil {
				return nil, err
			}

			return &RegisterResponse{
				AccessToken:  data.AccessToken,
				RefreshToken: data.RefreshToken,
				UserId:       data.UserId,
			}, nil
		})
	}
}

type GetMeResponse struct {
	Id       string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Username string `json:"username" example:"user"`
	Email    string `json:"email" example:"user@example.com"`
}

// @Summary      get me
// @Description  get me
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Failure      401  {object}  responses.ErrorResponse
// @Failure      500  {object}  responses.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /auth/me [post]
func GetMeHandler(logger *slog.Logger, identityClient identityv1.AuthServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		base.HandleRequest(w, r, logger, func(ctx context.Context, req *struct{}) (*GetMeResponse, error) {
			data, err := identityClient.GetMe(ctx, &identityv1.GetMeRequest{})
			if err != nil {
				return nil, err
			}

			user := data.User

			return &GetMeResponse{
				Id:       user.Id,
				Username: user.Username,
				Email:    user.Email,
			}, nil
		})
	}
}
