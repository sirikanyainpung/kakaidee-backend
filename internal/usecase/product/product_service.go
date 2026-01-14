package product

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"kakaidee-backend/internal/models"
	payloadProduct "kakaidee-backend/internal/payload/product"
)

type ProductService struct {
	db *mongo.Database
}

func NewProductService(db *mongo.Database) *ProductService {
	return &ProductService{db: db}
}

func (s *ProductService) Create(
	ctx context.Context,
	req payloadProduct.CreateProductRequest,
) error {

	if s.db == nil {
		return errors.New("database not initialized")
	}

	now := time.Now()

	productStock := models.ProductStock{
		Barcode:       req.Barcode,
		SKUCode:       req.SKUCode,
		LotNo:         req.LotNo,
		WarehouseName: req.WarehouseName,
		WarehouseZone: req.WarehouseZone,
		Bin:           req.Bin,
		StockType:     req.StockType,
		ReceiveQty:    req.ReceiveQty,
		BalanceQty:    req.ReceiveQty,
		SellingQty:    0,
		MFG:           req.MFG,
		EXP:           req.EXP,
		Status:        "active",
		CreatedBy:     req.CreatedBy,
		UpdatedBy:     req.UpdatedBy,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	product := models.Product{
		Barcode:            req.Barcode,
		SKUCode:            req.SKUCode,
		ProductName:        req.ProductName,
		ProductDescription: req.ProductDescription,
		CategoryCode:       req.CategoryCode,
		SupplierCode:       req.SupplierCode,
		BrandCode:          req.BrandCode,
		BalanceQty:         req.ReceiveQty,
		Unit:               req.Unit,
		CostPrice:          req.CostPrice,
		Status:             "active",
		CreatedBy:          req.CreatedBy,
		UpdatedBy:          req.UpdatedBy,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if _, err := s.db.
		Collection(models.ProductStock{}.CollectionName()).
		InsertOne(ctx, productStock); err != nil {
		return err
	}

	if _, err := s.db.
		Collection(models.Product{}.CollectionName()).
		InsertOne(ctx, product); err != nil {
		return err
	}

	return nil
}

// func (s *ProductService) Create(ctx context.Context, product models.Product) error {
// 	if product.Barcode == "" {
// 		return productErr.ErrInvalidProduct
// 	}

// 	product.Status = "active"
// 	product.CreatedAt = time.Now()
// 	product.UpdatedAt = time.Now()

// 	return nil
// }
