package services_test

import (
	"context"
	"testing"
	"time"

	"server-monitoring/internal/models"
	"server-monitoring/internal/services"
	setup_test "server-monitoring/tests/setup"
)

func TestRecordPM2Snapshot_BuildsCorrectShape(t *testing.T) {
	repo := &fakeHistoryRepo{}
	svc := setup_test.NewHistoryServiceWithRepo(repo, time.Hour)

	procs := []models.PM2Process{
		{Name: "api", PMID: 0, PM2Env: models.PM2Env{Status: "online"}, Monit: models.PM2Monit{CPU: 1.5, Memory: 1024}},
		{Name: "worker", PMID: 1, PM2Env: models.PM2Env{Status: "stopped"}, Monit: models.PM2Monit{CPU: 0, Memory: 0}},
	}
	if err := svc.RecordPM2Snapshot(context.Background(), "host-7", procs); err != nil {
		t.Fatal(err)
	}
	if len(repo.batches) != 1 {
		t.Fatalf("batches = %d", len(repo.batches))
	}
	got := repo.batches[0]
	if len(got) != 2 {
		t.Fatalf("snapshots = %d", len(got))
	}
	if got[0].Meta.ServiceID != "0" || got[0].Meta.Provider != "pm2" ||
		got[0].Meta.ServerID != "host-7" || got[0].Meta.Name != "api" {
		t.Fatalf("bad meta: %+v", got[0].Meta)
	}
	if got[0].Status != "online" || !got[0].Available {
		t.Fatalf("bad status/available: %+v", got[0])
	}
	if got[0].CPU != 1.5 || got[0].Memory != 1024 {
		t.Fatalf("monit not copied: %+v", got[0])
	}
	if got[1].Status != "stopped" || got[1].Available {
		t.Fatalf("second snapshot: %+v", got[1])
	}
}

func TestRecordDockerSnapshot_BuildsCorrectShape(t *testing.T) {
	repo := &fakeHistoryRepo{}
	svc := setup_test.NewHistoryServiceWithRepo(repo, time.Hour)

	containers := []models.DockerContainer{
		{ID: "c1", Name: "web", State: "running", Usage: models.DockerContainerUsage{CpuPercent: 12, MemPercent: 34, MemUsage: 5678}},
		{ID: "c2", Name: "db", State: "exited"},
	}
	if err := svc.RecordDockerSnapshot(context.Background(), "host-7", containers); err != nil {
		t.Fatal(err)
	}
	got := repo.batches[0]
	if len(got) != 2 {
		t.Fatalf("snapshots = %d", len(got))
	}
	if got[0].Meta.Provider != "docker" || got[0].Meta.ServerID != "host-7" {
		t.Fatalf("bad meta: %+v", got[0].Meta)
	}
	if got[0].Status != "running" || !got[0].Available {
		t.Fatalf("bad running snapshot: %+v", got[0])
	}
	if got[0].CPU != 12 || got[0].Memory != 34 || got[0].MemUsage != 5678 {
		t.Fatalf("usage not copied: %+v", got[0])
	}
	if got[1].Status != "exited" || got[1].Available {
		t.Fatalf("bad exited snapshot: %+v", got[1])
	}
}

func TestRecordSystemSnapshot_ProducesFiveMetrics(t *testing.T) {
	repo := &fakeHistoryRepo{}
	svc := setup_test.NewHistoryServiceWithRepo(repo, time.Hour)

	usage := models.SystemUsage{
		CpuUsage:  10,
		MemUsage:  models.Usage{UsedPercent: 50, Used: 1024},
		DiskUsage: models.Usage{UsedPercent: 70},
		NetIO:     models.NetworkUsage{Send: 100, Received: 200},
	}
	if err := svc.RecordSystemSnapshot(context.Background(), "host-7", usage); err != nil {
		t.Fatal(err)
	}
	got := repo.batches[0]
	if len(got) != 5 {
		t.Fatalf("snapshots = %d, want 5", len(got))
	}
	wantNames := []string{"CPU", "Memory", "Disk", "Network Sent", "Network Received"}
	for i, snap := range got {
		if snap.Meta.Provider != "system" {
			t.Fatalf("snapshot %d: provider = %q", i, snap.Meta.Provider)
		}
		if snap.Meta.Name != wantNames[i] {
			t.Fatalf("snapshot %d: name = %q, want %q", i, snap.Meta.Name, wantNames[i])
		}
	}
	if got[0].CPU != 10 {
		t.Fatalf("CPU snapshot cpu = %v", got[0].CPU)
	}
	if got[1].Memory != 50 || got[1].MemUsage != 1024 {
		t.Fatalf("Memory snapshot = %+v", got[1])
	}
	if got[2].Disk != 70 {
		t.Fatalf("Disk snapshot disk = %v", got[2].Disk)
	}
	if got[3].NetSent != 100 || got[4].NetRecv != 200 {
		t.Fatalf("Net snapshots wrong: %+v %+v", got[3], got[4])
	}
}

func TestRecordProviderUnavailable_Sentinel(t *testing.T) {
	repo := &fakeHistoryRepo{}
	svc := setup_test.NewHistoryServiceWithRepo(repo, time.Hour)

	if err := svc.RecordProviderUnavailable(context.Background(), "host-7", "pm2"); err != nil {
		t.Fatal(err)
	}
	got := repo.batches[0]
	if len(got) != 1 {
		t.Fatalf("snapshots = %d", len(got))
	}
	snap := got[0]
	if snap.Meta.ServiceID != models.ProviderStatusSentinelID {
		t.Fatalf("sentinel ID = %q", snap.Meta.ServiceID)
	}
	if snap.Meta.Name != models.ProviderStatusSentinelName {
		t.Fatalf("sentinel name = %q", snap.Meta.Name)
	}
	if snap.Available {
		t.Fatalf("sentinel should be unavailable")
	}
}

func TestInsertHistoryBatch_PassThrough(t *testing.T) {
	repo := &fakeHistoryRepo{}
	svc := setup_test.NewHistoryServiceWithRepo(repo, time.Hour)

	batch := []models.ServiceSnapshot{
		{Meta: models.SnapshotMeta{ServiceID: "0", Provider: "pm2", ServerID: "host-7"}},
	}
	if err := svc.InsertHistoryBatch(context.Background(), batch); err != nil {
		t.Fatal(err)
	}
	if len(repo.batches) != 1 || len(repo.batches[0]) != 1 {
		t.Fatalf("batches = %+v", repo.batches)
	}
}

func TestNewHistoryService_RejectsZeroRetention(t *testing.T) {
	if _, err := services.NewHistoryService("ignored", 0); err == nil {
		t.Fatal("expected error for zero retention")
	}
}

func TestFindSeries_NegativeBucketClampsToZero(t *testing.T) {
	repo := &fakeHistoryRepo{}
	svc := setup_test.NewHistoryServiceWithRepo(repo, time.Hour)

	if _, err := svc.FindSeries(context.Background(), "h", "pm2", "0", time.Now(), time.Now(), -time.Second); err != nil {
		t.Fatalf("FindSeries: %v", err)
	}
}
