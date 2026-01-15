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
	now := time.Now()
	if s.db == nil {
		return errors.New("database not initialized")
	}
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

	logProductStock := models.TransactionLog{
		RequestID:          "",
		FunctionEndpoint:   "product/create",
		FunctionMethod:     "POST",
		FunctionName:       "CreateProduct",
		FunctionController: "ProductService",
		Environment:        "local",
		QueryCollection:    "product_stock",
		QueryType:          "insert",
		StartTime:          now,
		EndTime:            time.Now(),
		DurationMs:         time.Since(now).Milliseconds(),
		CountData:          1,
		StatusCode:         201,
		StatusMessage:      "created",
		CreatedBy:          req.CreatedBy,
		CreatedAt:          time.Now(),
	}

	if _, err := s.db.
		Collection(models.Product{}.CollectionName()).
		InsertOne(ctx, product); err != nil {
		return err
	}

	_, _ = s.db.Collection(logProductStock.CollectionName()).InsertOne(ctx, logProductStock)

	logProduct := models.TransactionLog{
		RequestID:          "",
		FunctionEndpoint:   "product/create",
		FunctionMethod:     "POST",
		FunctionName:       "CreateProduct",
		FunctionController: "ProductService",
		Environment:        "local",
		QueryCollection:    "product",
		QueryType:          "insert",
		StartTime:          now,
		EndTime:            time.Now(),
		DurationMs:         time.Since(now).Milliseconds(),
		CountData:          1,
		StatusCode:         201,
		StatusMessage:      "created",
		CreatedBy:          req.CreatedBy,
		CreatedAt:          time.Now(),
	}

	_, _ = s.db.Collection(logProduct.CollectionName()).InsertOne(ctx, logProduct)

	return nil
}
