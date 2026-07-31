package database_test

import (
	"fmt"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"be-eventgate/internal/database"
	"be-eventgate/internal/models"
)

func setupDB(t *testing.T) *gorm.DB {
	t.Helper()
	name := strings.ReplaceAll(t.Name(), "/", "_")
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", name)

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestSeedRoles_CreatesAllExpectedRoles(t *testing.T) {
	db := setupDB(t)

	if err := database.SeedRoles(db); err != nil {
		t.Fatalf("SeedRoles returned error: %v", err)
	}

	expected := []string{
		models.RoleSuperAdmin,
		models.RoleAdminPanitia,
		models.RoleStafLapangan,
		models.RolePeserta,
		models.RoleSchoolReviewer,
	}

	seenIDs := make(map[uint]bool)
	for _, roleName := range expected {
		var role models.Role
		if err := db.Where("role_name = ?", roleName).First(&role).Error; err != nil {
			t.Errorf("expected role %q to exist after seeding, got error: %v", roleName, err)
		}

		if role.ID == 0 {
			t.Errorf("role %q has ID 0, which means auto-increment might not be working", roleName)
		}
		if seenIDs[role.ID] {
			t.Errorf("duplicate role ID %d found for role %q", role.ID, roleName)
		}
		seenIDs[role.ID] = true
	}
}

func TestSeedRoles_IsIdempotent(t *testing.T) {
	db := setupDB(t)

	if err := database.SeedRoles(db); err != nil {
		t.Fatalf("first SeedRoles call failed: %v", err)
	}
	if err := database.SeedRoles(db); err != nil {
		t.Fatalf("second SeedRoles call failed: %v", err)
	}

	var count int64
	db.Model(&models.Role{}).Where("role_name = ?", models.RoleSuperAdmin).Count(&count)
	if count != 1 {
		t.Fatalf("expected exactly 1 super_admin role row after seeding twice, got %d", count)
	}
}
