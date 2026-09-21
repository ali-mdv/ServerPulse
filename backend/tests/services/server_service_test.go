package services_test

import (
	"context"
	"testing"

	"server-monitoring/internal/dtos"
	"server-monitoring/internal/services"
	apperrors "server-monitoring/pkg/errors"
	setup_test "server-monitoring/tests/setup"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestCreateServer_DoesNotExposeToken(t *testing.T) {
	repo := newFakeServerRepo()
	svc := setup_test.NewServerServiceWithRepo(repo)

	created, err := svc.Create(context.Background(), dtos.CreateServerDTO{
		Name:        "web-01",
		Description: "frontend host",
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	if created.AgentToken != "" {
		t.Fatalf("expected no token on create, got %q", created.AgentToken)
	}
	if created.Name != "web-01" {
		t.Fatalf("expected name web-01, got %q", created.Name)
	}
	if created.Description != "frontend host" {
		t.Fatalf("expected description, got %q", created.Description)
	}
}

func TestGenerateAgentToken_CreatesAndRotatesKey(t *testing.T) {
	repo := newFakeServerRepo()
	svc := setup_test.NewServerServiceWithRepo(repo)

	created, err := svc.Create(context.Background(), dtos.CreateServerDTO{Name: "db-01"})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// First key generation.
	_, key1, err := svc.GenerateAgentToken(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("first token generation failed: %v", err)
	}
	if key1 == "" {
		t.Fatal("expected non-empty token")
	}

	resolved, err := svc.ResolveByAgentToken(context.Background(), key1)
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if resolved == nil || resolved.ID != created.ID {
		t.Fatal("token did not resolve to the created server")
	}

	// Rotation invalidates the previous key.
	_, key2, err := svc.GenerateAgentToken(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("second token generation failed: %v", err)
	}
	if key1 == key2 {
		t.Fatal("rotated token should differ from the previous one")
	}

	resolvedOld, _ := svc.ResolveByAgentToken(context.Background(), key1)
	if resolvedOld != nil {
		t.Fatal("old token should be invalid after rotation")
	}

	resolvedNew, _ := svc.ResolveByAgentToken(context.Background(), key2)
	if resolvedNew == nil || resolvedNew.ID != created.ID {
		t.Fatal("new token did not resolve to the created server")
	}
}

func TestGenerateAgentToken_UnknownServer(t *testing.T) {
	repo := newFakeServerRepo()
	svc := setup_test.NewServerServiceWithRepo(repo)

	_, _, err := svc.GenerateAgentToken(context.Background(), bson.NewObjectID())
	if err != services.ErrServerNotFound {
		t.Fatalf("expected ErrServerNotFound, got %v", err)
	}
}

func TestDelete_LocalServerIsProtected(t *testing.T) {
	repo := newFakeServerRepo()
	svc := setup_test.NewServerServiceWithRepo(repo)

	local, err := svc.EnsureLocalServer(context.Background())
	if err != nil {
		t.Fatalf("ensure local server failed: %v", err)
	}

	err = svc.Delete(context.Background(), local.ID)
	if err != apperrors.ErrForbidden {
		t.Fatalf("expected ErrForbidden when deleting local server, got %v", err)
	}
}
