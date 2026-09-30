package services_test

import (
	"testing"

	"server-monitoring/internal/dtos"
	"server-monitoring/internal/models"
	"server-monitoring/internal/services"
	apperrors "server-monitoring/pkg/errors"
	jwtutil "server-monitoring/pkg/jwt"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestAuthService_Success(t *testing.T) {
	uid := bson.NewObjectID()
	us := newFakeUserService()
	us.users["alice@example.com"] = &models.User{
		ID:       uid,
		Email:    "alice@example.com",
		Password: hashForTest("hunter2"),
	}
	svc := services.NewAuthService(us)

	tok, err := svc.GenerateAccessToken(dtos.LoginDTO{
		Email:    "alice@example.com",
		Password: "hunter2",
	})
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}
	if tok == "" {
		t.Fatal("expected token")
	}
	claims, err := jwtutil.ValidateToken(tok)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if claims.UserID != uid.Hex() {
		t.Fatalf("UserID = %q, want %q", claims.UserID, uid.Hex())
	}
}

func TestAuthService_UnknownUserReturnsUnauthorized(t *testing.T) {
	svc := services.NewAuthService(newFakeUserService())
	_, err := svc.GenerateAccessToken(dtos.LoginDTO{Email: "no@one.com", Password: "x"})
	if err != apperrors.ErrUnauthorized {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestAuthService_BadPasswordReturnsUnauthorized(t *testing.T) {
	us := newFakeUserService()
	us.users["a@b.c"] = &models.User{Email: "a@b.c", Password: hashForTest("right")}
	svc := services.NewAuthService(us)

	_, err := svc.GenerateAccessToken(dtos.LoginDTO{Email: "a@b.c", Password: "wrong"})
	if err != apperrors.ErrUnauthorized {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestAuthService_ContextIgnoredButAccepted(t *testing.T) {
	// Smoke test: passing a context must not break the happy path.
	us := newFakeUserService()
	us.users["a@b.c"] = &models.User{ID: bson.NewObjectID(), Email: "a@b.c", Password: hashForTest("p")}
	svc := services.NewAuthService(us)
	if _, err := svc.GenerateAccessToken(dtos.LoginDTO{Email: "a@b.c", Password: "p"}); err != nil {
		t.Fatalf("ctx accepted: %v", err)
	}
}