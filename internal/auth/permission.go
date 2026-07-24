package auth

import "be-eventgate/internal/models"

// Permission mendefinisikan struktur dasar untuk sistem otorisasi berbasis
// sumber daya dan tindakan. Struktur ini diimplementasikan sebagai fondasi,
// sehingga penambahan izin di masa mendatang dapat dilakukan langsung pada
// matriks tanpa memodifikasi middleware inti.
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

// IsAllowed memverifikasi hak akses suatu peran (role) terhadap tindakan
// spesifik pada sumber daya tertentu berdasarkan matriks perizinan.
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
