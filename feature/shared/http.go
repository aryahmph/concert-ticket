package shared

import (
	"encoding/json"
	"net/http"
)

type StandardResponse struct {
	Message string      `json:"message,omitempty"`
	Errors  []string    `json:"errors,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func WriteErrorResponse(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(StandardResponse{
		Message: err.Error(),
		Data:    nil,
	})
}

func WriteInternalServerErrorResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(StandardResponse{
		Message: http.StatusText(http.StatusInternalServerError),
	})
}

func WriteSuccessResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(StandardResponse{
		Data: data,
	})
}
