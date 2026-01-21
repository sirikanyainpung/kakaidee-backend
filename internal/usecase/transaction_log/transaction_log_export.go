package transactionLog

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"kakaidee-backend/internal/models"
)

func (s *TransactionLogService) ExportTransactionLog(
	ctx context.Context,
	now time.Time,
) (string, string, error) {

	endpoint := "transaction-log/export"
	rand.Seed(time.Now().UnixNano())
	randomInt := rand.Intn(1000000000) + 1
	requestID := "202-" + strconv.Itoa(randomInt)

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
		return "", "", err
	}
	defer cur.Close(ctx)

	var result []bson.M
	if err := cur.All(ctx, &result); err != nil {
		return "", "", err
	}

	f := excelize.NewFile()
	sheet := "Products"
	f.SetSheetName("Sheet1", sheet)

	// ===== header =====
	headers := []string{
		"Function Name", "Function Method", "Query Type", "Request Id",
		"Start Time", "Total time (ms)", "Total Data", "Status Code", "Status message", "Created By", "User Id",
	}

	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// ===== result =====
	for r, row := range result {
		rowNum := r + 2

		f.SetCellValue(sheet, "A"+strconv.Itoa(rowNum), row["function_name"])
		f.SetCellValue(sheet, "B"+strconv.Itoa(rowNum), row["function_method"])
		f.SetCellValue(sheet, "C"+strconv.Itoa(rowNum), row["query_type"])
		f.SetCellValue(sheet, "D"+strconv.Itoa(rowNum), row["request_id"])
		f.SetCellValue(sheet, "E"+strconv.Itoa(rowNum), row["start_time"])
		f.SetCellValue(sheet, "F"+strconv.Itoa(rowNum), row["duration_ms"])
		f.SetCellValue(sheet, "G"+strconv.Itoa(rowNum), row["count_data"])
		f.SetCellValue(sheet, "H"+strconv.Itoa(rowNum), row["status_code"])
		f.SetCellValue(sheet, "I"+strconv.Itoa(rowNum), row["status_message"])
		f.SetCellValue(sheet, "J"+strconv.Itoa(rowNum), row["created_by"])
		f.SetCellValue(sheet, "K"+strconv.Itoa(rowNum), row["user_id"])
	}

	// ===== encode base64 =====
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return "", "", err
	}

	base64File := base64.StdEncoding.EncodeToString(buf.Bytes())

	fileName := fmt.Sprintf(
		"transaction_log_export_%s.xlsx",
		time.Now().Format("20060102_150405"),
	)

	loc, _ := time.LoadLocation("Asia/Bangkok")
	end := time.Now().In(loc)
	logProduct := models.TransactionLog{
		RequestID:          requestID,
		FunctionEndpoint:   endpoint,
		FunctionMethod:     "GET",
		FunctionName:       "ExportTransactionLog",
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
	_, _ = s.db.Collection(logProduct.CollectionName()).InsertOne(ctx, logProduct)

	return fileName, base64File, nil
}
