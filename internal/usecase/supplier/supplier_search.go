package supplier

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"kakaidee-backend/internal/models"
)

type SupplierService struct {
	db *mongo.Database
}

func NewSupplierService(db *mongo.Database) *SupplierService {
	return &SupplierService{db: db}
}

func (s *SupplierService) Get(
	ctx context.Context,
	keyword string,
	now time.Time,
) ([]bson.M, error) {
	pipeline := mongo.Pipeline{}

	if keyword != "" {
		pipeline = append(pipeline, bson.D{
			{"$match", bson.M{
				"$or": []bson.M{
					{"supplier_code": bson.M{"$regex": keyword, "$options": "i"}},
					{"supplier_name": bson.M{"$regex": keyword, "$options": "i"}},
				},
			}},
		})
	}

	cur, err := s.db.
		Collection(models.Supplier{}.CollectionName()).
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
		FunctionName:       "SearchSupplier",
		FunctionController: "Supplier",
		Environment:        "local",
		QueryCollection:    "supplier_masters",
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
