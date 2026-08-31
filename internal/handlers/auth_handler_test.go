package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"be-eventgate/internal/handlers"
	"be-eventgate/internal/models"
	"be-eventgate/internal/testutil"
)

func TestLogin_Success(t *testing.T) {
	db := testutil.MustSetupDB(t)
	_, err := testutil.CreateTestUser(db, "adminpanitia1", "panitia1@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	h := handlers.NewAuthHandler(db, "test-secret", 24)

	body, _ := json.Marshal(map[string]string{
		"email":    "panitia1@eventgate.test",
		"password": "Password123!",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", rr.Code, rr.Body.String())
	}

	var resp struct {
		Success bool                   `json:"success"`
		Message string                 `json:"message"`
		Data    handlers.LoginResponse `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Success {
		t.Errorf("expected success true, got false")
	}
	if resp.Data.Token == "" {
		t.Fatal("expected a non-empty token in the login response")
	}
	if resp.Data.User.RoleName != models.RoleAdminPanitia {
		t.Errorf("expected role_name %s, got %s", models.RoleAdminPanitia, resp.Data.User.RoleName)
	}
	if resp.Data.User.Email != "panitia1@eventgate.test" {
		t.Errorf("expected email panitia1@eventgate.test, got %s", resp.Data.User.Email)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	db := testutil.MustSetupDB(t)
	_, err := testutil.CreateTestUser(db, "stafflap1", "staff1@eventgate.test", "Password123!", models.RoleStafLapangan, true)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	h := handlers.NewAuthHandler(db, "test-secret", 24)

	body, _ := json.Marshal(map[string]string{
		"email":    "staff1@eventgate.test",
		"password": "WrongPassword",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for wrong password, got %d", rr.Code)
	}
}

func TestLogin_UnknownEmail(t *testing.T) {
	db := testutil.MustSetupDB(t)
	h := handlers.NewAuthHandler(db, "test-secret", 24)

	body, _ := json.Marshal(map[string]string{
		"email":    "notregistered@eventgate.test",
		"password": "whatever",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for an email that is not registered, got %d", rr.Code)
	}
}

func TestLogin_InactiveUser(t *testing.T) {
	db := testutil.MustSetupDB(t)
	_, err := testutil.CreateTestUser(db, "inactive1", "inactive1@eventgate.test", "Password123!", models.RoleAdminPanitia, false)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	h := handlers.NewAuthHandler(db, "test-secret", 24)

	body, _ := json.Marshal(map[string]string{
		"email":    "inactive1@eventgate.test",
		"password": "Password123!",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for an inactive user, got %d", rr.Code)
	}
}

func TestLogin_MissingFields(t *testing.T) {
	db := testutil.MustSetupDB(t)
	h := handlers.NewAuthHandler(db, "test-secret", 24)

	body, _ := json.Marshal(map[string]string{"email": "onlyemail@eventgate.test"})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when password is missing, got %d", rr.Code)
	}
}
