package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"be-eventgate/internal/handlers"
	"be-eventgate/internal/middleware"
	"be-eventgate/internal/models"
	"be-eventgate/internal/testutil"
)

func contextWithUser(ctx context.Context, userID uint, username, roleName string) context.Context {
	ctx = context.WithValue(ctx, middleware.ContextUserID, userID)
	ctx = context.WithValue(ctx, middleware.ContextUsername, username)
	ctx = context.WithValue(ctx, middleware.ContextRoleName, roleName)
	return ctx
}

func TestMe_Success(t *testing.T) {
	db := testutil.MustSetupDB(t)
	user, err := testutil.CreateTestUser(db, "superadmin1", "superadmin1@eventgate.test", "Password123!", models.RoleSuperAdmin, true)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	h := handlers.NewUserHandler(db)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req = req.WithContext(contextWithUser(req.Context(), user.ID, user.Username, models.RoleSuperAdmin))
	rr := httptest.NewRecorder()

	h.Me(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", rr.Code, rr.Body.String())
	}

	profile := testutil.DecodeData[handlers.UserProfile](t, rr.Body.Bytes())
	if profile.Username != "superadmin1" {
		t.Errorf("expected username superadmin1, got %s", profile.Username)
	}
	if profile.RoleName != models.RoleSuperAdmin {
		t.Errorf("expected role_name %s, got %s", models.RoleSuperAdmin, profile.RoleName)
	}
	if !profile.IsActive {
		t.Error("expected is_active to be true")
	}
}

func TestMe_Unauthorized_NoContext(t *testing.T) {
	db := testutil.MustSetupDB(t)
	h := handlers.NewUserHandler(db)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	rr := httptest.NewRecorder()

	h.Me(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 when there is no authenticated user in context, got %d", rr.Code)
	}
}
