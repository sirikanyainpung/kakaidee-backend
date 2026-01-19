package helper

import (
	"strconv"
	"strings"
	"time"
)

type APIResponse struct {
	StatusCode    int         `json:"status_code"`
	StatusMessage string      `json:"status_message"`
	Message       string      `json:"message"`
	Result        interface{} `json:"result"`
}

func Success(message string, result interface{}) APIResponse {
	return APIResponse{
		StatusCode:    200,
		StatusMessage: "success",
		Message:       message,
		Result:        result,
	}
}

func Create(message string, result interface{}) APIResponse {
	return APIResponse{
		StatusCode:    201,
		StatusMessage: "created",
		Message:       message,
		Result:        result,
	}
}

func Error(code int, status, message string) APIResponse {
	return APIResponse{
		StatusCode:    code,
		StatusMessage: status,
		Message:       message,
		Result:        nil,
	}
}

func ToInt32(val string) int32 {
	i, _ := strconv.Atoi(strings.TrimSpace(val))
	return int32(i)
}

func ToInt(val string) int {
	i, _ := strconv.Atoi(strings.TrimSpace(val))
	return i
}

func ToFloat64(val string) float64 {
	v := strings.ReplaceAll(strings.TrimSpace(val), ",", "")
	f, _ := strconv.ParseFloat(v, 64)
	return f
}

func ToDate(val string) time.Time {
	if val == "" {
		return time.Time{}
	}
	t, _ := time.Parse("2006-01-02", val)
	return t
}
