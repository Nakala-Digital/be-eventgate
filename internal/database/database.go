package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"be-eventgate/config"
	"be-eventgate/internal/models"
)

// Connect menginisialisasi dan membuka koneksi ke PostgreSQL melalui GORM.
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

// Migrate mengeksekusi migrasi skema basis data secara otomatis untuk model terkait.
// Fungsionalitas ini dapat disesuaikan apabila proyek telah menggunakan
// sistem migrasi berbasis SQL secara terpisah.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Role{}, 
		&models.User{}, 
		&models.Event{}, 
		&models.EventApprovalLog{}, 
		&models.TicketType{},
		&models.DynamicQuestion{},
		&models.QuestionOption{},
	)
}
