package models

// Role merepresentasikan tabel `roles` di database.
type Role struct {
	ID          uint      `gorm:"primaryKey;column:role_id;autoIncrement" json:"id"`
	RoleName    string    `gorm:"uniqueIndex;size:50;not null" json:"role_name"`
	Description string    `gorm:"size:255" json:"description,omitempty"`
}

func (Role) TableName() string { return "roles" }

// Konstanta nama role. JANGAN hardcode string literal role di tempat lain,
// selalu pakai konstanta ini supaya konsisten di seluruh codebase.
const (
	RoleSuperAdmin     = "super_admin"
	RoleAdminPanitia   = "admin_panitia"
	RoleStafLapangan   = "staf_lapangan"
	RolePeserta        = "peserta"
	RoleSchoolReviewer = "school_reviewer"
)
