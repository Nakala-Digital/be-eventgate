package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type responseEnvelopeTest struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	Errors  json.RawMessage `json:"errors"`
}

func decodeResponseEnvelope(t *testing.T, rr *httptest.ResponseRecorder) responseEnvelopeTest {
	t.Helper()
	var envelope responseEnvelopeTest
	if err := json.Unmarshal(rr.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to decode response envelope: %v", err)
	}
	if len(envelope.Data) == 0 || len(envelope.Errors) == 0 {
		t.Fatalf("expected both data and errors fields in response: %s", rr.Body.String())
	}
	return envelope
}

func TestWriteJSON_UsesStandardEnvelope(t *testing.T) {
	rr := httptest.NewRecorder()
	WriteJSON(rr, http.StatusOK, map[string]string{"id": "1"})

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected application/json content type, got %q", got)
	}

	envelope := decodeResponseEnvelope(t, rr)
	if !envelope.Success {
		t.Fatal("expected success=true")
	}
	if envelope.Message == "" {
		t.Fatal("expected a non-empty message")
	}
	if string(envelope.Errors) != "null" {
		t.Fatalf("expected errors=null for success response, got %s", envelope.Errors)
	}
}

func TestWriteError_UsesStandardEnvelope(t *testing.T) {
	rr := httptest.NewRecorder()
	WriteError(rr, http.StatusBadRequest, "invalid request body")

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}

	envelope := decodeResponseEnvelope(t, rr)
	if envelope.Success {
		t.Fatal("expected success=false")
	}
	if envelope.Message != "invalid request body" {
		t.Fatalf("expected error message to be preserved, got %q", envelope.Message)
	}
	if string(envelope.Data) != "null" {
		t.Fatalf("expected data=null for error response, got %s", envelope.Data)
	}
}
