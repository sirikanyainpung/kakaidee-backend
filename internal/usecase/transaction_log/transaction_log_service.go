package transactionLog

import (
	"context"
	"math"
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
	page int,
	limit int,
	skip int,
) (map[string]interface{}, error) {

	rand.Seed(time.Now().UnixNano())
	randomInt := rand.Intn(1000000000) + 1
	requestID := "201-" + strconv.Itoa(randomInt)

	// Step 1: สร้าง pipeline สำหรับการคำนวณจำนวนข้อมูลทั้งหมด
	countPipeline := mongo.Pipeline{
		{{"$group", bson.D{
			{Key: "_id", Value: nil},                                    // ไม่แบ่งกลุ่ม
			{Key: "totalCount", Value: bson.D{{Key: "$sum", Value: 1}}}, // นับจำนวนเอกสาร
		}}},
	}

	// ดึงจำนวนข้อมูลทั้งหมดจาก countPipeline
	cur, err := s.db.
		Collection(models.TransactionLog{}.CollectionName()).
		Aggregate(ctx, countPipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var countResult []struct {
		TotalCount int `bson:"totalCount"`
	}

	if err := cur.All(ctx, &countResult); err != nil {
		return nil, err
	}

	// จำนวนข้อมูลทั้งหมด
	totalCount := 0
	if len(countResult) > 0 {
		totalCount = countResult[0].TotalCount
	}

	// คำนวณจำนวนหน้าทั้งหมด
	totalPages := 0
	if totalCount > 0 {
		totalPages = int(math.Ceil(float64(totalCount) / float64(limit)))
	}

	// Step 3: สร้าง pipeline สำหรับการดึงข้อมูลที่กรองตาม filter (match, skip, limit, project)
	pipeline := mongo.Pipeline{
		{{"$sort", bson.D{{"created_at", -1}}}}, // sort by created_at DESC
		{{"$skip", skip}},                       // skip ตามหน้าที่ต้องการ
		{{"$limit", limit}},                     // limit จำนวนข้อมูลตามที่ต้องการ
		{{"$project", bson.M{
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
		}}},
		{{"$unset", bson.A{"created_at", "end_time", "environment", "function_controller", "function_endpoint", "role"}}},
	}

	// Step 4: ดึงข้อมูลจาก MongoDB ตาม pipeline ที่กำหนด
	cur, err = s.db.
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

	// Step 5: บันทึกข้อมูล Log
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

	// Step 6: ส่งผลลัพธ์ทั้งหมด
	response := map[string]interface{}{
		"total_data":  totalCount, // จำนวนข้อมูลทั้งหมด
		"total_pages": totalPages, // จำนวนหน้าทั้งหมด
		"data":        result,     // ข้อมูลที่กรองตาม filter
	}

	return response, nil
}
