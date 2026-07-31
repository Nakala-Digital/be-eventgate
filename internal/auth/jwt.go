package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken dikembalikan untuk semua kondisi token tidak valid: salah
// signature, sudah kedaluwarsa, atau format rusak. Sengaja tidak dibedakan
// detailnya ke caller supaya tidak membocorkan informasi ke client.
var ErrInvalidToken = errors.New("invalid or expired token")

// Claims adalah payload custom yang disisipkan ke dalam JWT.
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	RoleName string `json:"role_name"`
	jwt.RegisteredClaims
}

// GenerateToken membuat JWT bertanda tangan HS256 yang berisi identitas user
// dan role-nya. expiryHours menentukan masa berlaku token.
func GenerateToken(secret string, expiryHours int, userID uint, username, roleName string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		RoleName: roleName,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expiryHours) * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken memvalidasi signature & masa berlaku token, lalu mengembalikan
// dan mengembalikan objek claims apabila token dinyatakan valid.
func ParseToken(secret, tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
