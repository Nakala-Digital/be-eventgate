package auth

import "testing"

func TestGenerateAndParseToken(t *testing.T) {
	secret := "test-secret"

	token, err := GenerateToken(secret, 1, 42, "johndoe", "admin_panitia")
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	claims, err := ParseToken(secret, token)
	if err != nil {
		t.Fatalf("ParseToken returned error: %v", err)
	}

	if claims.UserID != 42 {
		t.Errorf("expected user_id 42, got %d", claims.UserID)
	}
	if claims.Username != "johndoe" {
		t.Errorf("expected username johndoe, got %s", claims.Username)
	}
	if claims.RoleName != "admin_panitia" {
		t.Errorf("expected role_name admin_panitia, got %s", claims.RoleName)
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	token, err := GenerateToken("secret-a", 1, 1, "user", "staf_lapangan")
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	if _, err := ParseToken("secret-b", token); err == nil {
		t.Fatal("expected an error when parsing a token signed with a different secret")
	}
}

func TestParseToken_Expired(t *testing.T) {
	// expiryHours negatif -> token sudah kedaluwarsa sejak dibuat.
	token, err := GenerateToken("secret", -1, 1, "user", "super_admin")
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	if _, err := ParseToken("secret", token); err == nil {
		t.Fatal("expected an error when parsing an expired token")
	}
}

func TestParseToken_GarbageString(t *testing.T) {
	if _, err := ParseToken("secret", "not-a-real-jwt-token"); err == nil {
		t.Fatal("expected an error when parsing a malformed token string")
	}
}
