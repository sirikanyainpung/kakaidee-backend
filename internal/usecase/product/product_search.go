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
			"from":         "category_master",
			"localField":   "category_code",
			"foreignField": "category_code",
			"as":           "category",
		}},
	})

	// ===== 3. lookup supplier =====
	pipeline = append(pipeline, bson.D{
		{"$lookup", bson.M{
			"from":         "supplier_master",
			"localField":   "supplier_code",
			"foreignField": "supplier_code",
			"as":           "supplier",
		}},
	})

	// ===== 4. lookup brand =====
	pipeline = append(pipeline, bson.D{
		{"$lookup", bson.M{
			"from":         "brand_master",
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

	// // ===== 6. project (flat response) =====
	// pipeline = append(pipeline, bson.D{
	// 	{"$project", bson.M{
	// 		"_id": 0, // ⭐ ปิด _id

	// 		"barcode":             1,
	// 		"sku_code":            1,
	// 		"product_name":        1,
	// 		"product_description": 1,

	// 		"category_code": 1,
	// 		"category_name": "$category.category_name",

	// 		"supplier_code": 1,
	// 		"supplier_name": "$supplier.supplier_name",

	// 		"brand_code": 1,
	// 		"brand_name": "$brand.brand_name",

	// 		"balance_qty": 1,
	// 		"unit":        1,
	// 		"cost_price":  1,
	// 		"status":      1,
	// 		"created_by":  1,
	// 		"updated_by":  1,
	// 		"created_at":  1,
	// 		"updated_at":  1,
	// 	}},
	// })

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
		FunctionController: "ProductSearch",
		Environment:        "local",
		QueryCollection:    "product",
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
// ) ([]models.Product, error) {

// 	filter := bson.M{}

// 	if keyword != "" {
// 		filter = bson.M{
// 			"$or": []bson.M{
// 				{"barcode": bson.M{"$regex": keyword, "$options": "i"}},
// 				{"sku_code": bson.M{"$regex": keyword, "$options": "i"}},
// 				{"product_name": bson.M{"$regex": keyword, "$options": "i"}},
// 			},
// 		}
// 	}

// 	helper.PrintStructJson("----- keyword -----")
// 	helper.PrintStructJson(keyword)

// 	helper.PrintStructJson("----- filter -----")
// 	helper.PrintStructJson(filter)

// 	cur, err := s.db.
// 		Collection(models.Product{}.CollectionName()).
// 		Find(ctx, filter, options.Find().SetSort(bson.M{"created_at": -1}))
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer cur.Close(ctx)

// 	var products []models.Product
// 	if err := cur.All(ctx, &products); err != nil {
// 		return nil, err
// 	}

// 	now := time.Now()
// 	logProduct := models.TransactionLog{
// 		RequestID:          "",
// 		FunctionEndpoint:   "product?keyword=" + keyword,
// 		FunctionMethod:     "GET",
// 		FunctionName:       "SearchProduct",
// 		FunctionController: "ProductSearch",
// 		Environment:        "local",
// 		QueryCollection:    "product",
// 		QueryType:          "query",
// 		StartTime:          now,
// 		EndTime:            time.Now(),
// 		DurationMs:         time.Since(now).Milliseconds(),
// 		CountData:          len(products),
// 		StatusCode:         200,
// 		StatusMessage:      "success",
// 		CreatedBy:          "admin",
// 		CreatedAt:          time.Now(),
// 	}
// 	_, _ = s.db.Collection(logProduct.CollectionName()).InsertOne(ctx, logProduct)

// 	return products, nil
// }
