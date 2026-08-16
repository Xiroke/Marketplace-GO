package server

import (
	pb "catalog/internal/grpc/catalog/v1"
	"catalog/internal/services"
	"catalog/internal/utils"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

var _ pb.CatalogServiceServer = (*server)(nil)

type server struct {
	pb.UnsafeCatalogServiceServer
	db *pgxpool.Pool
    productService *services.ProductService
    categoryService *services.CategoryService
}

func NewServer(
        db *pgxpool.Pool,
        productService *services.ProductService,
        categoryService *services.CategoryService,
    ) *server {
	return &server{db: db, productService: productService, categoryService: categoryService}
}

func (s *server) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.CreateCategoryResponse, error) {
	if err := utils.CloseCanceledRequestByContext(ctx); err != nil {
		return nil, err
	}

    res, err := s.categoryService.CreateCategory(ctx, req)
    if err != nil {
        return nil, err
    }

    return res, nil
}

func (s *server) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	if err := utils.CloseCanceledRequestByContext(ctx); err != nil {
		return nil, err
	}

    res, err := s.productService.CreateProduct(ctx, req)
    if err != nil {
        return nil, err
    }

    return res, nil
}

func (s *server) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	if err := utils.CloseCanceledRequestByContext(ctx); err != nil {
		return nil, err
	}

    res, err := s.productService.GetProduct(ctx, req)
    if err != nil {
        return nil, err
    }

    return res, nil
}

func (s *server) GetProductsByCategory(ctx context.Context, req *pb.GetProductsByCategoryRequest) (*pb.GetProductsByCategoryResponse, error) {
	if err := utils.CloseCanceledRequestByContext(ctx); err != nil {
		return nil, err
	}

    res, err := s.productService.GetProductsByCategory(ctx, req)
    if err != nil {
        return nil, err
    }

    return res, nil
}

func (s *server) GetProductsByCreator(ctx context.Context, req *pb.GetProductsByCreatorRequest) (*pb.GetProductsByCreatorResponse, error) {
	if err := utils.CloseCanceledRequestByContext(ctx); err != nil {
		return nil, err
	}

    res, err := s.productService.GetProductsByCreator(ctx, req)
    if err != nil {
        return nil, err
    }

    return res, nil
}

