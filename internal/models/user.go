package models

import (
	"time"

	"gorm.io/gorm"
)

// User mendefinisikan skema entitas pengguna sistem dalam basis data.
// Skema ini dikhususkan bagi peran yang memerlukan akses autentikasi log masuk 
// (seperti super_admin, admin_panitia, dll).
type User struct {
	ID          uint           `gorm:"primaryKey;column:user_id;autoIncrement" json:"id"`
	RoleID      uint           `gorm:"not null;index" json:"role_id"`
	Role        Role           `json:"role"`
	Username    string         `gorm:"uniqueIndex;size:100;not null" json:"username"`
	Email       string         `gorm:"uniqueIndex;size:150;not null" json:"email"`
	Password    string         `gorm:"size:255;not null" json:"-"`
	// Properti IsActive diatur secara manual melalui logika kode.
	// Penetapan nilai default GORM ditiadakan guna mencegah anomali data
	// saat penyimpanan nilai boolean false.
	IsActive bool `gorm:"not null" json:"is_active"`
	LastLoginAt *time.Time     `json:"last_login_at,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string { return "users" }
