package services

import (
	"catalog/internal/db"
	"catalog/internal/errs"
	pb "catalog/internal/grpc/catalog/v1"
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

//mockery:generate: true
type CategoryRepository interface {
	CreateCategory(ctx context.Context, arg db.CreateCategoryParams) (db.Category, error)
}

type CategoryService struct {
	repo   CategoryRepository
}

func NewCategoryService(repo CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.CreateCategoryResponse, error) {
	parentId := pgtype.Int4{Int32: req.ParentId, Valid: true}

	category, err := s.repo.CreateCategory(ctx, db.CreateCategoryParams{
		Name:     req.Name,
		ParentID: parentId,
	})
	if err != nil {
		return nil, errs.Internal("failed to create category", err)
	}

	return &pb.CreateCategoryResponse{
		Name:      category.Name,
		ParentId:  category.ParentID.Int32,
		Id:        category.ID,
		CreatedAt: timestamppb.New(category.CreatedAt.Time),
	}, nil
}

