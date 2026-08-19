package services

import (
	"context"
	"errors"
	"fmt"

	"catalog/internal/db"
	"catalog/internal/errs"
	pb "catalog/internal/grpc/catalog/v1"
	"catalog/internal/types"
	"catalog/internal/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/grpc/codes"
)

//mockery:generate: true
type ProductRepository interface {
	CreateProduct(ctx context.Context, arg db.CreateProductParams) (db.Product, error)
	GetProduct(ctx context.Context, id pgtype.UUID) (db.Product, error)
	GetProductsByCategory(ctx context.Context, categoryID int32) ([]db.Product, error)
	GetProductsByCreator(ctx context.Context, creatorID pgtype.UUID) ([]db.Product, error)
}

type ProductService struct {
	repo ProductRepository
}

func NewProductService(repo ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	userID, ok := ctx.Value(types.UserIDKey).(*pgtype.UUID)
	if !ok {
		return nil, errs.Unauthenticated()
	}

	price, appErr := utils.StringToNumeric(req.Price)
	if appErr != nil {
		return nil, appErr
	}

	product, err := s.repo.CreateProduct(ctx, db.CreateProductParams{
		Name:        req.Name,
		Description: req.Description,
		Price:       price,
		Attributes:  []byte(req.Attributes),
		CreatorID:   *userID,
		CategoryID:  req.CategoryId,
	})
	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to create product: %w", err))
	}

	priceStr, appErr := utils.NumericToString(product.Price)
	if appErr != nil {
		return nil, appErr
	}

	return &pb.CreateProductResponse{
		Product: &pb.Product{
			Name:        product.Name,
			Description: product.Description,
			Price:       priceStr,
			Attributes:  string(product.Attributes),
			CreatorId:   product.CreatorID.String(),
			CategoryId:  product.CategoryID,
			Id:          product.ID.String(),
			Status:      utils.ProductStatusFromDBToGRPC(product.Status),
			UpdatedAt:   utils.TimestamptzToGRPCTimestamp(product.UpdatedAt),
			CreatedAt:   utils.TimestamptzToGRPCTimestamp(product.CreatedAt),
		},
	}, nil
}

func (s *ProductService) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	productUUID, appErr := utils.StringToUUID(req.Id)
	if appErr != nil {
		return nil, appErr
	}

	product, err := s.repo.GetProduct(ctx, productUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &errs.AppError{Code: codes.NotFound}
		}

		return nil, errs.Internal(fmt.Errorf("failed to get product: %w", err))
	}

	priceStr, appErr := utils.NumericToString(product.Price)
	if appErr != nil {
		return nil, appErr
	}

	return &pb.GetProductResponse{
		Product: &pb.Product{
			Name:        product.Name,
			Description: product.Description,
			Price:       priceStr,
			Attributes:  string(product.Attributes),
			CreatorId:   product.CreatorID.String(),
			CategoryId:  product.CategoryID,
			Id:          product.ID.String(),
			Status:      utils.ProductStatusFromDBToGRPC(product.Status),
			UpdatedAt:   utils.TimestamptzToGRPCTimestamp(product.UpdatedAt),
			CreatedAt:   utils.TimestamptzToGRPCTimestamp(product.CreatedAt),
		},
	}, nil
}

func (s *ProductService) GetProductsByCategory(ctx context.Context, req *pb.GetProductsByCategoryRequest) (*pb.GetProductsByCategoryResponse, error) {
	products, err := s.repo.GetProductsByCategory(ctx, req.CategoryId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &errs.AppError{Code: codes.NotFound}
		}

		return nil, errs.Internal(fmt.Errorf("failed to get products by category: %w", err))
	}

	var productsResponse []*pb.Product = make([]*pb.Product, 0, len(products))
	for _, product := range products {
		priceStr, appErr := utils.NumericToString(product.Price)
		if appErr != nil {
			return nil, appErr
		}

		productsResponse = append(productsResponse, &pb.Product{
			Name:        product.Name,
			Description: product.Description,
			Price:       priceStr,
			Attributes:  string(product.Attributes),
			CreatorId:   product.CreatorID.String(),
			CategoryId:  product.CategoryID,
			Id:          product.ID.String(),
			Status:      utils.ProductStatusFromDBToGRPC(product.Status),
			UpdatedAt:   utils.TimestamptzToGRPCTimestamp(product.UpdatedAt),
			CreatedAt:   utils.TimestamptzToGRPCTimestamp(product.CreatedAt),
		})
	}

	return &pb.GetProductsByCategoryResponse{
		Products: productsResponse,
	}, nil
}

func (s *ProductService) GetProductsByCreator(ctx context.Context, req *pb.GetProductsByCreatorRequest) (*pb.GetProductsByCreatorResponse, error) {
	creatorUUID, appErr := utils.StringToUUID(req.CreatorId)
	if appErr != nil {
		return nil, appErr
	}

	products, err := s.repo.GetProductsByCreator(ctx, creatorUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &errs.AppError{Code: codes.NotFound, Msg: "products are not found"}
		}

		return nil, errs.Internal(fmt.Errorf("failed to get products by category: %w", err))
	}

	var productsResponse []*pb.Product = make([]*pb.Product, 0, len(products))
	for _, product := range products {
		priceStr, appErr := utils.NumericToString(product.Price)
		if appErr != nil {
			return nil, appErr
		}

		productsResponse = append(productsResponse, &pb.Product{
			Name:        product.Name,
			Description: product.Description,
			Price:       priceStr,
			Attributes:  string(product.Attributes),
			CreatorId:   product.CreatorID.String(),
			CategoryId:  product.CategoryID,
			Id:          product.ID.String(),
			Status:      utils.ProductStatusFromDBToGRPC(product.Status),
			UpdatedAt:   utils.TimestamptzToGRPCTimestamp(product.UpdatedAt),
			CreatedAt:   utils.TimestamptzToGRPCTimestamp(product.CreatedAt),
		})
	}

	return &pb.GetProductsByCreatorResponse{
		Products: productsResponse,
	}, nil
}
