package handlers_test

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

func createAuthToken(t *testing.T, r http.Handler, email, password string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("login failed with status %d: %s", w.Code, w.Body.String())
	}

	var res struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	return res.Token
}

func TestEventHandler_CRUD(t *testing.T) {
	db := testutil.MustSetupDB(t)

	panitia1, err := testutil.CreateTestUser(db, "panitia1", "panitia1@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	if err != nil {
		t.Fatalf("failed to create panitia1: %v", err)
	}

	panitia2, err := testutil.CreateTestUser(db, "panitia2", "panitia2@eventgate.test", "Password123!", models.RoleAdminPanitia, true)
	if err != nil {
		t.Fatalf("failed to create panitia2: %v", err)
	}

	superadmin, err := testutil.CreateTestUser(db, "superadmin", "superadmin@eventgate.test", "Password123!", models.RoleSuperAdmin, true)
	if err != nil {
		t.Fatalf("failed to create superadmin: %v", err)
	}

	staf, err := testutil.CreateTestUser(db, "staf", "staf@eventgate.test", "Password123!", models.RoleStafLapangan, true)
	if err != nil {
		t.Fatalf("failed to create staf: %v", err)
	}

	r := router.New(db, "test-secret", 24)

	tokenPanitia1 := createAuthToken(t, r, panitia1.Email, "Password123!")
	tokenPanitia2 := createAuthToken(t, r, panitia2.Email, "Password123!")
	tokenSuperAdmin := createAuthToken(t, r, superadmin.Email, "Password123!")
	tokenStaf := createAuthToken(t, r, staf.Email, "Password123!")

	start := time.Now().Add(24 * time.Hour)
	end := start.Add(4 * time.Hour)

	// 1. Validation Failures on Create
	t.Run("Create Event Validation Failures", func(t *testing.T) {
		// Mandatory fields empty
		body, _ := json.Marshal(map[string]interface{}{
			"title": "",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+tokenPanitia1)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for empty fields, got %d", w.Code)
		}

		// Paid event with 0 price
		bodyPaidErr, _ := json.Marshal(map[string]interface{}{
			"title":       "Seminar Paid Error",
			"description": "Tech conference description",
			"banner":      "https://example.com/banner.jpg",
			"location":    "Jakarta",
			"start_time":  start,
			"end_time":    end,
			"is_paid":     true,
			"price":       0,
			"quota":       100,
		})
		reqPaidErr := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(bodyPaidErr))
		reqPaidErr.Header.Set("Authorization", "Bearer "+tokenPanitia1)
		reqPaidErr.Header.Set("Content-Type", "application/json")
		wPaidErr := httptest.NewRecorder()
		r.ServeHTTP(wPaidErr, reqPaidErr)
		if wPaidErr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for paid event with 0 price, got %d", wPaidErr.Code)
		}

		// Invalid End Time (end before start)
		bodyDateErr, _ := json.Marshal(map[string]interface{}{
			"title":       "Seminar Date Error",
			"description": "Tech conference description",
			"banner":      "https://example.com/banner.jpg",
			"location":    "Jakarta",
			"start_time":  end,
			"end_time":    start,
			"is_paid":     false,
			"price":       0,
			"quota":       100,
		})
		reqDateErr := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(bodyDateErr))
		reqDateErr.Header.Set("Authorization", "Bearer "+tokenPanitia1)
		reqDateErr.Header.Set("Content-Type", "application/json")
		wDateErr := httptest.NewRecorder()
		r.ServeHTTP(wDateErr, reqDateErr)
		if wDateErr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for end_time before start_time, got %d", wDateErr.Code)
		}
	})

	// 2. Role Restriction on Create
	t.Run("Create Event Role Authorization", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"title":       "Staf Event Attempt",
			"description": "Description",
			"banner":      "https://example.com/banner.jpg",
			"location":    "Location",
			"start_time":  start,
			"end_time":    end,
			"is_paid":     false,
			"quota":       50,
		})
		req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+tokenStaf)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403 for staf_lapangan creating event, got %d", w.Code)
		}
	})

	// 3. Successful Create by Admin Panitia & Super Admin
	var event1ID uint
	t.Run("Successful Event Creation", func(t *testing.T) {
		// Admin Panitia creates Event 1
		body1, _ := json.Marshal(map[string]interface{}{
			"title":       "Go Workshop 2026",
			"description": "Belajar Golang dari dasar hingga mahir.",
			"banner":      "https://example.com/go.png",
			"location":    "Auditorium JTK",
			"start_time":  start,
			"end_time":    end,
			"is_paid":     true,
			"price":       50000,
			"quota":       100,
		})
		req1 := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(body1))
		req1.Header.Set("Authorization", "Bearer "+tokenPanitia1)
		req1.Header.Set("Content-Type", "application/json")
		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, req1)
		if w1.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created for panitia1, got %d: %s", w1.Code, w1.Body.String())
		}

		var createdEvent models.Event
		_ = json.Unmarshal(w1.Body.Bytes(), &createdEvent)
		if createdEvent.ID == 0 || createdEvent.Title != "Go Workshop 2026" {
			t.Fatalf("unexpected event response: %+v", createdEvent)
		}
		if createdEvent.Status != models.EventStatusDraft {
			t.Fatalf("expected initial status 'draft', got '%s'", createdEvent.Status)
		}
		if createdEvent.CreatedByID != panitia1.ID {
			t.Fatalf("expected created_by %d, got %d", panitia1.ID, createdEvent.CreatedByID)
		}
		event1ID = createdEvent.ID

		// Super Admin creates Event 2
		body2, _ := json.Marshal(map[string]interface{}{
			"title":       "Global Tech Summit",
			"description": "Konferensi Teknologi Tahunan",
			"banner":      "https://example.com/summit.png",
			"location":    "Main Hall",
			"start_time":  start,
			"end_time":    end,
			"is_paid":     false,
			"quota":       500,
		})
		req2 := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(body2))
		req2.Header.Set("Authorization", "Bearer "+tokenSuperAdmin)
		req2.Header.Set("Content-Type", "application/json")
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, req2)
		if w2.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created for superadmin, got %d", w2.Code)
		}
	})

	// 4. List Events & Detail
	t.Run("List and Detail Events", func(t *testing.T) {
		// List
		reqList := httptest.NewRequest(http.MethodGet, "/api/events", nil)
		reqList.Header.Set("Authorization", "Bearer "+tokenStaf)
		wList := httptest.NewRecorder()
		r.ServeHTTP(wList, reqList)
		if wList.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for list events, got %d", wList.Code)
		}
		var list []models.Event
		_ = json.Unmarshal(wList.Body.Bytes(), &list)
		if len(list) < 2 {
			t.Fatalf("expected at least 2 events in list, got %d", len(list))
		}

		// Detail
		reqDetail := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/events/%d", event1ID), nil)
		reqDetail.Header.Set("Authorization", "Bearer "+tokenStaf)
		wDetail := httptest.NewRecorder()
		r.ServeHTTP(wDetail, reqDetail)
		if wDetail.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for detail event, got %d", wDetail.Code)
		}
	})

	// 5. Ownership RBAC on Update & Delete
	t.Run("Update and Delete Ownership RBAC", func(t *testing.T) {
		// Panitia 2 tries to update Panitia 1's event -> 403 Forbidden
		bodyUpdate, _ := json.Marshal(map[string]interface{}{
			"title":       "Hijacked Workshop",
			"description": "Deskripsi baru",
			"banner":      "https://example.com/go.png",
			"location":    "Auditorium JTK",
			"start_time":  start,
			"end_time":    end,
			"is_paid":     true,
			"price":       75000,
			"quota":       100,
		})
		reqUp2 := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/events/%d", event1ID), bytes.NewReader(bodyUpdate))
		reqUp2.Header.Set("Authorization", "Bearer "+tokenPanitia2)
		reqUp2.Header.Set("Content-Type", "application/json")
		wUp2 := httptest.NewRecorder()
		r.ServeHTTP(wUp2, reqUp2)
		if wUp2.Code != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden when non-owner panitia updates event, got %d", wUp2.Code)
		}

		// Panitia 1 updates own event -> 200 OK
		reqUp1 := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/events/%d", event1ID), bytes.NewReader(bodyUpdate))
		reqUp1.Header.Set("Authorization", "Bearer "+tokenPanitia1)
		reqUp1.Header.Set("Content-Type", "application/json")
		wUp1 := httptest.NewRecorder()
		r.ServeHTTP(wUp1, reqUp1)
		if wUp1.Code != http.StatusOK {
			t.Fatalf("expected 200 OK when owner panitia updates event, got %d: %s", wUp1.Code, wUp1.Body.String())
		}

		// Panitia 2 tries to delete Panitia 1's event -> 403 Forbidden
		reqDel2 := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/events/%d", event1ID), nil)
		reqDel2.Header.Set("Authorization", "Bearer "+tokenPanitia2)
		wDel2 := httptest.NewRecorder()
		r.ServeHTTP(wDel2, reqDel2)
		if wDel2.Code != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden when non-owner panitia deletes event, got %d", wDel2.Code)
		}

		// Super Admin deletes Panitia 1's event (SuperAdmin override) -> 200 OK
		reqDelSA := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/events/%d", event1ID), nil)
		reqDelSA.Header.Set("Authorization", "Bearer "+tokenSuperAdmin)
		wDelSA := httptest.NewRecorder()
		r.ServeHTTP(wDelSA, reqDelSA)
		if wDelSA.Code != http.StatusOK {
			t.Fatalf("expected 200 OK when superadmin deletes event, got %d", wDelSA.Code)
		}

		// Verify event 1 is deleted
		reqGetDeleted := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/events/%d", event1ID), nil)
		reqGetDeleted.Header.Set("Authorization", "Bearer "+tokenPanitia1)
		wGetDeleted := httptest.NewRecorder()
		r.ServeHTTP(wGetDeleted, reqGetDeleted)
		if wGetDeleted.Code != http.StatusNotFound {
			t.Fatalf("expected 404 Not Found for deleted event, got %d", wGetDeleted.Code)
		}
	})
}
