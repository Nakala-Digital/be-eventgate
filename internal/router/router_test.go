package router_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestRouter_EventApprovalWorkflow(t *testing.T) {
	db := testutil.MustSetupDB(t)

	if _, err := testutil.CreateTestUser(db, "panitia_workflow", "panitia_workflow@eventgate.test", "Password123!", models.RoleAdminPanitia, true); err != nil {
		t.Fatalf("failed to create admin_panitia test user: %v", err)
	}
	if _, err := testutil.CreateTestUser(db, "superadmin_workflow", "superadmin_workflow@eventgate.test", "Password123!", models.RoleSuperAdmin, true); err != nil {
		t.Fatalf("failed to create super_admin test user: %v", err)
	}
	if _, err := testutil.CreateTestUser(db, "staf_workflow", "staf_workflow@eventgate.test", "Password123!", models.RoleStafLapangan, true); err != nil {
		t.Fatalf("failed to create staf_lapangan test user: %v", err)
	}

	r := router.New(db, "test-secret", 24)
	server := httptest.NewServer(r)
	defer server.Close()

	// 1. Tanpa token -> 401
	noAuthResp, err := http.Get(server.URL + "/api/events/1")
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	if noAuthResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", noAuthResp.StatusCode)
	}

	_, panitiaToken := doLogin(t, server.URL, "panitia_workflow@eventgate.test", "Password123!")
	_, superAdminToken := doLogin(t, server.URL, "superadmin_workflow@eventgate.test", "Password123!")
	_, stafToken := doLogin(t, server.URL, "staf_workflow@eventgate.test", "Password123!")

	// 2. Admin Panitia membuat event draft — payload sesuai kontrak API yang ADA
	createBody, _ := json.Marshal(map[string]interface{}{
		"title":       "Workflow Test Event",
		"description": "Deskripsi test",
		"banner":      "http://example.com/banner.jpg",
		"location":    "Aula",
		"start_time":  time.Now().Add(48 * time.Hour).Format(time.RFC3339),
		"end_time":    time.Now().Add(72 * time.Hour).Format(time.RFC3339),
		"is_paid":     false,
		"quota":       100,
	})
	createReq, _ := http.NewRequest(http.MethodPost, server.URL+"/api/events", bytes.NewReader(createBody))
	createReq.Header.Set("Authorization", "Bearer "+panitiaToken)
	createReq.Header.Set("Content-Type", "application/json")
	createResp, err := http.DefaultClient.Do(createReq)
	if err != nil {
		t.Fatalf("create event request error: %v", err)
	}
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 creating event, got %d", createResp.StatusCode)
	}
	var created struct {
		ID uint `json:"id"`
	}
	_ = json.NewDecoder(createResp.Body).Decode(&created)

	// 3. Admin Panitia mencoba publish -> harus 403 lewat middleware sungguhan
	publishAsPanitiaReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/events/%d/publish", server.URL, created.ID), nil)
	publishAsPanitiaReq.Header.Set("Authorization", "Bearer "+panitiaToken)
	publishAsPanitiaResp, err := http.DefaultClient.Do(publishAsPanitiaReq)
	if err != nil {
		t.Fatalf("publish request error: %v", err)
	}
	if publishAsPanitiaResp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 when admin_panitia tries to publish, got %d", publishAsPanitiaResp.StatusCode)
	}

	// 4. Butuh ticket type dulu (lewat DB langsung, belum ada endpoint CRUD-nya), lalu submit
	if _, err := testutil.CreateTestTicketType(db, created.ID); err != nil {
		t.Fatalf("failed to create ticket type: %v", err)
	}
	submitReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/events/%d/submit", server.URL, created.ID), nil)
	submitReq.Header.Set("Authorization", "Bearer "+panitiaToken)
	submitResp, err := http.DefaultClient.Do(submitReq)
	if err != nil {
		t.Fatalf("submit request error: %v", err)
	}
	if submitResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 submitting event, got %d", submitResp.StatusCode)
	}

	// 5. Super Admin approve -> approved
	approveReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/events/%d/approve", server.URL, created.ID), bytes.NewReader([]byte("{}")))
	approveReq.Header.Set("Authorization", "Bearer "+superAdminToken)
	approveResp, err := http.DefaultClient.Do(approveReq)
	if err != nil {
		t.Fatalf("approve request error: %v", err)
	}
	if approveResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 approving event, got %d", approveResp.StatusCode)
	}

	// 6. Super Admin publish -> published
	publishReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/events/%d/publish", server.URL, created.ID), bytes.NewReader([]byte("{}")))
	publishReq.Header.Set("Authorization", "Bearer "+superAdminToken)
	publishResp, err := http.DefaultClient.Do(publishReq)
	if err != nil {
		t.Fatalf("publish request error: %v", err)
	}
	if publishResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 publishing event, got %d", publishResp.StatusCode)
	}
	var published struct {
		Status string `json:"status"`
	}
	_ = json.NewDecoder(publishResp.Body).Decode(&published)
	if published.Status != models.EventStatusPublished {
		t.Fatalf("expected final status published, got %s", published.Status)
	}

	// 7. Staf Lapangan tidak boleh lihat detail event (scope URD: scan QR saja)
	stafGetReq, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/events/%d", server.URL, created.ID), nil)
	stafGetReq.Header.Set("Authorization", "Bearer "+stafToken)
	stafGetResp, err := http.DefaultClient.Do(stafGetReq)
	if err != nil {
		t.Fatalf("staf get request error: %v", err)
	}
	if stafGetResp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for staf_lapangan viewing event detail, got %d", stafGetResp.StatusCode)
	}
}
