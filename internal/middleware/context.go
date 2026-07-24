package middleware

import "context"

type contextKey string

// Key context yang dipakai untuk menyisipkan data user hasil validasi token
// ke dalam request context oleh RequireAuth, lalu dibaca oleh handler /
// middleware lain (mis. RequireRole).
const (
	ContextUserID   contextKey = "user_id"
	ContextUsername contextKey = "username"
	ContextRoleName contextKey = "role_name"
)

// GetUserID mengambil user_id dari context. ok=false kalau tidak ada
// (artinya request belum melewati RequireAuth).
func GetUserID(ctx context.Context) (uint, bool) {
	v, ok := ctx.Value(ContextUserID).(uint)
	return v, ok
}

// GetUsername mengambil username dari context.
func GetUsername(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ContextUsername).(string)
	return v, ok
}

// GetRoleName mengambil role_name dari context.
func GetRoleName(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ContextRoleName).(string)
	return v, ok
}
