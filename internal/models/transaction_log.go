package models

import "time"

type TransactionLog struct {
	RequestID          string    `json:"request_id" bson:"request_id"`
	FunctionEndpoint   string    `json:"function_endpoint" bson:"function_endpoint"`
	FunctionController string    `json:"function_controller" bson:"function_controller"`
	FunctionName       string    `json:"function_name" bson:"function_name"`
	FunctionMethod     string    `json:"function_method" bson:"function_method"`
	Environment        string    `json:"environment" bson:"environment"`
	QueryCollection    string    `json:"query_collection" bson:"query_collection"`
	QueryType          string    `json:"query_type" bson:"query_type"`
	StartTime          time.Time `json:"start_time" bson:"start_time"`
	EndTime            time.Time `json:"end_time" bson:"end_time"`
	DurationMs         int64     `json:"duration_ms" bson:"duration_ms"`
	CountData          int       `json:"count_data" bson:"count_data"`
	StatusCode         int       `json:"status_code" bson:"status_code"`
	StatusMessage      string    `json:"status_message" bson:"status_message"`
	UserID             string    `json:"user_id" bson:"user_id"`
	Role               string    `json:"role" bson:"role"`
	Createbsony        string    `json:"created_by" bson:"created_by"`
	CreatedAt          time.Time `json:"created_at" bson:"created_at"`
}

func (TransactionLog) CollectionName() string {
	return "transaction_log"
}
