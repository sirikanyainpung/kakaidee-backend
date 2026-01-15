package helper

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
