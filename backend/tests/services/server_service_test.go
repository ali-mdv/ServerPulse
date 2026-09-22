package services_test

import (
	"context"
	"testing"
	"time"

	"server-monitoring/internal/dtos"
	"server-monitoring/internal/models"
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

func TestMarkStaleDown_MarksOnlyStaleAndUnseen(t *testing.T) {
	repo := newFakeServerRepo()
	svc := setup_test.NewServerServiceWithRepo(repo)

	old := time.Now().UTC().Add(-2 * time.Hour)
	fresh := time.Now().UTC()
	repo.servers = []models.Server{
		{ID: bson.NewObjectID(), Name: "stale", Status: models.ServerStatusOnline, LastSeen: &old},
		{ID: bson.NewObjectID(), Name: "fresh", Status: models.ServerStatusOnline, LastSeen: &fresh},
		{ID: bson.NewObjectID(), Name: "already-down", Status: models.ServerStatusDown, LastSeen: &old},
		{ID: bson.NewObjectID(), Name: "never-seen", Status: models.ServerStatusUnknown},
	}

	down, err := svc.MarkStaleDown(context.Background(), time.Minute)
	if err != nil {
		t.Fatalf("MarkStaleDown: %v", err)
	}
	if len(down) != 2 {
		t.Fatalf("marked down = %d, want 2", len(down))
	}

	status := map[string]models.ServerStatus{}
	for _, s := range repo.servers {
		status[s.Name] = s.Status
	}
	if status["stale"] != models.ServerStatusDown {
		t.Fatalf("stale status = %q, want down", status["stale"])
	}
	if status["never-seen"] != models.ServerStatusDown {
		t.Fatalf("never-seen status = %q, want down", status["never-seen"])
	}
	if status["fresh"] != models.ServerStatusOnline {
		t.Fatalf("fresh status = %q, want online", status["fresh"])
	}
	if status["already-down"] != models.ServerStatusDown {
		t.Fatalf("already-down status = %q", status["already-down"])
	}

	// Idempotent: a second sweep finds nothing new.
	again, err := svc.MarkStaleDown(context.Background(), time.Minute)
	if err != nil {
		t.Fatalf("second MarkStaleDown: %v", err)
	}
	if len(again) != 0 {
		t.Fatalf("second sweep marked %d, want 0", len(again))
	}
}
