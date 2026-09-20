package services

import (
	"context"
	"testing"
	"time"

	"server-monitoring/internal/dtos"
	"server-monitoring/internal/models"
	apperrors "server-monitoring/pkg/errors"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// fakeServerRepo is an in-memory stand-in for repository.ServerRepository
// focused on the agent-token workflow.
type fakeServerRepo struct {
	servers []models.Server
}

func (r *fakeServerRepo) Insert(_ context.Context, s models.Server) (*models.Server, error) {
	if s.ID.IsZero() {
		s.ID = bson.NewObjectID()
	}
	s.CreatedAt = time.Now().UTC()
	s.UpdatedAt = s.CreatedAt
	r.servers = append(r.servers, s)
	return &s, nil
}

func (r *fakeServerRepo) FindByID(_ context.Context, id bson.ObjectID) (*models.Server, error) {
	for i := range r.servers {
		if r.servers[i].ID == id {
			return &r.servers[i], nil
		}
	}
	return nil, nil
}

func (r *fakeServerRepo) FindByName(_ context.Context, name string) (*models.Server, error) {
	for i := range r.servers {
		if r.servers[i].Name == name {
			return &r.servers[i], nil
		}
	}
	return nil, nil
}

func (r *fakeServerRepo) FindByAgentToken(_ context.Context, token string) (*models.Server, error) {
	for i := range r.servers {
		if r.servers[i].AgentToken == token {
			return &r.servers[i], nil
		}
	}
	return nil, nil
}

func (r *fakeServerRepo) List(_ context.Context) ([]models.Server, error) {
	return r.servers, nil
}

func (r *fakeServerRepo) UpdateByID(_ context.Context, id bson.ObjectID, update bson.D) (*models.Server, error) {
	for i := range r.servers {
		if r.servers[i].ID != id {
			continue
		}
		// Simplistic $set application sufficient for these tests.
		for _, top := range update {
			if top.Key != "$set" {
				continue
			}
			sets, ok := top.Value.(bson.D)
			if !ok {
				continue
			}
			for _, set := range sets {
				switch set.Key {
				case "name":
					r.servers[i].Name = set.Value.(string)
				case "host":
					r.servers[i].Host = set.Value.(string)
				case "port":
					r.servers[i].Port = set.Value.(int)
				case "description":
					r.servers[i].Description = set.Value.(string)
				case "agentToken":
					r.servers[i].AgentToken = set.Value.(string)
				case "updatedAt":
					r.servers[i].UpdatedAt = set.Value.(time.Time)
				}
			}
		}
		return &r.servers[i], nil
	}
	return nil, nil
}

func (r *fakeServerRepo) DeleteByID(_ context.Context, _ bson.ObjectID) error {
	return nil
}

func (r *fakeServerRepo) TouchSeen(_ context.Context, _ bson.ObjectID, _ models.ServerStatus) error {
	return nil
}

func TestCreateServer_DoesNotExposeToken(t *testing.T) {
	repo := &fakeServerRepo{}
	svc := &serverService{repo: repo}

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
	repo := &fakeServerRepo{}
	svc := &serverService{repo: repo}

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
	repo := &fakeServerRepo{}
	svc := &serverService{repo: repo}

	_, _, err := svc.GenerateAgentToken(context.Background(), bson.NewObjectID())
	if err != ErrServerNotFound {
		t.Fatalf("expected ErrServerNotFound, got %v", err)
	}
}

func TestDelete_LocalServerIsProtected(t *testing.T) {
	repo := &fakeServerRepo{}
	svc := &serverService{repo: repo}

	local, err := svc.EnsureLocalServer(context.Background())
	if err != nil {
		t.Fatalf("ensure local server failed: %v", err)
	}

	err = svc.Delete(context.Background(), local.ID)
	if err != apperrors.ErrForbidden {
		t.Fatalf("expected ErrForbidden when deleting local server, got %v", err)
	}
}
