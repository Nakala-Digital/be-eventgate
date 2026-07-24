package auth

import "testing"

func TestHashPassword_And_CheckPassword(t *testing.T) {
	plain := "SuperSecret123!"

	hash, err := HashPassword(plain)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == plain {
		t.Fatal("hash should not equal the plain text password")
	}

	if !CheckPassword(plain, hash) {
		t.Fatal("CheckPassword should return true for the correct password")
	}

	if CheckPassword("WrongPassword", hash) {
		t.Fatal("CheckPassword should return false for an incorrect password")
	}
}
