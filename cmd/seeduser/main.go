package main

import (
	"fmt"
	"log"

	"be-eventgate/config"
	"be-eventgate/internal/auth"
	"be-eventgate/internal/database"
	"be-eventgate/internal/models"
)

func main() {
	// Load config environment
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Gagal membaca konfigurasi: %v", err)
	}

	// Koneksi ke DB menggunakan GORM (karena kita mau insert data model)
	db, err := database.Connect(*cfg)
	if err != nil {
		log.Fatalf("Gagal connect ke DB: %v", err)
	}

	// Data user testing yang akan dibuat
	testUsers := []struct {
		Username string
		Email    string
		Password string
		RoleName string
		IsActive bool
	}{
		{"superadmin", "superadmin@eventgate.com", "Rahasia123!", models.RoleSuperAdmin, true},
		{"panitia", "panitia@eventgate.com", "Rahasia123!", models.RoleAdminPanitia, true},
		{"staf", "staf@eventgate.com", "Rahasia123!", models.RoleStafLapangan, true},
		{"reviewer", "reviewer@eventgate.com", "Rahasia123!", models.RoleSchoolReviewer, true},
		{"stafnonaktif", "stafnonaktif@eventgate.com", "Rahasia123!", models.RoleStafLapangan, false},
	}

	for _, tu := range testUsers {
		// Ambil ID role berdasarkan nama
		var role models.Role
		if err := db.Where("role_name = ?", tu.RoleName).First(&role).Error; err != nil {
			log.Printf("Gagal mencari role %s: %v. Pastikan aplikasi pernah dijalankan agar Role ter-seed.", tu.RoleName, err)
			continue
		}

		// Cek apakah user sudah ada
		var count int64
		db.Model(&models.User{}).Where("email = ?", tu.Email).Count(&count)
		if count > 0 {
			fmt.Printf("User %s sudah ada, skip.\n", tu.Email)
			continue
		}

		// Hash password
		hashed, err := auth.HashPassword(tu.Password)
		if err != nil {
			log.Fatalf("Gagal hash password: %v", err)
		}

		// Insert user
		user := models.User{
			RoleID:   role.ID,
			Username: tu.Username,
			Email:    tu.Email,
			Password: hashed,
			IsActive: tu.IsActive,
		}

		if err := db.Create(&user).Error; err != nil {
			log.Printf("Gagal insert user %s: %v", tu.Email, err)
		} else {
			fmt.Printf("✅ User berhasil dibuat! Email: %s | Pass: %s | Role: %s\n", tu.Email, tu.Password, tu.RoleName)
		}
	}

	fmt.Println("Proses pembuatan test user selesai!")
}
