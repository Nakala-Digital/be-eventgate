package models

import (
	"time"

	"gorm.io/gorm"
)

// User merepresentasikan tabel `users` di database.
//
// PENTING (lihat context project §2.1): role "peserta" TIDAK PERNAH punya
// baris User. Peserta event terdaftar lewat mekanisme guest registration
// terpisah (di luar scope EVG-41), bukan lewat tabel ini. Field RoleID di
// sini hanya dipakai untuk role yang punya akun login: super_admin,
// admin_panitia, staf_lapangan, (opsional) school_reviewer.
type User struct {
	ID          uint           `gorm:"primaryKey;column:user_id;autoIncrement" json:"id"`
	RoleID      uint           `gorm:"not null;index" json:"role_id"`
	Role        Role           `json:"role"`
	Username    string         `gorm:"uniqueIndex;size:100;not null" json:"username"`
	Email       string         `gorm:"uniqueIndex;size:150;not null" json:"email"`
	Password    string         `gorm:"size:255;not null" json:"-"`
	// PENTING: sengaja TIDAK pakai tag `default:true` di sini. GORM punya
	// perilaku: untuk field dengan Go zero-value (bool false, int 0, dst)
	// yang juga punya tag `default`, GORM akan mengira field itu "belum
	// diisi" saat Create() dan malah memakai default value dari DB — jadi
	// User{IsActive: false} akan ke-insert sebagai true. Karena IsActive
	// SELALU di-set eksplisit di kode Go (lihat handlers & seeding), default
	// di level DB tidak dibutuhkan dan justru berbahaya di sini.
	IsActive bool `gorm:"not null" json:"is_active"`
	LastLoginAt *time.Time     `json:"last_login_at,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string { return "users" }
