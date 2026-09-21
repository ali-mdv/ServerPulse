package dtos_test

import (
	"testing"
	"time"

	"server-monitoring/internal/dtos"
	"server-monitoring/internal/models"
)

func TestToServerUsageDTO(t *testing.T) {
	in := models.ServerUsageState{
		ServerID:  "abc",
		UpdatedAt: time.Now().UTC(),
		Usage:     models.SystemUsage{CpuUsage: 0.5},
	}
	got := dtos.ToServerUsageDTO(in)
	if got == nil {
		t.Fatal("nil dto")
	}
	if got.ServerID != "abc" {
		t.Fatalf("serverID = %q", got.ServerID)
	}
	if got.Usage.CpuUsage != 0.5 {
		t.Fatalf("usage not copied: %+v", got.Usage)
	}
}

func TestToServerUsageDTOs(t *testing.T) {
	in := []models.ServerUsageState{
		{ServerID: "a"},
		{ServerID: "b"},
	}
	out := dtos.ToServerUsageDTOs(in)
	if len(out) != 2 {
		t.Fatalf("len = %d", len(out))
	}
	if out[0].ServerID != "a" || out[1].ServerID != "b" {
		t.Fatalf("mapping wrong: %+v", out)
	}
}

func TestToProviderStateDTO_NestsSnapshots(t *testing.T) {
	now := time.Now().UTC()
	p := models.ProviderState{
		ServerID:  "abc",
		Provider:  "pm2",
		UpdatedAt: now,
		Available: true,
		Services: []models.ServiceSnapshot{
			{Ts: now, Meta: models.SnapshotMeta{ServiceID: "0", Provider: "pm2"}, CPU: 1.0},
		},
	}
	dto := dtos.ToProviderStateDTO(p)
	if dto.Provider != "pm2" || !dto.Available {
		t.Fatalf("bad dto: %+v", dto)
	}
	if len(dto.Services) != 1 {
		t.Fatalf("services len = %d", len(dto.Services))
	}
}

func TestToProviderStateDTOs(t *testing.T) {
	in := []models.ProviderState{{Provider: "pm2"}, {Provider: "docker"}}
	out := dtos.ToProviderStateDTOs(in)
	if len(out) != 2 || out[0].Provider != "pm2" || out[1].Provider != "docker" {
		t.Fatalf("bad slice: %+v", out)
	}
}