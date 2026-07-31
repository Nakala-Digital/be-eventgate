package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword melakukan enkripsi satu arah pada kata sandi menggunakan algoritma bcrypt.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword memvalidasi kecocokan antara kata sandi teks dengan hash yang tersimpan di basis data.
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
