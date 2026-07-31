package router_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"be-eventgate/internal/models"
	"be-eventgate/internal/router"
	"be-eventgate/internal/testutil"
)

func doLogin(t *testing.T, baseURL, email, password string) (int, string) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	resp, err := http.Post(baseURL+"/api/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("login request error: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Token string `json:"token"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&result)
	return resp.StatusCode, result.Token
}

// TestRouter_FullAuthFlow menguji seluruh acceptance criteria EVG-41 secara
// end-to-end lewat HTTP nyata:
//  1. Akses protected route tanpa token -> 401
//  2. Login dengan credential valid -> 200 + token
//  3. Akses protected route dengan token valid -> 200
//  4. Akses route yang butuh role tertentu, dengan role yang salah -> 403
//  5. Akses route yang butuh role tertentu, dengan role yang benar -> 200
func TestRouter_FullAuthFlow(t *testing.T) {
	db := testutil.MustSetupDB(t)

	if _, err := testutil.CreateTestUser(db, "panitia_e2e", "panitia_e2e@eventgate.test", "Password123!", models.RoleAdminPanitia, true); err != nil {
		t.Fatalf("failed to create admin_panitia test user: %v", err)
	}
	if _, err := testutil.CreateTestUser(db, "superadmin_e2e", "superadmin_e2e@eventgate.test", "Password123!", models.RoleSuperAdmin, true); err != nil {
		t.Fatalf("failed to create super_admin test user: %v", err)
	}

	r := router.New(db, "test-secret", 24)
	server := httptest.NewServer(r)
	defer server.Close()

	// 1. Tanpa token -> 401
	resp, err := http.Get(server.URL + "/api/auth/me")
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for /me without token, got %d", resp.StatusCode)
	}

	// 2. Login admin_panitia
	status, panitiaToken := doLogin(t, server.URL, "panitia_e2e@eventgate.test", "Password123!")
	if status != http.StatusOK {
		t.Fatalf("expected 200 on login, got %d", status)
	}
	if panitiaToken == "" {
		t.Fatal("expected a non-empty token after successful login")
	}

	// 3. Akses /me dengan token valid -> 200
	req, _ := http.NewRequest(http.MethodGet, server.URL+"/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+panitiaToken)
	meResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("me request error: %v", err)
	}
	if meResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for /me with valid token, got %d", meResp.StatusCode)
	}

	// 4. admin_panitia mencoba akses route khusus super_admin -> 403
	req2, _ := http.NewRequest(http.MethodGet, server.URL+"/api/admin/ping", nil)
	req2.Header.Set("Authorization", "Bearer "+panitiaToken)
	pingResp, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("ping request error: %v", err)
	}
	if pingResp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for admin_panitia accessing super_admin-only route, got %d", pingResp.StatusCode)
	}

	// 5. super_admin login lalu akses route khusus super_admin -> 200
	status, superAdminToken := doLogin(t, server.URL, "superadmin_e2e@eventgate.test", "Password123!")
	if status != http.StatusOK {
		t.Fatalf("expected 200 on super_admin login, got %d", status)
	}

	req3, _ := http.NewRequest(http.MethodGet, server.URL+"/api/admin/ping", nil)
	req3.Header.Set("Authorization", "Bearer "+superAdminToken)
	pingResp2, err := http.DefaultClient.Do(req3)
	if err != nil {
		t.Fatalf("ping request error: %v", err)
	}
	if pingResp2.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for super_admin accessing super_admin-only route, got %d", pingResp2.StatusCode)
	}
}
