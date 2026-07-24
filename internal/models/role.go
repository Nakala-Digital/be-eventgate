package models

// Role mendefinisikan skema entitas peran dalam basis data.
type Role struct {
	ID          uint      `gorm:"primaryKey;column:role_id;autoIncrement" json:"id"`
	RoleName    string    `gorm:"uniqueIndex;size:50;not null" json:"role_name"`
	Description string    `gorm:"size:255" json:"description,omitempty"`
}

func (Role) TableName() string { return "roles" }

// Kumpulan konstanta peran pengguna. Direkomendasikan untuk senantiasa 
// menggunakan referensi ini demi menjaga integritas dan konsistensi data.
const (
	RoleSuperAdmin     = "super_admin"
	RoleAdminPanitia   = "admin_panitia"
	RoleStafLapangan   = "staf_lapangan"
	RolePeserta        = "peserta"
	RoleSchoolReviewer = "school_reviewer"
)
