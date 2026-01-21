package unit

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"kakaidee-backend/internal/models"
)

type UnitService struct {
	db *mongo.Database
}

func NewUnitService(db *mongo.Database) *UnitService {
	return &UnitService{db: db}
}

func (s *UnitService) Get(
	ctx context.Context,
	keyword string,
	now time.Time,
) ([]bson.M, error) {
	pipeline := mongo.Pipeline{}
	if keyword != "" {
		pipeline = append(pipeline, bson.D{
			{"$match", bson.M{
				"$or": []bson.M{
					{"name": bson.M{"$regex": keyword, "$options": "i"}},
				},
			}},
		})
	}

	cur, err := s.db.
		Collection(models.Unit{}.CollectionName()).
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
	logUnit := models.TransactionLog{
		RequestID:          "",
		FunctionEndpoint:   "unit?keyword=" + keyword,
		FunctionMethod:     "GET",
		FunctionName:       "SearchUnit",
		FunctionController: "Unit",
		Environment:        "local",
		QueryCollection:    "unit",
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
	_, _ = s.db.Collection(logUnit.CollectionName()).InsertOne(ctx, logUnit)

	return result, nil
}
