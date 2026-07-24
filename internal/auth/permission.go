package auth

import "be-eventgate/internal/models"

// Permission adalah struktur awal untuk sistem otorisasi berbasis
// resource+action, disiapkan sebagai fondasi (EVG-41 acceptance criteria:
// "struktur permission awal"). Task berikutnya (mis. EVG-45/47/49) tinggal
// menambahkan entri baru ke permissionMatrix tanpa perlu mengubah middleware.
//
// Untuk kebutuhan RBAC dasar EVG-41 sendiri, middleware.RequireRole (role-only)
// sudah cukup. Fungsi IsAllowed di sini untuk dipakai kalau nanti butuh
// kontrol yang lebih granular per resource+action.
type Permission struct {
	Resource string
	Action   string
}

var permissionMatrix = map[Permission][]string{
	{Resource: "event", Action: "approve"}: {models.RoleSuperAdmin},
	{Resource: "event", Action: "publish"}: {models.RoleSuperAdmin},
	{Resource: "event", Action: "create"}:  {models.RoleAdminPanitia},
	{Resource: "user", Action: "read_self"}: {
		models.RoleSuperAdmin, models.RoleAdminPanitia, models.RoleStafLapangan,
	},
}

// IsAllowed mengecek apakah sebuah role boleh melakukan action tertentu atas
// sebuah resource, berdasarkan permissionMatrix di atas.
func IsAllowed(roleName, resource, action string) bool {
	allowedRoles, ok := permissionMatrix[Permission{Resource: resource, Action: action}]
	if !ok {
		return false
	}
	for _, r := range allowedRoles {
		if r == roleName {
			return true
		}
	}
	return false
}
