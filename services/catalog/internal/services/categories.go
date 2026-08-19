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
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

//mockery:generate: true
type CategoryRepository interface {
	CreateCategory(ctx context.Context, arg db.CreateCategoryParams) (db.Category, error)
	GetCategories(ctx context.Context) ([]db.Category, error)
}

type CategoryService struct {
	repo CategoryRepository
}

func NewCategoryService(repo CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.CreateCategoryResponse, error) {
	_, ok := ctx.Value(types.UserIDKey).(*pgtype.UUID)
	if !ok {
		return nil, errs.Internal(errors.New("failed to get user from ctx"))
	}

	var parentID pgtype.Int4

	if req.ParentId == nil || *req.ParentId == 0 {
		parentID = pgtype.Int4{Int32: 0, Valid: false}
	} else {
		parentID = pgtype.Int4{Int32: *req.ParentId, Valid: true}
	}

	category, err := s.repo.CreateCategory(ctx, db.CreateCategoryParams{
		Name:     req.Name,
		ParentID: parentID,
	})
	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to create category: %w", err))
	}

	var parentIDPtr *int32
	if category.ParentID.Valid {
		val := category.ParentID.Int32
		parentIDPtr = &val
	}

	return &pb.CreateCategoryResponse{
		Category: &pb.Category{
			Name:      category.Name,
			ParentId:  parentIDPtr,
			Id:        category.ID,
			CreatedAt: timestamppb.New(category.CreatedAt.Time),
		},
	}, nil
}

func (s *CategoryService) GetCategories(ctx context.Context, req *pb.GetCategoriesRequest) (*pb.GetCategoriesResponse, error) {
	categories, err := s.repo.GetCategories(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &pb.GetCategoriesResponse{
				Categories: make([]*pb.Category, 0),
			}, nil
		}

		return nil, errs.Internal(fmt.Errorf("failled to get categories: %w", err))
	}

	categoryResponse := make([]*pb.Category, 0, len(categories))
	for _, i := range categories {
		var parentIDPtr *int32
		if i.ParentID.Valid {
			val := i.ParentID.Int32
			parentIDPtr = &val
		}

		categoryResponse = append(categoryResponse, &pb.Category{
			Name:      i.Name,
			ParentId:  parentIDPtr,
			Id:        i.ID,
			CreatedAt: utils.TimestamptzToGRPCTimestamp(i.CreatedAt),
		})
	}

	return &pb.GetCategoriesResponse{
		Categories: categoryResponse,
	}, nil
}
