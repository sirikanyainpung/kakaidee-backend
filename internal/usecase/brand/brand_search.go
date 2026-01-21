package brand

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"kakaidee-backend/internal/models"
)

type BrandService struct {
	db *mongo.Database
}

func NewBrandService(db *mongo.Database) *BrandService {
	return &BrandService{db: db}
}

func (s *BrandService) Get(
	ctx context.Context,
	keyword string,
	now time.Time,
) ([]bson.M, error) {

	pipeline := mongo.Pipeline{}

	// ===== 1. match (search) =====
	if keyword != "" {
		pipeline = append(pipeline, bson.D{
			{"$match", bson.M{
				"$or": []bson.M{
					{"brand_code": bson.M{"$regex": keyword, "$options": "i"}},
					{"brand_name": bson.M{"$regex": keyword, "$options": "i"}},
				},
			}},
		})
	}

	cur, err := s.db.
		Collection(models.Brand{}.CollectionName()).
		Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var result []bson.M
	if err := cur.All(ctx, &result); err != nil {
		return nil, err
	}

	loc, _ := time.LoadLocation("Asia/Bangkok")
	end := time.Now().In(loc)
	logBrand := models.TransactionLog{
		RequestID:          "",
		FunctionEndpoint:   "brand?keyword=" + keyword,
		FunctionMethod:     "GET",
		FunctionName:       "SearchBrand",
		FunctionController: "Brand",
		Environment:        "local",
		QueryCollection:    "brand_masters",
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
	_, _ = s.db.Collection(logBrand.CollectionName()).InsertOne(ctx, logBrand)

	return result, nil
}
