package httpx

import (
	"net/http"

	"be-eventgate/pkg/utils/response"
)

// Response merepresentasikan struktur JSON baku untuk respons API.
type Response = response.Response

// WriteJSON menyusun dan mengirimkan payload respons HTTP sukses dalam format JSON terstandarisasi envelope {success, message, data}.
func WriteJSON(w http.ResponseWriter, status int, payload interface{}) {
	response.Success(w, status, "success", payload)
}

// WriteSuccess menyusun dan mengirimkan respons HTTP sukses dengan pesan deskriptif dan payload.
func WriteSuccess(w http.ResponseWriter, status int, message string, payload interface{}) {
	response.Success(w, status, message, payload)
}

// WriteError mengirimkan respons kesalahan HTTP menggunakan format envelope {success: false, message: message}.
func WriteError(w http.ResponseWriter, status int, message string) {
	response.Error(w, status, message, nil)
}

