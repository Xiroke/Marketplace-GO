package services

import (
	"catalog/internal/db"
	catalogv1 "catalog/internal/grpc/v1"
	"catalog/internal/services/mocks"
	"catalog/internal/utils"
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestProductService_CreateProduct(t *testing.T) {
    repositoryMock := mocks.NewMockProductRepository(t)

    price, appErr := utils.StringToNumeric("100.56")
    require.Nil(t, appErr)

    attributes := map[string]string{"Color": "white", "Length": "34m"}
    attributesJSON, err := json.Marshal(attributes)
    require.NoError(t, err)

    repositoryMock.
        EXPECT().
        CreateProduct(mock.Anything, mock.AnythingOfType("db.CreateProductParams")).
        Return(db.Product{
            Price: price,
            Attributes: attributesJSON,
        }, nil)

    service := NewProductService(repositoryMock)
    res, err := service.CreateProduct(context.Background(), &catalogv1.CreateProductRequest{
        Name: "Test product",
        Description: "Test description",
        Price: "100.56",
        Attributes: string(attributesJSON),
        CreatorId: uuid.New().String(),
        CategoryId: 2,
    })

    require.NoError(t, err)
    require.Equal(t, res.Product.Price, "100.56")
    require.NotEmpty(t, res.Product.Attributes)
    parsedAttributes := map[string]string{}
    err = json.Unmarshal([]byte(res.Product.Attributes), &parsedAttributes)
    require.NoError(t, err)
    require.Equal(t, parsedAttributes["Color"], "white")
}

func TestProductService_GetProduct(t *testing.T) {
    repositoryMock := mocks.NewMockProductRepository(t)

    price, appErr := utils.StringToNumeric("100.56")
    require.Nil(t, appErr)

    attributes := map[string]string{"Color": "white", "Length": "34m"}
    attributesJSON, err := json.Marshal(attributes)
    require.NoError(t, err)

    repositoryMock.
        EXPECT().
        GetProduct(mock.Anything, mock.AnythingOfType("pgtype.UUID")).
        Return(db.Product{
            Price: price,
            Attributes: attributesJSON,
        }, nil)

    service := NewProductService(repositoryMock)
    res, err := service.GetProduct(context.Background(), &catalogv1.GetProductRequest{Id: uuid.New().String()})

    require.NoError(t, err)
    require.Equal(t, res.Product.Price, "100.56")
    require.NotEmpty(t, res.Product.Attributes)
    parsedAttributes := map[string]string{}
    err = json.Unmarshal([]byte(res.Product.Attributes), &parsedAttributes)
    require.NoError(t, err)
    require.Equal(t, parsedAttributes["Color"], "white")
}

func TestProductService_GetProductsByCategory(t *testing.T) {
    repositoryMock := mocks.NewMockProductRepository(t)

    price, appErr := utils.StringToNumeric("100.56")
    require.Nil(t, appErr)

    attributes := map[string]string{"Color": "white", "Length": "34m"}
    attributesJSON, err := json.Marshal(attributes)
    require.NoError(t, err)

    repositoryMock.
        EXPECT().
        GetProductsByCategory(mock.Anything, mock.AnythingOfType("int32")).
        Return([]db.Product{
            {
                Price: price,
                Attributes: attributesJSON,
            },
        }, nil)

    service := NewProductService(repositoryMock)
    res, err := service.GetProductsByCategory(context.Background(), &catalogv1.GetProductsByCategoryRequest{CategoryId: 1})

    require.NoError(t, err)
    require.Equal(t, res.Products[0].Price, "100.56")
    require.NotEmpty(t, res.Products[0].Attributes)
    parsedAttributes := map[string]string{}
    err = json.Unmarshal([]byte(res.Products[0].Attributes), &parsedAttributes)
    require.NoError(t, err)
    require.Equal(t, parsedAttributes["Color"], "white")
}

func TestProductService_GetProductsByCreator(t *testing.T) {
    repositoryMock := mocks.NewMockProductRepository(t)

    price, appErr := utils.StringToNumeric("100.56")
    require.Nil(t, appErr)

    attributes := map[string]string{"Color": "white", "Length": "34m"}
    attributesJSON, err := json.Marshal(attributes)
    require.NoError(t, err)

    repositoryMock.
        EXPECT().
        GetProductsByCreator(mock.Anything, mock.AnythingOfType("pgtype.UUID")).
        Return([]db.Product{
            {
                Price: price,
                Attributes: attributesJSON,
            },
        }, nil)

    service := NewProductService(repositoryMock)
    res, err := service.GetProductsByCreator(context.Background(), &catalogv1.GetProductsByCreatorRequest{CreatorId: uuid.New().String()})

    require.NoError(t, err)
    require.Equal(t, res.Products[0].Price, "100.56")
    require.NotEmpty(t, res.Products[0].Attributes)
    parsedAttributes := map[string]string{}
    err = json.Unmarshal([]byte(res.Products[0].Attributes), &parsedAttributes)
    require.NoError(t, err)
    require.Equal(t, parsedAttributes["Color"], "white")
}
