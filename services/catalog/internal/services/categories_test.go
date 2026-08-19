package services

import (
	"context"
	"testing"
	"time"

	"catalog/internal/db"
	catalogv1 "catalog/internal/grpc/catalog/v1"
	"catalog/internal/services/mocks"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCategoryService_CreateCategory(t *testing.T) {
	repositoryMock := mocks.NewMockCategoryRepository(t)

	repositoryMock.
		EXPECT().
		CreateCategory(mock.Anything, mock.AnythingOfType("db.CreateCategoryParams")).
		Return(db.Category{
			ID:        2,
			Name:      "Test category",
			ParentID:  pgtype.Int4{Int32: 1, Valid: true},
			CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}, nil)

	service := NewCategoryService(repositoryMock)
	var parentID int32 = 1
	res, err := service.CreateCategory(context.Background(), &catalogv1.CreateCategoryRequest{
		Name:     "Test category",
		ParentId: &parentID,
	})

	require.NoError(t, err)
	require.Equal(t, res.Category.Name, "Test category")
}
