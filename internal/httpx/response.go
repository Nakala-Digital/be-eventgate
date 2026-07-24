package httpx

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse adalah format JSON konsisten untuk semua error di API.
type ErrorResponse struct {
	Error string `json:"error"`
}

// WriteJSON menulis payload sebagai JSON dengan status code tertentu.
func WriteJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// WriteError menulis error response JSON konsisten: {"error": "..."}.
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, ErrorResponse{Error: message})
}
