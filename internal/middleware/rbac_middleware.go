package middleware

import (
	"net/http"

	"be-eventgate/internal/httpx"
)

// RequireRole membatasi akses route hanya untuk role yang disebutkan.
// WAJIB dipasang SETELAH RequireAuth di chain middleware (butuh role_name
// yang sudah disisipkan RequireAuth ke context). Kalau role tidak sesuai
// -> 403 Forbidden. Kalau context kosong (berarti RequireAuth belum jalan,
// salah urutan pemasangan middleware) -> 401 Unauthorized.
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(allowedRoles))
	for _, r := range allowedRoles {
		allowed[r] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roleName, ok := GetRoleName(r.Context())
			if !ok || roleName == "" {
				httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			if !allowed[roleName] {
				httpx.WriteError(w, http.StatusForbidden, "you don't have permission to access this resource")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
