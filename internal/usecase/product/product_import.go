package product

import (
	"context"
	"kakaidee-backend/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *ProductService) ImportFromExcel(
	ctx context.Context,
	reqs []payloadProduct.ImportProductRequest,
	createdBy string,
) (map[string]interface{}, error) {

	now := time.Now()

	if s.db == nil {
		return nil, errors.New("database not initialized")
	}

	var (
		success int
		failed  []string
	)

	for i, req := range reqs {

		startTime := time.Now()

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
			CreatedBy:     createdBy,
			UpdatedBy:     createdBy,
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
			CreatedBy:          createdBy,
			UpdatedBy:          createdBy,
			CreatedAt:          now,
			UpdatedAt:          now,
		}

		// ===== insert product_stock =====
		if _, err := s.db.
			Collection(models.ProductStock{}.CollectionName()).
			InsertOne(ctx, productStock); err != nil {

			failed = append(failed,
				fmt.Sprintf("row %d: insert product_stock failed", i+1),
			)
			continue
		}

		logStock := models.TransactionLog{
			FunctionEndpoint:   "product/import/excel",
			FunctionMethod:     "POST",
			FunctionName:       "ImportProductExcel",
			FunctionController: "Product",
			Environment:        "local",
			QueryCollection:    "product_stock",
			QueryType:          "insert",
			StartTime:          startTime,
			EndTime:            time.Now(),
			DurationMs:         time.Since(startTime).Milliseconds(),
			CountData:          1,
			StatusCode:         201,
			StatusMessage:      "created",
			CreatedBy:          createdBy,
			CreatedAt:          time.Now(),
		}

		_, _ = s.db.
			Collection(logStock.CollectionName()).
			InsertOne(ctx, logStock)

		// ===== upsert product master =====
		filter := bson.M{
			"sku_code": req.SKUCode,
		}

		update := bson.M{
			"$set": product,
			"$inc": bson.M{
				"balance_qty": req.ReceiveQty,
			},
		}

		_, err := s.db.
			Collection(models.Product{}.CollectionName()).
			UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))

		if err != nil {
			failed = append(failed,
				fmt.Sprintf("row %d: upsert product failed", i+1),
			)
			continue
		}

		logProduct := models.TransactionLog{
			FunctionEndpoint:   "product/import/excel",
			FunctionMethod:     "POST",
			FunctionName:       "ImportProductExcel",
			FunctionController: "Product",
			Environment:        "local",
			QueryCollection:    "product_master",
			QueryType:          "upsert",
			StartTime:          startTime,
			EndTime:            time.Now(),
			DurationMs:         time.Since(startTime).Milliseconds(),
			CountData:          1,
			StatusCode:         201,
			StatusMessage:      "created",
			CreatedBy:          createdBy,
			CreatedAt:          time.Now(),
		}

		_, _ = s.db.
			Collection(logProduct.CollectionName()).
			InsertOne(ctx, logProduct)

		success++
	}

	return map[string]interface{}{
		"imported": success,
		"failed":   failed,
	}, nil
}

