package httpx

import (
	"net/http"

	appresponse "be-eventgate/pkg/utils/response"
)

// WriteJSON menyusun respons sukses menggunakan JSON envelope baku aplikasi.
func WriteJSON(w http.ResponseWriter, status int, payload interface{}) {
	appresponse.Success(w, status, "request completed successfully", payload)
}

// WriteError mengirimkan respons kesalahan menggunakan JSON envelope baku aplikasi.
func WriteError(w http.ResponseWriter, status int, message string) {
	appresponse.Error(w, status, message, nil)
}
