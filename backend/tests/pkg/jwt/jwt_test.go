package jwtutil_test

import (
	"testing"
	"time"

	jwtutil "server-monitoring/pkg/jwt"
)

func TestGenerateAndValidateToken_RoundTrip(t *testing.T) {
	tok, err := jwtutil.GenerateToken("user-123")
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	if tok == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := jwtutil.ValidateToken(tok)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims.UserID != "user-123" {
		t.Fatalf("UserID = %q, want %q", claims.UserID, "user-123")
	}
	if claims.ExpiresAt == nil {
		t.Fatal("expected ExpiresAt to be set")
	}
	if time.Until(claims.ExpiresAt.Time) <= 0 {
		t.Fatal("expected token to be in the future")
	}
}

func TestValidateToken_RejectsGarbage(t *testing.T) {
	if _, err := jwtutil.ValidateToken("not-a-jwt"); err == nil {
		t.Fatal("expected error for garbage token")
	}
}

func TestValidateToken_RejectsWrongSignature(t *testing.T) {
	tok, err := jwtutil.GenerateToken("u")
	if err != nil {
		t.Fatal(err)
	}
	// Flip a character in the signature segment.
	bad := tok[:len(tok)-2] + "ZZ"
	if _, err := jwtutil.ValidateToken(bad); err == nil {
		t.Fatal("expected signature mismatch error")
	}
}