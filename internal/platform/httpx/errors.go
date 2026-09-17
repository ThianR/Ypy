package httpx

import (
	"encoding/json"
	"net/http"
)

type Error struct {
	Code          string `json:"code"`
	Message       string `json:"message"`
	CorrelationID string `json:"correlation_id"`
	Retryable     bool   `json:"retryable"`
}

func WriteError(w http.ResponseWriter, status int, code, message, correlationID string, retryable bool) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Error{Code: code, Message: message, CorrelationID: correlationID, Retryable: retryable})
}
