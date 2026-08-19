package server

import (
	"context"

	pb "catalog/internal/grpc/catalog/v1"
	"catalog/internal/services"

	"github.com/jackc/pgx/v5/pgxpool"
)

var _ pb.CatalogServiceServer = (*server)(nil)

type server struct {
	pb.UnsafeCatalogServiceServer
	db              *pgxpool.Pool
	productService  *services.ProductService
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
	return s.categoryService.CreateCategory(ctx, req)
}

func (s *server) GetCategories(ctx context.Context, req *pb.GetCategoriesRequest) (*pb.GetCategoriesResponse, error) {
	return s.categoryService.GetCategories(ctx, req)
}

func (s *server) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	return s.productService.CreateProduct(ctx, req)
}

func (s *server) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	return s.productService.GetProduct(ctx, req)
}

func (s *server) GetProductsByCategory(ctx context.Context, req *pb.GetProductsByCategoryRequest) (*pb.GetProductsByCategoryResponse, error) {
	return s.productService.GetProductsByCategory(ctx, req)
}

func (s *server) GetProductsByCreator(ctx context.Context, req *pb.GetProductsByCreatorRequest) (*pb.GetProductsByCreatorResponse, error) {
	return s.productService.GetProductsByCreator(ctx, req)
}
