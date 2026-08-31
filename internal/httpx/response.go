package httpx

import (
	"net/http"

	appresponse "be-eventgate/pkg/utils/response"
)

// Response merepresentasikan struktur JSON baku untuk respons API.
type Response = appresponse.Response

// WriteJSON menyusun respons sukses menggunakan JSON envelope baku aplikasi.
func WriteJSON(w http.ResponseWriter, status int, payload interface{}) {
	appresponse.Success(w, status, "request completed successfully", payload)
}

// WriteSuccess menyusun respons sukses dengan pesan deskriptif dan payload.
func WriteSuccess(w http.ResponseWriter, status int, message string, payload interface{}) {
	appresponse.Success(w, status, message, payload)
}

// WriteError mengirimkan respons kesalahan menggunakan JSON envelope baku aplikasi.
func WriteError(w http.ResponseWriter, status int, message string) {
	appresponse.Error(w, status, message, nil)
}
