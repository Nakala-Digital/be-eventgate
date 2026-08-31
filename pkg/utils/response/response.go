package response

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Errors  interface{} `json:"errors"`
}

func JSON(w http.ResponseWriter, statusCode int, success bool, message string, data interface{}, errors interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := Response{
		Success: success,
		Message: message,
		Data:    data,
		Errors:  errors,
	}

	_ = json.NewEncoder(w).Encode(resp)
}

func Success(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	JSON(w, statusCode, true, message, data, nil)
}

func Error(w http.ResponseWriter, statusCode int, message string, errs interface{}) {
	JSON(w, statusCode, false, message, nil, errs)
}
