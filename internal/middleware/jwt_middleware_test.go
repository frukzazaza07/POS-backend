package middleware

import (
	"os"
	"testing"
)

func TestGenerateToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key-for-unit-tests")

	token, err := GenerateToken("user-123", "test@example.com", "admin")
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestParseValidToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key-for-unit-tests")

	token, err := GenerateToken("user-abc", "cashier@pos.local", "cashier")
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	claims, err := parseToken(token)
	if err != nil {
		t.Fatalf("parseToken: %v", err)
	}
	if claims.UserID != "user-abc" {
		t.Errorf("want UserID=user-abc, got %s", claims.UserID)
	}
	if claims.Email != "cashier@pos.local" {
		t.Errorf("want Email=cashier@pos.local, got %s", claims.Email)
	}
	if claims.Role != "cashier" {
		t.Errorf("want Role=cashier, got %s", claims.Role)
	}
}

func TestParseInvalidToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key-for-unit-tests")

	_, err := parseToken("not.a.valid.token")
	if err == nil {
		t.Fatal("expected error for malformed token")
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	os.Setenv("JWT_SECRET", "correct-secret")
	token, err := GenerateToken("u1", "a@b.com", "admin")
	if err != nil {
		t.Fatal(err)
	}

	os.Setenv("JWT_SECRET", "wrong-secret")
	_, err = parseToken(token)
	if err == nil {
		t.Fatal("expected error when verifying with wrong secret")
	}
}
