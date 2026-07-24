package httpx

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse mendefinisikan struktur JSON baku untuk respons kesalahan API.
type ErrorResponse struct {
	Error string `json:"error"`
}

// WriteJSON menyusun dan mengirimkan payload respons HTTP dalam format JSON.
func WriteJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// WriteError mengirimkan respons kesalahan HTTP menggunakan format ErrorResponse.
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, ErrorResponse{Error: message})
}
