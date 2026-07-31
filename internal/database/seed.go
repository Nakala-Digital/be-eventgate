package database

import (
	"gorm.io/gorm"

	"be-eventgate/internal/models"
)

// SeedRoles menginjeksi peran (roles) sistem ke dalam basis data.
// Operasi ini bersifat idempoten dan aman dieksekusi berulang kali.
// Catatan: Peran 'peserta' secara arsitektural tidak memiliki akun log masuk
// dan hanya digunakan untuk validasi relasional pada pendaftaran tamu.
func SeedRoles(db *gorm.DB) error {
	roles := []models.Role{
		{RoleName: models.RoleSuperAdmin, Description: "Pemilik sistem: approve/reject/publish event, kelola semua user"},
		{RoleName: models.RoleAdminPanitia, Description: "Mengelola event miliknya sendiri"},
		{RoleName: models.RoleStafLapangan, Description: "Scan QR ticket untuk proses check-in di lapangan"},
		{RoleName: models.RolePeserta, Description: "Peserta event — tidak punya akun login, terdaftar via guest registration"},
		{RoleName: models.RoleSchoolReviewer, Description: "Opsional: role tambahan untuk memisahkan approval dari super_admin"},
	}

	for _, r := range roles {
		var existing models.Role
		err := db.Where("role_name = ?", r.RoleName).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&r).Error; err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
		// Peran sudah eksis, proses dilewati untuk menjaga sifat idempoten.
	}
	return nil
}
