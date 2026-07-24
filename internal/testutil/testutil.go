// Package testutil berisi helper YANG HANYA DIPAKAI OLEH TEST, tidak pernah
// diimpor oleh kode production (main.go / router.go). Karena itu dependency
// gorm.io/driver/sqlite di sini tidak ikut ke binary production, hanya ke
// binary test.
package testutil

import (
	"fmt"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"be-eventgate/internal/auth"
	"be-eventgate/internal/database"
	"be-eventgate/internal/models"
)

// MustSetupDB membuat instance SQLite in-memory yang unik per test (supaya
// antar test tidak saling mengganggu), sudah dimigrasi, dan sudah di-seed
// role. Gagal langsung t.Fatal kalau ada error.
func MustSetupDB(t *testing.T) *gorm.DB {
	t.Helper()

	name := strings.ReplaceAll(t.Name(), "/", "_")
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", name)

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}
	if err := database.SeedRoles(db); err != nil {
		t.Fatalf("failed to seed roles: %v", err)
	}
	return db
}

// CreateTestUser membuat user dengan password sudah di-hash dan role dicari
// berdasarkan nama role (harus sudah ada, mis. hasil SeedRoles).
func CreateTestUser(db *gorm.DB, username, email, plainPassword, roleName string, isActive bool) (*models.User, error) {
	var role models.Role
	if err := db.Where("role_name = ?", roleName).First(&role).Error; err != nil {
		return nil, fmt.Errorf("role %q not found, did you call MustSetupDB (which seeds roles)? %w", roleName, err)
	}

	hashed, err := auth.HashPassword(plainPassword)
	if err != nil {
		return nil, err
	}

	user := models.User{
		RoleID:   role.ID,
		Username: username,
		Email:    email,
		Password: hashed,
		IsActive: isActive,
	}
	if err := db.Create(&user).Error; err != nil {
		return nil, err
	}
	user.Role = role
	return &user, nil
}
