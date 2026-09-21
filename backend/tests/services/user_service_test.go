package services_test

import (
	"testing"

	"server-monitoring/internal/dtos"
	apperrors "server-monitoring/pkg/errors"
	setup_test "server-monitoring/tests/setup"

	"golang.org/x/crypto/bcrypt"
)

func TestCreateUser_HashesAndPersists(t *testing.T) {
	svc := setup_test.NewUserServiceWithRepo(newFakeUserRepo())

	user, err := svc.CreateUser(dtos.CreateUserDTO{
		Email:    "alice@example.com",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if user.Email != "alice@example.com" {
		t.Fatalf("email = %q", user.Email)
	}
	if user.Password == "secret123" {
		t.Fatal("expected password to be hashed, got plaintext")
	}
	if user.ID.IsZero() {
		t.Fatal("expected generated ID")
	}
}

func TestCreateUser_RejectsDuplicateEmail(t *testing.T) {
	repo := newFakeUserRepo()
	svc := setup_test.NewUserServiceWithRepo(repo)

	if _, err := svc.CreateUser(dtos.CreateUserDTO{Email: "a@b.c", Password: "longenough"}); err != nil {
		t.Fatal(err)
	}
	_, err := svc.CreateUser(dtos.CreateUserDTO{Email: "a@b.c", Password: "longenough"})
	if err != apperrors.ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestGenerateHash_VerifiesAgainstBcrypt(t *testing.T) {
	svc := setup_test.NewUserServiceWithRepo(newFakeUserRepo())

	h, err := svc.GenerateHash("hunter2")
	if err != nil {
		t.Fatal(err)
	}
	if h == nil || *h == "hunter2" {
		t.Fatalf("hash = %v", h)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*h), []byte("hunter2")); err != nil {
		t.Fatalf("verify failed: %v", err)
	}
}

func TestUpdateUserByID_HashesPasswordWhenProvided(t *testing.T) {
	repo := newFakeUserRepo()
	svc := setup_test.NewUserServiceWithRepo(repo)

	user, err := svc.CreateUser(dtos.CreateUserDTO{Email: "x@y.z", Password: "longenough"})
	if err != nil {
		t.Fatal(err)
	}

	newEmail := "x2@y.z"
	newPwd := "rotred123"
	updated, err := svc.UpdateUserByID(user.ID.Hex(), dtos.UpdateUserDTO{
		Email:    &newEmail,
		Password: &newPwd,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Email != newEmail {
		t.Fatalf("email = %q", updated.Email)
	}
	if updated.Password == "rotred123" {
		t.Fatal("password not hashed")
	}
}

func TestFindUserByID_InvalidID(t *testing.T) {
	svc := setup_test.NewUserServiceWithRepo(newFakeUserRepo())
	if _, err := svc.FindUserByID("not-an-objectid"); err == nil {
		t.Fatal("expected error for invalid id")
	}
}
