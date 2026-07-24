package middleware

import (
	"net/http"

	"be-eventgate/internal/httpx"
)

// RequireRole berfungsi membatasi akses rute hanya bagi peran (role) yang diizinkan.
// Middleware ini harus diimplementasikan setelah RequireAuth
// guna mengevaluasi identitas pengguna yang tersimpan di dalam konteks.
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
