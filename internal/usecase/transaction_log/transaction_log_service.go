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
