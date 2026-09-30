package dtos_test

import (
	"testing"
	"time"

	"server-monitoring/internal/dtos"
	"server-monitoring/internal/models"
)

func TestToSnapshotDTO_SentinelOmitsMetrics(t *testing.T) {
	snap := models.ServiceSnapshot{
		Ts:   time.Now().UTC(),
		Meta: models.SnapshotMeta{ServiceID: models.ProviderStatusSentinelID, Provider: "pm2"},
	}
	dto := dtos.ToSnapshotDTO(snap)
	if dto.CPU != nil || dto.Memory != nil || dto.Disk != nil {
		t.Fatalf("sentinel should not carry metric fields: %+v", dto)
	}
	if dto.Status != "" {
		t.Fatalf("status = %q, want empty", dto.Status)
	}
}

func TestToSnapshotDTO_NonZeroMetricsIncluded(t *testing.T) {
	cpu := 12.5
	mem := 67.0
	snap := models.ServiceSnapshot{
		Ts:      time.Now().UTC(),
		Meta:    models.SnapshotMeta{ServiceID: "0", Provider: "pm2", Name: "api"},
		Status:  "online",
		CPU:     cpu,
		Memory:  mem,
		Disk:    1.0,
		NetSent: 100,
		NetRecv: 200,
	}
	dto := dtos.ToSnapshotDTO(snap)
	if dto.CPU == nil || *dto.CPU != cpu {
		t.Fatalf("cpu = %v", dto.CPU)
	}
	if dto.Memory == nil || *dto.Memory != mem {
		t.Fatalf("memory = %v", dto.Memory)
	}
	if dto.Disk == nil {
		t.Fatalf("disk missing")
	}
	if dto.NetSent == nil || *dto.NetSent != 100 {
		t.Fatalf("netSent = %v", dto.NetSent)
	}
}

func TestToSnapshotDTO_ZeroMetricsOmitted(t *testing.T) {
	snap := models.ServiceSnapshot{
		Ts:   time.Now().UTC(),
		Meta: models.SnapshotMeta{ServiceID: "0", Provider: "pm2"},
	}
	dto := dtos.ToSnapshotDTO(snap)
	if dto.CPU != nil || dto.Memory != nil || dto.Disk != nil ||
		dto.NetSent != nil || dto.NetRecv != nil || dto.MemUsage != nil {
		t.Fatalf("zero metrics should be omitted: %+v", dto)
	}
}

func TestToTrackedServiceDTO_RoundTrip(t *testing.T) {
	m := models.SnapshotMeta{
		ServerID:  "local",
		Provider:  "pm2",
		ServiceID: "0",
		Name:      "api",
	}
	dto := dtos.ToTrackedServiceDTO(m)
	if dto.ServerID != "local" || dto.Provider != "pm2" ||
		dto.ServiceID != "0" || dto.Name != "api" {
		t.Fatalf("fields not copied: %+v", dto)
	}
}