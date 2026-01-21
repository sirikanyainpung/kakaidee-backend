package transactionLog

import (
	"context"
	"math/rand"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"kakaidee-backend/internal/models"
)

type TransactionLogService struct {
	db *mongo.Database
}

func NewTransactionLogService(db *mongo.Database) *TransactionLogService {
	return &TransactionLogService{db: db}
}

func (s *TransactionLogService) GetTransactionLog(
	ctx context.Context,
	now time.Time,
) ([]bson.M, error) {

	rand.Seed(time.Now().UnixNano())
	randomInt := rand.Intn(1000000000) + 1
	requestID := "201-" + strconv.Itoa(randomInt)

	pipeline := mongo.Pipeline{}

	pipeline = append(pipeline, bson.D{
		{"$sort", bson.D{{"created_at", -1}}}, // sort by created_at DESC
	})

	// ขั้นตอนแรก: Project ฟิลด์ที่ต้องการ
	pipeline = append(pipeline, bson.D{
		{"$project", bson.M{
			"_id":              0,
			"count_data":       1,
			"created_by":       1,
			"duration_ms":      1,
			"function_method":  1,
			"function_name":    1,
			"query_collection": 1,
			"query_type":       1,
			"request_id":       1,
			"start_time":       1,
			"status_code":      1,
			"status_message":   1,
			"user_id":          1,
		}},
	})

	// ขั้นตอนที่สอง: ใช้ $unset เพื่อลบฟิลด์ที่ไม่ต้องการ
	pipeline = append(pipeline, bson.D{
		{"$unset", bson.A{"created_at", "end_time", "environment", "function_controller", "function_endpoint", "role"}},
	})

	cur, err := s.db.
		Collection(models.TransactionLog{}.CollectionName()).
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
	logCategory := models.TransactionLog{
		RequestID:          requestID,
		FunctionEndpoint:   "transaction-log",
		FunctionMethod:     "GET",
		FunctionName:       "GetTransactionLog",
		FunctionController: "TransactionLog",
		Environment:        "local",
		QueryCollection:    "transaction_log",
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
