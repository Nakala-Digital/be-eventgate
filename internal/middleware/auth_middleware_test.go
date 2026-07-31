package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"be-eventgate/internal/auth"
	"be-eventgate/internal/middleware"
)

func dummyOKHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestRequireAuth_MissingHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rr := httptest.NewRecorder()

	handler := middleware.RequireAuth("secret")(dummyOKHandler())
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 when Authorization header is missing, got %d", rr.Code)
	}
}

func TestRequireAuth_MalformedHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "NotBearer sometoken")
	rr := httptest.NewRecorder()

	handler := middleware.RequireAuth("secret")(dummyOKHandler())
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for malformed Authorization header, got %d", rr.Code)
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rr := httptest.NewRecorder()

	handler := middleware.RequireAuth("secret")(dummyOKHandler())
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for an invalid token, got %d", rr.Code)
	}
}

func TestRequireAuth_ValidToken(t *testing.T) {
	token, err := auth.GenerateToken("secret", 1, 1, "johndoe", "admin_panitia")
	if err != nil {
		t.Fatalf("failed generating token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler := middleware.RequireAuth("secret")(dummyOKHandler())
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for a valid token, got %d", rr.Code)
	}
}
