package product

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"kakaidee-backend/internal/helper"
	"kakaidee-backend/internal/models"
)

func (s *ProductService) Get(
	ctx context.Context,
	keyword string,
) ([]models.Product, error) {

	filter := bson.M{}

	if keyword != "" {
		filter = bson.M{
			"$or": []bson.M{
				{"barcode": bson.M{"$regex": keyword, "$options": "i"}},
				{"sku_code": bson.M{"$regex": keyword, "$options": "i"}},
				{"product_name": bson.M{"$regex": keyword, "$options": "i"}},
			},
		}
	}

	helper.PrintStructJson("----- keyword -----")
	helper.PrintStructJson(keyword)

	helper.PrintStructJson("----- filter -----")
	helper.PrintStructJson(filter)

	cur, err := s.db.
		Collection(models.Product{}.CollectionName()).
		Find(ctx, filter, options.Find().SetSort(bson.M{"created_at": -1}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var products []models.Product
	if err := cur.All(ctx, &products); err != nil {
		return nil, err
	}

	now := time.Now()
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
		EndTime:            time.Now(),
		DurationMs:         time.Since(now).Milliseconds(),
		CountData:          len(products),
		StatusCode:         200,
		StatusMessage:      "success",
		CreatedBy:          "admin",
		CreatedAt:          time.Now(),
	}
	_, _ = s.db.Collection(logProduct.CollectionName()).InsertOne(ctx, logProduct)

	return products, nil
}
