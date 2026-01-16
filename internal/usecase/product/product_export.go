package product

import (
	"context"
	"kakaidee-backend/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *ProductService) Export(
	ctx context.Context,
	keyword string,
	startDate string,
	endDate string,
	now time.Time,
) ([]bson.M, error) {

	pipeline := mongo.Pipeline{}

	// ===== keyword filter =====
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

	// ===== date filter =====
	if startDate != "" || endDate != "" {
		dateCond := bson.M{}

		if startDate != "" {
			start, _ := time.Parse("2006-01-02", startDate)
			dateCond["$gte"] = start
		}
		if endDate != "" {
			end, _ := time.Parse("2006-01-02", endDate)
			dateCond["$lte"] = end.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}

		pipeline = append(pipeline, bson.D{
			{"$match", bson.M{
				"created_at": dateCond,
			}},
		})
	}

	// ===== lookup master =====
	pipeline = append(pipeline,
		bson.D{{"$lookup", bson.M{
			"from":         "category_masters",
			"localField":   "category_code",
			"foreignField": "category_code",
			"as":           "category",
		}}},
		bson.D{{"$lookup", bson.M{
			"from":         "supplier_masters",
			"localField":   "supplier_code",
			"foreignField": "supplier_code",
			"as":           "supplier",
		}}},
		bson.D{{"$lookup", bson.M{
			"from":         "brand_masters",
			"localField":   "brand_code",
			"foreignField": "brand_code",
			"as":           "brand",
		}}},
	)

	// ===== unwind =====
	pipeline = append(pipeline,
		bson.D{{"$unwind", bson.M{"path": "$category", "preserveNullAndEmptyArrays": true}}},
		bson.D{{"$unwind", bson.M{"path": "$supplier", "preserveNullAndEmptyArrays": true}}},
		bson.D{{"$unwind", bson.M{"path": "$brand", "preserveNullAndEmptyArrays": true}}},
	)

	// ===== project =====
	pipeline = append(pipeline, bson.D{
		{"$project", bson.D{
			{"_id", 0},
			{"barcode", 1},
			{"sku_code", 1},
			{"product_name", 1},
			{"product_description", 1},
			{"category_code", 1},
			{"category_name", "$category.category_name"},
			{"supplier_code", 1},
			{"supplier_name", "$supplier.supplier_name"},
			{"brand_code", 1},
			{"brand_name", "$brand.brand_name"},
			{"balance_qty", 1},
			{"unit", 1},
			{"cost_price", 1},
			{"created_at", 1},
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
		FunctionName:       "ExportProduct",
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
