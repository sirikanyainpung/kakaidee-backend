package category

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"kakaidee-backend/internal/models"
)

type CategoryService struct {
	db *mongo.Database
}

func NewCategoryService(db *mongo.Database) *CategoryService {
	return &CategoryService{db: db}
}

func (s *CategoryService) Get(
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
					{"category_code": bson.M{"$regex": keyword, "$options": "i"}},
					{"category_name": bson.M{"$regex": keyword, "$options": "i"}},
				},
			}},
		})
	}

	cur, err := s.db.
		Collection(models.Category{}.CollectionName()).
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
	logCategory := models.TransactionLog{
		RequestID:          "",
		FunctionEndpoint:   "category?keyword=" + keyword,
		FunctionMethod:     "GET",
		FunctionName:       "SearchCategory",
		FunctionController: "CategorySearch",
		Environment:        "local",
		QueryCollection:    "category_masters",
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
	_, _ = s.db.Collection(logCategory.CollectionName()).InsertOne(ctx, logCategory)

	return result, nil
}
