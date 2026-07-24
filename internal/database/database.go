package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"be-eventgate/config"
	"be-eventgate/internal/models"
)

// Connect membuka koneksi ke PostgreSQL menggunakan GORM.
func Connect(cfg config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db, nil
}

// Migrate menjalankan auto-migration untuk model-model yang terkait EVG-41.
// KALAU project Anda sudah punya migration system sendiri (misal golang-migrate
// dengan file .sql), GANTI fungsi ini / jangan dipanggil, dan buat migration
// .sql terpisah untuk tabel roles & users mengikuti pola project yang sudah ada.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.Role{}, &models.User{})
}
