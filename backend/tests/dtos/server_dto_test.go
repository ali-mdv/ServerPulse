package dtos_test

import (
	"testing"
	"time"

	"server-monitoring/internal/dtos"
	"server-monitoring/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestToServerDTO_StripsAgentToken(t *testing.T) {
	id := bson.NewObjectID()
	now := time.Now().UTC()
	srv := models.Server{
		ID:          id,
		Name:        "web-01",
		Host:        "10.0.0.1",
		Port:        22,
		Description: "frontend",
		Status:      models.ServerStatusOnline,
		AgentToken:  "supersecret",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	dto := dtos.ToServerDTO(srv, nil, nil)
	if dto.ID != id.Hex() {
		t.Fatalf("id = %q, want %q", dto.ID, id.Hex())
	}
	if dto.Name != "web-01" || dto.Host != "10.0.0.1" || dto.Port != 22 {
		t.Fatalf("fields not copied: %+v", dto)
	}
	if dto.Description != "frontend" {
		t.Fatalf("description = %q", dto.Description)
	}
}

func TestToServerDTO_AttachesUsageAndProviders(t *testing.T) {
	id := bson.NewObjectID()
	srv := models.Server{ID: id, Name: "n", Status: models.ServerStatusOnline, CreatedAt: time.Now(), UpdatedAt: time.Now()}

	usage := &models.ServerUsageState{
		ServerID:  id.Hex(),
		UpdatedAt: time.Now().UTC(),
		Usage:     models.SystemUsage{CpuUsage: 1.5},
	}
	providers := map[string]models.ProviderState{
		models.ProviderPM2: {ServerID: id.Hex(), Provider: models.ProviderPM2, Available: true, UpdatedAt: time.Now().UTC()},
	}

	dto := dtos.ToServerDTO(srv, usage, providers)
	if dto.Usage == nil || dto.Usage.Usage.CpuUsage != 1.5 {
		t.Fatalf("usage not attached: %+v", dto.Usage)
	}
	if dto.Providers == nil {
		t.Fatal("providers not attached")
	}
	if dto.Providers[models.ProviderPM2].Available != true {
		t.Fatal("provider availability wrong")
	}
}

func TestToServerProviderAvailabilityDTOs_EmptyReturnsNil(t *testing.T) {
	if got := dtos.ToServerProviderAvailabilityDTOs(map[string]models.ProviderState{}); got != nil {
		t.Fatalf("expected nil for empty map, got %+v", got)
	}
}

func TestToServerProviderAvailabilityDTOs_DropsZeroUpdatedAt(t *testing.T) {
	got := dtos.ToServerProviderAvailabilityDTOs(map[string]models.ProviderState{
		"pm2": {Available: true},
	})
	if got["pm2"].UpdatedAt != nil {
		t.Fatalf("expected nil UpdatedAt for zero time, got %v", got["pm2"].UpdatedAt)
	}
}