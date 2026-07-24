package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"be-eventgate/internal/middleware"
)

func TestRequireRole_Forbidden_WrongRole(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/admin-only", nil)
	ctx := context.WithValue(req.Context(), middleware.ContextRoleName, "admin_panitia")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler := middleware.RequireRole("super_admin")(dummyOKHandler())
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for a role not in the allowed list, got %d", rr.Code)
	}
}

func TestRequireRole_Allowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/admin-only", nil)
	ctx := context.WithValue(req.Context(), middleware.ContextRoleName, "super_admin")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler := middleware.RequireRole("super_admin")(dummyOKHandler())
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for an allowed role, got %d", rr.Code)
	}
}

func TestRequireRole_NoRoleInContext(t *testing.T) {
	// Simulasi kesalahan urutan middleware: RequireRole dipasang tanpa
	// RequireAuth sebelumnya, jadi context kosong.
	req := httptest.NewRequest(http.MethodGet, "/admin-only", nil)
	rr := httptest.NewRecorder()

	handler := middleware.RequireRole("super_admin")(dummyOKHandler())
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 when role is missing from context, got %d", rr.Code)
	}
}

func TestRequireRole_MultipleAllowedRoles(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/staff-or-admin", nil)
	ctx := context.WithValue(req.Context(), middleware.ContextRoleName, "staf_lapangan")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler := middleware.RequireRole("super_admin", "staf_lapangan")(dummyOKHandler())
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 when role is one of several allowed roles, got %d", rr.Code)
	}
}
