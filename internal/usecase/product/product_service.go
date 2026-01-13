package product

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"kakaidee-backend/internal/models"
	productErr "kakaidee-backend/internal/usecase/error/product"
)

type ProductService struct {
	db *mongo.Database
}

func NewProductService(db *mongo.Database) *ProductService {
	return &ProductService{db: db}
}

func (s *ProductService) Create(ctx context.Context, product models.Product) error {
	if product.Barcode == "" {
		return productErr.ErrInvalidProduct
	}

	product.Status = "active"
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	return nil
}
