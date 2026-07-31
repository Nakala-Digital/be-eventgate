package middleware

import "context"

type contextKey string

// Kunci-kunci konteks ini difungsikan untuk mendistribusikan data pengguna 
// hasil validasi JWT secara aman ke seluruh lapisan siklus permintaan.
const (
	ContextUserID   contextKey = "user_id"
	ContextUsername contextKey = "username"
	ContextRoleName contextKey = "role_name"
)

// GetUserID mengekstraksi ID pengguna dari konteks yang aktif.
// Mengembalikan nilai boolean false jika konteks tidak terautentikasi.
func GetUserID(ctx context.Context) (uint, bool) {
	v, ok := ctx.Value(ContextUserID).(uint)
	return v, ok
}

// GetUsername mengekstraksi nama pengguna dari konteks yang aktif.
func GetUsername(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ContextUsername).(string)
	return v, ok
}

// GetRoleName mengekstraksi nama peran pengguna dari konteks yang aktif.
func GetRoleName(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ContextRoleName).(string)
	return v, ok
}
