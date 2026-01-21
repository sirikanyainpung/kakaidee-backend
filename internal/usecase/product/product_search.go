package product

import (
	"context"
	"math/rand"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"kakaidee-backend/internal/helper"
	"kakaidee-backend/internal/models"
)

// func (s *ProductService) GetV1(
// 	ctx context.Context,
// 	keyword string,
// 	now time.Time,
// ) ([]bson.M, error) {

// 	// now := time.Now()

// 	pipeline := mongo.Pipeline{}

// 	// ===== 1. match (search) =====
// 	if keyword != "" {
// 		pipeline = append(pipeline, bson.D{
// 			{"$match", bson.M{
// 				"$or": []bson.M{
// 					{"barcode": bson.M{"$regex": keyword, "$options": "i"}},
// 					{"sku_code": bson.M{"$regex": keyword, "$options": "i"}},
// 					{"product_name": bson.M{"$regex": keyword, "$options": "i"}},
// 				},
// 			}},
// 		})
// 	}

// 	// ===== 2. lookup category =====
// 	pipeline = append(pipeline, bson.D{
// 		{"$lookup", bson.M{
// 			"from":         "category_masters",
// 			"localField":   "category_code",
// 			"foreignField": "category_code",
// 			"as":           "category",
// 		}},
// 	})

// 	// ===== 3. lookup supplier =====
// 	pipeline = append(pipeline, bson.D{
// 		{"$lookup", bson.M{
// 			"from":         "supplier_masters",
// 			"localField":   "supplier_code",
// 			"foreignField": "supplier_code",
// 			"as":           "supplier",
// 		}},
// 	})

// 	// ===== 4. lookup brand =====
// 	pipeline = append(pipeline, bson.D{
// 		{"$lookup", bson.M{
// 			"from":         "brand_masters",
// 			"localField":   "brand_code",
// 			"foreignField": "brand_code",
// 			"as":           "brand",
// 		}},
// 	})

// 	// ===== 5. unwind =====
// 	pipeline = append(pipeline,
// 		bson.D{{"$unwind", bson.M{"path": "$category", "preserveNullAndEmptyArrays": true}}},
// 		bson.D{{"$unwind", bson.M{"path": "$supplier", "preserveNullAndEmptyArrays": true}}},
// 		bson.D{{"$unwind", bson.M{"path": "$brand", "preserveNullAndEmptyArrays": true}}},
// 	)

// 	// ===== 6. project (flat response) =====
// 	pipeline = append(pipeline, bson.D{
// 		{"$project", bson.M{
// 			"_id": 0, // ⭐ ปิด _id

// 			"barcode":             1,
// 			"sku_code":            1,
// 			"product_name":        1,
// 			"product_description": 1,

// 			"category_code": 1,
// 			"category_name": "$category.category_name",

// 			"supplier_code": 1,
// 			"supplier_name": "$supplier.supplier_name",

// 			"brand_code": 1,
// 			"brand_name": "$brand.brand_name",

// 			"balance_qty": 1,
// 			"unit":        1,
// 			"cost_price":  1,
// 			"status":      1,
// 			"created_by":  1,
// 			"updated_by":  1,
// 			"created_at":  1,
// 			"updated_at":  1,
// 		}},
// 	})

// 	cur, err := s.db.
// 		Collection(models.Product{}.CollectionName()).
// 		Aggregate(ctx, pipeline)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer cur.Close(ctx)

// 	var result []bson.M
// 	if err := cur.All(ctx, &result); err != nil {
// 		return nil, err
// 	}

// 	end := time.Now()
// 	logProduct := models.TransactionLog{
// 		RequestID:          "",
// 		FunctionEndpoint:   "product?keyword=" + keyword,
// 		FunctionMethod:     "GET",
// 		FunctionName:       "SearchProduct",
// 		FunctionController: "Product",
// 		Environment:        "local",
// 		QueryCollection:    "product_master",
// 		QueryType:          "query",
// 		StartTime:          now,
// 		EndTime:            end,
// 		DurationMs:         end.Sub(now).Milliseconds(),
// 		CountData:          len(result),
// 		StatusCode:         200,
// 		StatusMessage:      "success",
// 		CreatedBy:          "admin",
// 		CreatedAt:          now,
// 	}
// 	_, _ = s.db.Collection(logProduct.CollectionName()).InsertOne(ctx, logProduct)

// 	return result, nil
// }

// =========== all-product ===========
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

	endpoint := "product/all-product?keyword=" + keyword + "&sku_code=" + skuCode + "&category_code=" + categoryCode + "&warehouse_name=" + warehouseName + "&lot_no=" + lotsNo + "&status=" + status
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
		match["category_code"] = ToInt32(categoryCode)
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
		"from":         "category_masters",
		"localField":   "category_code",
		"foreignField": "category_code",
		"as":           "category",
	}}})

	// ===== 6. lookup supplier =====
	pipeline = append(pipeline, bson.D{{"$lookup", bson.M{
		"from":         "supplier_masters",
		"localField":   "supplier_code",
		"foreignField": "supplier_code",
		"as":           "supplier",
	}}})

	// ===== 7. lookup brand =====
	pipeline = append(pipeline, bson.D{{"$lookup", bson.M{
		"from":         "brand_masters",
		"localField":   "brand_code",
		"foreignField": "brand_code",
		"as":           "brand",
	}}})

	// ===== 8. lookup unit =====
	pipeline = append(pipeline, bson.D{{"$lookup", bson.M{
		"from":         "unit",
		"localField":   "unit",      // product_master.unit
		"foreignField": "unit_code", // unit.unit_code
		"as":           "unit_info",
	}}})

	// ===== 9. unwind master =====
	pipeline = append(pipeline,
		bson.D{{"$unwind", bson.M{"path": "$stock", "preserveNullAndEmptyArrays": true}}},
		bson.D{{"$unwind", bson.M{"path": "$category", "preserveNullAndEmptyArrays": true}}},
		bson.D{{"$unwind", bson.M{"path": "$supplier", "preserveNullAndEmptyArrays": true}}},
		bson.D{{"$unwind", bson.M{"path": "$brand", "preserveNullAndEmptyArrays": true}}},
		bson.D{{"$unwind", bson.M{"path": "$unit_info", "preserveNullAndEmptyArrays": true}}},
	)

	// ===== 10. project (RESULT FINAL) =====
	pipeline = append(pipeline, bson.D{
		{"$project", bson.M{
			"_id": 0,

			"barcode":             1,
			"sku_code":            1,
			"product_name":        1,
			"product_description": 1,

			"brand_code": 1,
			"brand_name": "$brand.brand_name",

			"category_code":    1,
			"category_name_th": "$category.category_name_th",
			"category_name_en": "$category.category_name_en",

			"supplier_code": 1,
			"supplier_name": "$supplier.supplier_name",

			"cost_price":  1,
			"balance_qty": 1,
			"unit":        "$unit_info.unit_code",
			"unit_name":   "$unit_info.name",

			"warehouse_name": "$stock.warehouses_name",
			"warehouse_zone": "$stock.warehouses_zone",
			"bin":            "$stock.bin",
			"stock_type":     "$stock.stock_type",
			"lot_no":         "$stock.lots_no",
			"mfg":            "$stock.mfg",
			"exp":            "$stock.exp",

			"status":     1,
			"created_at": 1,
			"created_by": 1,
			"updated_at": 1,
			"updated_by": 1,
		}},
	})

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

// =========== filter-product ===========
func (s *ProductService) Get(
	ctx context.Context,
	keyword string,
	skuCode string,
	categoryCode string,
	warehouseName string,
	lotsNo string,
	status string,
	now time.Time,
	page int,
	limit int,
	skip int,
) ([]bson.M, error) {

	endpoint := "product?keyword="

	rand.Seed(time.Now().UnixNano())
	randomInt := rand.Intn(1000000000) + 1
	requestID := "102-" + strconv.Itoa(randomInt)

	pipeline := mongo.Pipeline{}

	// ===== 1. match product_master =====
	match := bson.M{}

	// Filter keyword
	if keyword != "" {
		match["$or"] = []bson.M{
			{"barcode": bson.M{"$regex": keyword, "$options": "i"}},
			{"product_name": bson.M{"$regex": keyword, "$options": "i"}},
		}
	}

	// Filter sku_code
	if skuCode != "" {
		match["sku_code"] = skuCode
	}

	// Filter category_code
	if categoryCode != "" {
		match["category_code"] = ToInt32(categoryCode)
	}

	// Filter status
	if status != "" {
		match["status"] = status
	}

	// Apply match filter
	if len(match) > 0 {
		pipeline = append(pipeline, bson.D{{"$match", match}})
	}

	helper.PrintStructJson(" ----- match ----- ")

	// ===== 2. lookup product_stock =====
	pipeline = append(pipeline, bson.D{{"$lookup", bson.M{
		"from":         "product_stock",
		"localField":   "barcode",
		"foreignField": "barcode",
		"as":           "stock",
	}}})

	helper.PrintStructJson(" ----- product_stock ----- ")

	// ===== 3. filter stock =====
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

	helper.PrintStructJson(" ----- filter stock ----- ")

	// ===== 4. lookup category =====
	pipeline = append(pipeline, bson.D{{"$lookup", bson.M{
		"from":         "category_masters",
		"localField":   "category_code",
		"foreignField": "category_code",
		"as":           "category",
	}}})

	helper.PrintStructJson(" ----- filter category ----- ")

	// ===== 5. lookup supplier =====
	pipeline = append(pipeline, bson.D{{"$lookup", bson.M{
		"from":         "supplier_masters",
		"localField":   "supplier_code",
		"foreignField": "supplier_code",
		"as":           "supplier",
	}}})

	helper.PrintStructJson(" ----- supplier ----- ")

	// ===== 6. lookup brand =====
	pipeline = append(pipeline, bson.D{{"$lookup", bson.M{
		"from":         "brand_masters",
		"localField":   "brand_code",
		"foreignField": "brand_code",
		"as":           "brand",
	}}})

	helper.PrintStructJson(" ----- brand ----- ")

	// ===== 7. lookup unit =====
	pipeline = append(pipeline, bson.D{{"$lookup", bson.M{
		"from":         "unit",
		"localField":   "unit",      // product_master.unit
		"foreignField": "unit_code", // unit.unit_code
		"as":           "unit_info",
	}}})

	helper.PrintStructJson(" ----- unit_info ----- ")

	// ===== 8. unwind master =====
	pipeline = append(pipeline,
		bson.D{{"$unwind", bson.M{"path": "$stock", "preserveNullAndEmptyArrays": true}}},
		bson.D{{"$unwind", bson.M{"path": "$category", "preserveNullAndEmptyArrays": true}}},
		bson.D{{"$unwind", bson.M{"path": "$supplier", "preserveNullAndEmptyArrays": true}}},
		bson.D{{"$unwind", bson.M{"path": "$brand", "preserveNullAndEmptyArrays": true}}},
		bson.D{{"$unwind", bson.M{"path": "$unit_info", "preserveNullAndEmptyArrays": true}}},
	)

	helper.PrintStructJson(" ----- unwind ----- ")

	// ===== 9. sort and pagination =====
	pipeline = append(pipeline, bson.D{
		{"$sort", bson.D{{"created_at", -1}}}, // sort by created_at DESC
	})

	pipeline = append(pipeline,
		bson.D{{"$skip", skip}},
		bson.D{{"$limit", limit}},
	)

	helper.PrintStructJson(" ----- skip ----- ")

	// ===== 10. project (RESULT FINAL) =====
	pipeline = append(pipeline, bson.D{
		{"$project", bson.M{
			"_id": 0,

			"barcode":             1,
			"sku_code":            1,
			"product_name":        1,
			"product_description": 1,

			"brand_code": 1,
			"brand_name": "$brand.brand_name",

			"category_code":    1,
			"category_name_th": "$category.category_name_th",
			"category_name_en": "$category.category_name_en",

			"supplier_code": 1,
			"supplier_name": "$supplier.supplier_name",

			"cost_price":  1,
			"balance_qty": 1,
			"unit":        "$unit_info.unit_code",
			"unit_name":   "$unit_info.name",

			"warehouse_name": "$stock.warehouses_name",
			"warehouse_zone": "$stock.warehouses_zone",
			"bin":            "$stock.bin",
			"stock_type":     "$stock.stock_type",
			"lot_no":         "$stock.lots_no",
			"mfg":            "$stock.mfg",
			"exp":            "$stock.exp",

			"status":     1,
			"created_at": 1,
			"created_by": 1,
			"updated_at": 1,
			"updated_by": 1,
		}},
	})

	helper.PrintStructJson(" ----- project ----- ")

	// ===== Execute Query =====
	cur, err := s.db.Collection(models.Product{}.CollectionName()).Aggregate(ctx, pipeline)
	if err != nil {
		helper.PrintStructJson(" ----- err Aggregate ----- ")
		helper.PrintStructJson(err)
		end := time.Now()
		logProduct := models.TransactionLog{
			RequestID:          requestID,
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
		helper.PrintStructJson(" ----- err Result ----- ")
		helper.PrintStructJson(err)
		end := time.Now()
		logProduct := models.TransactionLog{
			RequestID:          requestID,
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
		RequestID:          requestID,
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

// func (s *ProductService) Get(
// 	ctx context.Context,
// 	keyword string,
// 	skuCode string,
// 	categoryCode string,
// 	warehouseName string,
// 	lotsNo string,
// 	status string,
// 	now time.Time,
// 	page int,
// 	limit int,
// 	skip int,
// ) ([]bson.M, error) {

// 	endpoint := "product?keyword=" + keyword + "&sku_code=" + skuCode + "&category_code=" + categoryCode + "&warehouse_name=" + warehouseName + "&lot_no=" + lotsNo + "&status=" + status
// 	pipeline := mongo.Pipeline{}

// 	// ===== 1. match product_master =====
// 	match := bson.M{}

// 	if keyword != "" {
// 		match["$or"] = []bson.M{
// 			{"barcode": bson.M{"$regex": keyword, "$options": "i"}},
// 			{"product_name": bson.M{"$regex": keyword, "$options": "i"}},
// 		}
// 	}

// 	if skuCode != "" {
// 		match["sku_code"] = skuCode
// 	}

// 	if categoryCode != "" {
// 		match["category_code"] = ToInt32(categoryCode)
// 	}

// 	if status != "" {
// 		match["status"] = status
// 	}

// 	if len(match) > 0 {
// 		pipeline = append(pipeline, bson.D{{"$match", match}})
// 	}

// 	helper.PrintStructJson(" ----- match ----- ")

// 	// ===== 2. lookup product_stock =====
// 	pipeline = append(pipeline, bson.D{{"$lookup", bson.M{
// 		"from":         "product_stock",
// 		"localField":   "barcode",
// 		"foreignField": "barcode",
// 		"as":           "stock",
// 	}}})

// 	helper.PrintStructJson(" ----- product_stock ----- ")

// 	// ===== 4. filter stock =====
// 	stockMatch := bson.M{}

// 	if warehouseName != "" {
// 		stockMatch["stock.warehouses_name"] = warehouseName
// 	}

// 	if lotsNo != "" {
// 		stockMatch["stock.lots_no"] = lotsNo
// 	}

// 	if len(stockMatch) > 0 {
// 		pipeline = append(pipeline, bson.D{{"$match", stockMatch}})
// 	}

// 	helper.PrintStructJson(" ----- filter stock ----- ")

// 	// ===== 5. lookup category =====
// 	pipeline = append(pipeline, bson.D{{"$lookup", bson.M{
// 		"from":         "category_masters",
// 		"localField":   "category_code",
// 		"foreignField": "category_code",
// 		"as":           "category",
// 	}}})

// 	helper.PrintStructJson(" ----- filter category ----- ")

// 	// ===== 6. lookup supplier =====
// 	pipeline = append(pipeline, bson.D{{"$lookup", bson.M{
// 		"from":         "supplier_masters",
// 		"localField":   "supplier_code",
// 		"foreignField": "supplier_code",
// 		"as":           "supplier",
// 	}}})

// 	helper.PrintStructJson(" ----- supplier ----- ")

// 	// ===== 7. lookup brand =====
// 	pipeline = append(pipeline, bson.D{{"$lookup", bson.M{
// 		"from":         "brand_masters",
// 		"localField":   "brand_code",
// 		"foreignField": "brand_code",
// 		"as":           "brand",
// 	}}})

// 	helper.PrintStructJson(" ----- brand ----- ")

// 	// ===== 8. lookup unit =====
// 	pipeline = append(pipeline, bson.D{{"$lookup", bson.M{
// 		"from":         "unit",
// 		"localField":   "unit",      // product_master.unit
// 		"foreignField": "unit_code", // unit.unit_code
// 		"as":           "unit_info",
// 	}}})

// 	helper.PrintStructJson(" ----- unit_info ----- ")

// 	// ===== 9. unwind master =====
// 	pipeline = append(pipeline,
// 		bson.D{{"$unwind", bson.M{"path": "$stock", "preserveNullAndEmptyArrays": true}}},
// 		bson.D{{"$unwind", bson.M{"path": "$category", "preserveNullAndEmptyArrays": true}}},
// 		bson.D{{"$unwind", bson.M{"path": "$supplier", "preserveNullAndEmptyArrays": true}}},
// 		bson.D{{"$unwind", bson.M{"path": "$brand", "preserveNullAndEmptyArrays": true}}},
// 		bson.D{{"$unwind", bson.M{"path": "$unit_info", "preserveNullAndEmptyArrays": true}}},
// 	)

// 	helper.PrintStructJson(" ----- unwind ----- ")

// 	pipeline = append(pipeline, bson.D{
// 		{"$sort", bson.D{{"created_at", -1}}},
// 	})

// 	helper.PrintStructJson(" ----- sort ----- ")

// 	pipeline = append(pipeline,
// 		bson.D{{"$skip", skip}},
// 		bson.D{{"$limit", limit}},
// 	)

// 	helper.PrintStructJson(" ----- skip ----- ")

// 	// ===== 10. project (RESULT FINAL) =====
// 	pipeline = append(pipeline, bson.D{
// 		{"$project", bson.M{
// 			"_id": 0,

// 			"barcode":             1,
// 			"sku_code":            1,
// 			"product_name":        1,
// 			"product_description": 1,

// 			"brand_code": 1,
// 			"brand_name": "$brand.brand_name",

// 			"category_code":    1,
// 			"category_name_th": "$category.category_name_th",
// 			"category_name_en": "$category.category_name_en",

// 			"supplier_code": 1,
// 			"supplier_name": "$supplier.supplier_name",

// 			"cost_price":  1,
// 			"balance_qty": 1,
// 			"unit":        "$unit_info.unit_code",
// 			"unit_name":   "$unit_info.name",

// 			"warehouse_name": "$stock.warehouses_name",
// 			"warehouse_zone": "$stock.warehouses_zone",
// 			"bin":            "$stock.bin",
// 			"stock_type":     "$stock.stock_type",
// 			"lot_no":         "$stock.lots_no",
// 			"mfg":            "$stock.mfg",
// 			"exp":            "$stock.exp",

// 			"status":     1,
// 			"created_at": 1,
// 			"created_by": 1,
// 			"updated_at": 1,
// 			"updated_by": 1,
// 		}},
// 	})

// 	helper.PrintStructJson(" ----- project ----- ")

// 	cur, err := s.db.
// 		Collection(models.Product{}.CollectionName()).
// 		Aggregate(ctx, pipeline)
// 	if err != nil {
// 		helper.PrintStructJson(" ----- err Aggregate ----- ")
// 		helper.PrintStructJson(err)
// 		end := time.Now()
// 		logProduct := models.TransactionLog{
// 			RequestID:          "",
// 			FunctionEndpoint:   endpoint,
// 			FunctionMethod:     "GET",
// 			FunctionName:       "SearchProduct",
// 			FunctionController: "Product",
// 			Environment:        "local",
// 			QueryCollection:    "product_master",
// 			QueryType:          "query",
// 			StartTime:          now,
// 			EndTime:            end,
// 			DurationMs:         end.Sub(now).Milliseconds(),
// 			CountData:          0,
// 			StatusCode:         400,
// 			StatusMessage:      "Aggregate fali",
// 			CreatedBy:          "admin",
// 			CreatedAt:          now,
// 		}
// 		_, _ = s.db.Collection(logProduct.CollectionName()).InsertOne(ctx, logProduct)

// 		return nil, err
// 	}
// 	defer cur.Close(ctx)

// 	var result []bson.M
// 	if err := cur.All(ctx, &result); err != nil {
// 		helper.PrintStructJson(" ----- err Result ----- ")
// 		helper.PrintStructJson(err)
// 		end := time.Now()
// 		logProduct := models.TransactionLog{
// 			RequestID:          "",
// 			FunctionEndpoint:   endpoint,
// 			FunctionMethod:     "GET",
// 			FunctionName:       "SearchProduct",
// 			FunctionController: "Product",
// 			Environment:        "local",
// 			QueryCollection:    "product_master",
// 			QueryType:          "query",
// 			StartTime:          now,
// 			EndTime:            end,
// 			DurationMs:         end.Sub(now).Milliseconds(),
// 			CountData:          len(result),
// 			StatusCode:         400,
// 			StatusMessage:      "Result fali",
// 			CreatedBy:          "admin",
// 			CreatedAt:          now,
// 		}
// 		_, _ = s.db.Collection(logProduct.CollectionName()).InsertOne(ctx, logProduct)

// 		return nil, err
// 	}

// 	end := time.Now()
// 	logProduct := models.TransactionLog{
// 		RequestID:          "",
// 		FunctionEndpoint:   endpoint,
// 		FunctionMethod:     "GET",
// 		FunctionName:       "SearchProduct",
// 		FunctionController: "Product",
// 		Environment:        "local",
// 		QueryCollection:    "product_master",
// 		QueryType:          "query",
// 		StartTime:          now,
// 		EndTime:            end,
// 		DurationMs:         end.Sub(now).Milliseconds(),
// 		CountData:          len(result),
// 		StatusCode:         200,
// 		StatusMessage:      "success",
// 		CreatedBy:          "admin",
// 		CreatedAt:          now,
// 	}
// 	_, _ = s.db.Collection(logProduct.CollectionName()).InsertOne(ctx, logProduct)

// 	return result, nil
// }
