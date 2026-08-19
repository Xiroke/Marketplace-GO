package categories

import (
	"api-gateway/internal/api/base"
	catalogv1 "api-gateway/internal/grpc/catalog/v1"
	identityv1 "api-gateway/internal/grpc/identity/v1"
	"api-gateway/internal/utils"
	_ "api-gateway/internal/utils/responses"
	"context"
	"log/slog"
	"net/http"
	"time"
)

type Category struct {
	Id        int32     `json:"id" example:"1"`
	Name      string    `json:"name" example:"string"`
	ParentId  *int32    `json:"parent_id" example:"1"`
	CreatedAt time.Time `json:"created_at" example:"2016-08-19T15:16:00Z"`
}

type CreateCategoryRequest struct {
	Name     string `json:"name" binding:"required" example:"string"`
	ParentId int32  `json:"parent_id" example:"1"`
}

type CreateCategoryResponse Category

// @Summary      create category
// @Description  create category in marketplace
// @Tags         catalog
// @Accept       json
// @Produce      json
// @Param        request body CreateCategoryRequest true "body"
// @Success      200  {object}  CreateCategoryResponse
// @Failure      500  {object}  responses.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /auth/categories [post]
func CreateCategoryHandler(logger *slog.Logger, identityClient identityv1.AuthServiceClient, catalogClient catalogv1.CatalogServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		base.HandleRequest(w, r, logger, func(ctx context.Context, req *CreateCategoryRequest) (*CreateCategoryResponse, error) {
			ctx, err := utils.AddAuthorizationToCTX(ctx, logger)
			if err != nil {
				return nil, err
			}

			data, err := catalogClient.CreateCategory(ctx, &catalogv1.CreateCategoryRequest{
				Name:     req.Name,
				ParentId: &req.ParentId,
			})
			if err != nil {
				return nil, err
			}
			category := data.Category

			return &CreateCategoryResponse{
				Id:        category.Id,
				Name:      category.Name,
				ParentId:  category.ParentId,
				CreatedAt: category.CreatedAt.AsTime(),
			}, nil
		})
	}
}

type GetCategoriesResponse struct {
	Data []Category `json:"data"`
}

// @Summary      get categories
// @Description  get categories
// @Tags         catalog
// @Accept       json
// @Produce      json
// @Success      200  {object}  GetCategoriesResponse
// @Failure      500  {object}  responses.ErrorResponse
// @Router       /auth/categories [get]
func GetCategoriesHandler(logger *slog.Logger, identityClient identityv1.AuthServiceClient, catalogClient catalogv1.CatalogServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		base.HandleRequest(w, r, logger, func(ctx context.Context, req *struct{}) (*GetCategoriesResponse, error) {
			data, err := catalogClient.GetCategories(ctx, &catalogv1.GetCategoriesRequest{})
			if err != nil {
				return nil, err
			}

			categoryResponse := make([]Category, 0, len(data.Categories))
			for _, i := range data.Categories {
				categoryResponse = append(categoryResponse, Category{
					Id:        i.Id,
					Name:      i.Name,
					ParentId:  i.ParentId,
					CreatedAt: i.CreatedAt.AsTime(),
				})
			}

			return &GetCategoriesResponse{
				Data: categoryResponse,
			}, nil
		})
	}
}
