package products

import (
	"api-gateway/internal/api/base"
	catalogv1 "api-gateway/internal/grpc/catalog/v1"
	identityv1 "api-gateway/internal/grpc/identity/v1"
	"api-gateway/internal/utils"
	_ "api-gateway/internal/utils/responses"
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Product struct {
	Id          string    `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name        string    `json:"name" example:"string"`
	Description string    `json:"description" example:"string"`
	Price       string    `json:"price" example:"123.45"`
	Attributes  string    `json:"attributes" example:"{'color': 'red'}"`
	CreatorId   string    `json:"creator_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	CategoryId  int32     `json:"category_id" example:"1"`
	Status      string    `json:"status" example:"draft"`
	UpdatedAt   time.Time `json:"updated_at" example:"2016-08-19T15:16:00Z"`
	CreatedAt   time.Time `json:"created_at" example:"2016-08-19T15:16:00Z"`
}

type CreateProductRequest struct {
	Name        string `json:"name" binding:"required" example:"string"`
	Description string `json:"description" binding:"required" example:"string"`
	Price       string `json:"price" binding:"required" example:"123.45"`
	Attributes  string `json:"attributes" binding:"required" example:"{'color': 'red'}"`
	CategoryId  int32  `json:"category_id" binding:"required" example:"1"`
}

type CreateProductResponse Product

// @Summary      create product
// @Description  create product in marketplace
// @Tags         catalog
// @Accept       json
// @Produce      json
// @Param        request body CreateProductRequest true "body"
// @Success      200  {object}  CreateProductResponse
// @Failure      500  {object}  responses.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /auth/products [post]
func CreateProductHandler(logger *slog.Logger, identityClient identityv1.AuthServiceClient, catalogClient catalogv1.CatalogServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		base.HandleRequest(w, r, logger, func(ctx context.Context, req *CreateProductRequest) (*CreateProductResponse, error) {
			ctx, err := utils.AddAuthorizationToCTX(ctx, logger)
			if err != nil {
				return nil, err
			}

			data, err := catalogClient.CreateProduct(ctx, &catalogv1.CreateProductRequest{
				Name:        req.Name,
				Description: req.Description,
				Price:       req.Price,
				Attributes:  req.Attributes,
				CategoryId:  req.CategoryId,
			})
			if err != nil {
				return nil, err
			}
			product := data.Product

			profuctStatus, err := utils.ProductStatusFromGRPCToString(product.Status)
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, "invalid product status")
			}

			return &CreateProductResponse{
				Name:        product.Name,
				Description: product.Description,
				Price:       product.Price,
				Attributes:  product.Attributes,
				CreatorId:   product.CreatorId,
				CategoryId:  product.CategoryId,
				Id:          product.Id,
				Status:      profuctStatus,
				UpdatedAt:   product.UpdatedAt.AsTime(),
				CreatedAt:   product.CreatedAt.AsTime(),
			}, nil
		})
	}
}

type GetProductsByCreatorResponse struct {
	Data []Product `json:"data"`
}

// @Summary      get products by creator
// @Description  get products by creator
// @Tags         catalog
// @Accept       json
// @Produce      json
// @Param        creator_id path string true "user id as uuid"
// @Success      200  {object}  GetProductsByCreatorResponse
// @Failure      500  {object}  responses.ErrorResponse
// @Router       /auth/products/creator/{creator_id} [get]
func GetProductsByCreatorHandler(logger *slog.Logger, identityClient identityv1.AuthServiceClient, catalogClient catalogv1.CatalogServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		base.HandleRequest(w, r, logger, func(ctx context.Context, req *struct{}) (*GetProductsByCreatorResponse, error) {
			data, err := catalogClient.GetProductsByCreator(ctx, &catalogv1.GetProductsByCreatorRequest{
				CreatorId: chi.URLParam(r, "creator_id"),
			})
			if err != nil {
				return nil, err
			}

			itemsResponse := make([]Product, 0, len(data.Products))
			for _, i := range data.Products {
				profuctStatus, err := utils.ProductStatusFromGRPCToString(i.Status)
				if err != nil {
					return nil, status.Error(codes.InvalidArgument, "invalid product status")
				}

				itemsResponse = append(itemsResponse, Product{
					Id:          i.Id,
					Name:        i.Name,
					Description: i.Description,
					Price:       i.Price,
					Attributes:  i.Attributes,
					CreatorId:   i.CreatorId,
					CategoryId:  i.CategoryId,
					Status:      profuctStatus,
					UpdatedAt:   i.UpdatedAt.AsTime(),
					CreatedAt:   i.CreatedAt.AsTime(),
				})
			}

			return &GetProductsByCreatorResponse{
				Data: itemsResponse,
			}, nil
		})
	}
}

type GetProductsByCategoryResponse struct {
	Data []Product `json:"data"`
}

// @Summary      get products by categpry
// @Description  get products by categpry
// @Tags         catalog
// @Accept       json
// @Produce      json
// @Param        category_id path int32 true "category id"
// @Success      200  {object}  GetProductsByCategoryResponse
// @Failure      500  {object}  responses.ErrorResponse
// @Router       /auth/products/category/{category_id} [get]
func GetProductsByCategoryHandler(logger *slog.Logger, identityClient identityv1.AuthServiceClient, catalogClient catalogv1.CatalogServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parsedCategoryID, err := strconv.ParseInt(chi.URLParam(r, "category_id"), 10, 32)
		if err != nil {
			http.Error(w, "invalid category_id", http.StatusBadRequest)
			return
		}
		categoryID := int32(parsedCategoryID)

		base.HandleRequest(w, r, logger, func(ctx context.Context, req *struct{}) (*GetProductsByCategoryResponse, error) {
			data, err := catalogClient.GetProductsByCategory(ctx, &catalogv1.GetProductsByCategoryRequest{
				CategoryId: categoryID,
			})
			if err != nil {
				return nil, err
			}

			itemsResponse := make([]Product, 0, len(data.Products))
			for _, i := range data.Products {
				profuctStatus, err := utils.ProductStatusFromGRPCToString(i.Status)
				if err != nil {
					return nil, status.Error(codes.InvalidArgument, "invalid product status")
				}

				itemsResponse = append(itemsResponse, Product{
					Id:          i.Id,
					Name:        i.Name,
					Description: i.Description,
					Price:       i.Price,
					Attributes:  i.Attributes,
					CreatorId:   i.CreatorId,
					CategoryId:  i.CategoryId,
					Status:      profuctStatus,
					UpdatedAt:   i.UpdatedAt.AsTime(),
					CreatedAt:   i.CreatedAt.AsTime(),
				})
			}

			return &GetProductsByCategoryResponse{
				Data: itemsResponse,
			}, nil
		})
	}
}
