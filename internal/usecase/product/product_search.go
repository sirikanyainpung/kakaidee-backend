package product

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"kakaidee-backend/internal/models"
)

func (s *ProductService) Get(
	ctx context.Context,
	keyword string,
	now time.Time,
) ([]bson.M, error) {

	// now := time.Now()

	pipeline := mongo.Pipeline{}

	// ===== 1. match (search) =====
	if keyword != "" {
		pipeline = append(pipeline, bson.D{
			{"$match", bson.M{
				"$or": []bson.M{
					{"barcode": bson.M{"$regex": keyword, "$options": "i"}},
					{"sku_code": bson.M{"$regex": keyword, "$options": "i"}},
					{"product_name": bson.M{"$regex": keyword, "$options": "i"}},
				},
			}},
		})
	}

	// ===== 2. lookup category =====
	pipeline = append(pipeline, bson.D{
		{"$lookup", bson.M{
			"from":         "category_masters",
			"localField":   "category_code",
			"foreignField": "category_code",
			"as":           "category",
		}},
	})

	// ===== 3. lookup supplier =====
	pipeline = append(pipeline, bson.D{
		{"$lookup", bson.M{
			"from":         "supplier_masters",
			"localField":   "supplier_code",
			"foreignField": "supplier_code",
			"as":           "supplier",
		}},
	})

	// ===== 4. lookup brand =====
	pipeline = append(pipeline, bson.D{
		{"$lookup", bson.M{
			"from":         "brand_masters",
			"localField":   "brand_code",
			"foreignField": "brand_code",
			"as":           "brand",
		}},
	})

	// ===== 5. unwind =====
	pipeline = append(pipeline,
		bson.D{{"$unwind", bson.M{"path": "$category", "preserveNullAndEmptyArrays": true}}},
		bson.D{{"$unwind", bson.M{"path": "$supplier", "preserveNullAndEmptyArrays": true}}},
		bson.D{{"$unwind", bson.M{"path": "$brand", "preserveNullAndEmptyArrays": true}}},
	)

	// ===== 6. project (flat response) =====
	pipeline = append(pipeline, bson.D{
		{"$project", bson.M{
			"_id": 0, // ⭐ ปิด _id

			"barcode":             1,
			"sku_code":            1,
			"product_name":        1,
			"product_description": 1,

			"category_code": 1,
			"category_name": "$category.category_name",

			"supplier_code": 1,
			"supplier_name": "$supplier.supplier_name",

			"brand_code": 1,
			"brand_name": "$brand.brand_name",

			"balance_qty": 1,
			"unit":        1,
			"cost_price":  1,
			"status":      1,
			"created_by":  1,
			"updated_by":  1,
			"created_at":  1,
			"updated_at":  1,
		}},
	})

	cur, err := s.db.
		Collection(models.Product{}.CollectionName()).
		Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var result []bson.M
	if err := cur.All(ctx, &result); err != nil {
		return nil, err
	}

	end := time.Now()
	logProduct := models.TransactionLog{
		RequestID:          "",
		FunctionEndpoint:   "product?keyword=" + keyword,
		FunctionMethod:     "GET",
		FunctionName:       "SearchProduct",
		FunctionController: "Product",
		Environment:        "local",
		QueryCollection:    "product_master",
		QueryType:          "query",
		StartTime:          now,
		EndTime:            end,
		DurationMs:         end.Sub(now).Milliseconds(),
		CountData:          len(result),
		StatusCode:         200,
		StatusMessage:      "success",
		CreatedBy:          "admin",
		CreatedAt:          now,
	}
	_, _ = s.db.Collection(logProduct.CollectionName()).InsertOne(ctx, logProduct)

	return result, nil
}

func (s *ProductService) GetV2(
	ctx context.Context,
	keyword string,
	skuCode string,
	categoryCode string,
	warehouseName string,
	lotsNo string,
	status string,
	now time.Time,
) ([]bson.M, error) {

	endpoint := "product/search-v2?keyword=" + keyword + "&sku_code=" + skuCode + "&category_code=" + categoryCode + "&warehouse_name=" + warehouseName + "&lot_no=" + lotsNo + "&status=" + status
	pipeline := mongo.Pipeline{}

	// ===== 1. match product_master =====
	match := bson.M{}

	if keyword != "" {
		match["$or"] = []bson.M{
			{"barcode": bson.M{"$regex": keyword, "$options": "i"}},
			{"product_name": bson.M{"$regex": keyword, "$options": "i"}},
		}
	}

	if skuCode != "" {
		match["sku_code"] = skuCode
	}

	if categoryCode != "" {
		match["category_code"] = categoryCode
	}

	if status != "" {
		match["status"] = status
	}

	if len(match) > 0 {
		pipeline = append(pipeline, bson.D{{"$match", match}})
	}

	// ===== 2. lookup product_stock =====
	pipeline = append(pipeline, bson.D{{"$lookup", bson.M{
		"from":         "product_stock",
		"localField":   "sku_code",
		"foreignField": "sku_code",
		"as":           "stock",
	}}})

	// ===== 3. unwind stock =====
	pipeline = append(pipeline, bson.D{{"$unwind", bson.M{
		"path":                       "$stock",
		"preserveNullAndEmptyArrays": true,
	}}})

	// ===== 4. filter stock =====
	stockMatch := bson.M{}

	if warehouseName != "" {
		stockMatch["stock.warehouses_name"] = warehouseName
	}

	if lotsNo != "" {
		stockMatch["stock.lots_no"] = lotsNo
	}

	if len(stockMatch) > 0 {
		pipeline = append(pipeline, bson.D{{"$match", stockMatch}})
	}

	// ===== 5. lookup category =====
	pipeline = append(pipeline, bson.D{{"$lookup", bson.M{
		"from":         "category_master",
		"localField":   "category_code",
		"foreignField": "category_code",
		"as":           "category",
	}}})

	// ===== 6. lookup supplier =====
	pipeline = append(pipeline, bson.D{{"$lookup", bson.M{
		"from":         "supplier_master",
		"localField":   "supplier_code",
		"foreignField": "supplier_code",
		"as":           "supplier",
	}}})

	// ===== 7. lookup brand =====
	pipeline = append(pipeline, bson.D{{"$lookup", bson.M{
		"from":         "brand_master",
		"localField":   "brand_code",
		"foreignField": "brand_code",
		"as":           "brand",
	}}})

	// ===== 8. unwind master =====
	pipeline = append(pipeline,
		bson.D{{"$unwind", bson.M{"path": "$category", "preserveNullAndEmptyArrays": true}}},
		bson.D{{"$unwind", bson.M{"path": "$supplier", "preserveNullAndEmptyArrays": true}}},
		bson.D{{"$unwind", bson.M{"path": "$brand", "preserveNullAndEmptyArrays": true}}},
	)

	// ===== 9. project =====
	pipeline = append(pipeline, bson.D{
		{"$project", bson.M{
			"_id": 0,

			"barcode":             1,
			"sku_code":            1,
			"product_name":        1,
			"product_description": 1,

			"category_code": 1,
			"category_name": "$category.category_name",

			"supplier_code": 1,
			"supplier_name": "$supplier.supplier_name",

			"brand_code": 1,
			"brand_name": "$brand.brand_name",

			"warehouse_name": "$stock.warehouses_name",
			"lot_no":         "$stock.lots_no",

			"balance_qty": 1,
			"unit":        1,
			"cost_price":  1,
			"status":      1,
			"created_by":  1,
			"updated_by":  1,
			"created_at":  1,
			"updated_at":  1,
		}},
	},
	)

	cur, err := s.db.
		Collection(models.Product{}.CollectionName()).
		Aggregate(ctx, pipeline)
	if err != nil {

		end := time.Now()
		logProduct := models.TransactionLog{
			RequestID:          "",
			FunctionEndpoint:   endpoint,
			FunctionMethod:     "GET",
			FunctionName:       "SearchProduct",
			FunctionController: "Product",
			Environment:        "local",
			QueryCollection:    "product_master",
			QueryType:          "query",
			StartTime:          now,
			EndTime:            end,
			DurationMs:         end.Sub(now).Milliseconds(),
			CountData:          0,
			StatusCode:         400,
			StatusMessage:      "Aggregate fali",
			CreatedBy:          "admin",
			CreatedAt:          now,
		}
		_, _ = s.db.Collection(logProduct.CollectionName()).InsertOne(ctx, logProduct)

		return nil, err
	}
	defer cur.Close(ctx)

	var result []bson.M
	if err := cur.All(ctx, &result); err != nil {

		end := time.Now()
		logProduct := models.TransactionLog{
			RequestID:          "",
			FunctionEndpoint:   endpoint,
			FunctionMethod:     "GET",
			FunctionName:       "SearchProduct",
			FunctionController: "Product",
			Environment:        "local",
			QueryCollection:    "product_master",
			QueryType:          "query",
			StartTime:          now,
			EndTime:            end,
			DurationMs:         end.Sub(now).Milliseconds(),
			CountData:          len(result),
			StatusCode:         400,
			StatusMessage:      "Result fali",
			CreatedBy:          "admin",
			CreatedAt:          now,
		}
		_, _ = s.db.Collection(logProduct.CollectionName()).InsertOne(ctx, logProduct)

		return nil, err
	}

	end := time.Now()
	logProduct := models.TransactionLog{
		RequestID:          "",
		FunctionEndpoint:   endpoint,
		FunctionMethod:     "GET",
		FunctionName:       "SearchProduct",
		FunctionController: "Product",
		Environment:        "local",
		QueryCollection:    "product_master",
		QueryType:          "query",
		StartTime:          now,
		EndTime:            end,
		DurationMs:         end.Sub(now).Milliseconds(),
		CountData:          len(result),
		StatusCode:         200,
		StatusMessage:      "success",
		CreatedBy:          "admin",
		CreatedAt:          now,
	}
	_, _ = s.db.Collection(logProduct.CollectionName()).InsertOne(ctx, logProduct)

	return result, nil
}
