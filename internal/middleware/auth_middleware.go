package middleware

import (
	"context"
	"net/http"
	"strings"

	"be-eventgate/internal/auth"
	"be-eventgate/internal/httpx"
)

// RequireAuth bertugas memvalidasi token JWT pada header Authorization HTTP.
// Jika valid, data pengguna akan diinjeksi ke dalam konteks permintaan
// untuk digunakan oleh handler maupun middleware otorisasi selanjutnya.
func RequireAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				httpx.WriteError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				httpx.WriteError(w, http.StatusUnauthorized, "invalid authorization header format, expected: Bearer <token>")
				return
			}

			claims, err := auth.ParseToken(jwtSecret, parts[1])
			if err != nil {
				httpx.WriteError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), ContextUserID, claims.UserID)
			ctx = context.WithValue(ctx, ContextUsername, claims.Username)
			ctx = context.WithValue(ctx, ContextRoleName, claims.RoleName)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
