package database

import (
	"gorm.io/gorm"

	"be-eventgate/internal/models"
)

// SeedRoles memastikan role dasar tersedia di database:
// super_admin, admin_panitia, staf_lapangan, peserta, school_reviewer.
// Idempotent — aman dipanggil berkali-kali (misal setiap kali server start),
// tidak akan membuat duplikat.
//
// CATATAN PENTING: "peserta" TIDAK punya akun login (guest registration,
// lihat context project bagian keputusan Login Peserta). Role ini tetap
// dibuat di tabel Role supaya "dikenali sistem" sesuai acceptance criteria
// EVG-41, tapi tidak akan pernah ada baris User yang terhubung ke role ini,
// dan endpoint login akan selalu gagal untuk role ini karena memang tidak
// ada User dengan role tersebut.
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
		// role sudah ada, skip (idempotent)
	}
	return nil
}
