package auth

import (
	"testing"

	"be-eventgate/internal/models"
)

func TestIsAllowed(t *testing.T) {
	if !IsAllowed(models.RoleSuperAdmin, "event", "approve") {
		t.Error("super_admin should be allowed to approve event")
	}
	if IsAllowed(models.RoleAdminPanitia, "event", "approve") {
		t.Error("admin_panitia should NOT be allowed to approve event")
	}
	if IsAllowed(models.RolePeserta, "event", "publish") {
		t.Error("peserta should NOT be allowed to publish event")
	}
	if IsAllowed(models.RoleSuperAdmin, "resource_that_does_not_exist", "action") {
		t.Error("unknown resource+action combination should never be allowed")
	}
}
