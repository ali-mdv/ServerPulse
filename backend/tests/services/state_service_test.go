package services_test

import (
	"context"
	"testing"
	"time"

	"server-monitoring/internal/models"
	setup_test "server-monitoring/tests/setup"
)

func TestRecordProviderState_NilServicesBecomesEmptySlice(t *testing.T) {
	repo := &fakeStateRepo{}
	svc := setup_test.NewStateServiceWithRepo(repo)

	if err := svc.RecordProviderState(context.Background(), "local", "pm2", false, nil); err != nil {
		t.Fatal(err)
	}
	if len(repo.providerStates) != 1 {
		t.Fatalf("want 1 row, got %d", len(repo.providerStates))
	}
	got := repo.providerStates[0]
	if got.ServerID != "local" || got.Provider != "pm2" || got.Available {
		t.Fatalf("row: %+v", got)
	}
	if got.Services == nil {
		t.Fatal("Services should be non-nil empty slice, got nil")
	}
	if len(got.Services) != 0 {
		t.Fatalf("Services len = %d", len(got.Services))
	}
}

func TestRecordProviderState_StampsServerIDOnSnapshots(t *testing.T) {
	repo := &fakeStateRepo{}
	svc := setup_test.NewStateServiceWithRepo(repo)

	snaps := []models.ServiceSnapshot{
		{Meta: models.SnapshotMeta{ServiceID: "0", Name: "api"}, CPU: 1.0},
		{Meta: models.SnapshotMeta{ServiceID: "1", Name: "worker"}},
	}
	if err := svc.RecordProviderState(context.Background(), "host-7", "pm2", true, snaps); err != nil {
		t.Fatal(err)
	}
	got := repo.providerStates[0]
	for i, s := range got.Services {
		if s.Meta.ServerID != "host-7" {
			t.Fatalf("snapshot %d: serverID = %q, want host-7", i, s.Meta.ServerID)
		}
	}
}

func TestRecordServerUsage_RoundTrip(t *testing.T) {
	repo := &fakeStateRepo{}
	svc := setup_test.NewStateServiceWithRepo(repo)

	usage := models.SystemUsage{CpuUsage: 5.0}
	if err := svc.RecordServerUsage(context.Background(), "host-7", usage); err != nil {
		t.Fatal(err)
	}
	got, err := svc.GetServerUsage(context.Background(), "host-7")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected usage row")
	}
	if got.Usage.CpuUsage != 5.0 {
		t.Fatalf("cpu = %v", got.Usage.CpuUsage)
	}
	if got.UpdatedAt.IsZero() || got.UpdatedAt.After(time.Now()) {
		t.Fatalf("UpdatedAt = %v", got.UpdatedAt)
	}
}

func TestGetProviderState_UnknownReturnsNil(t *testing.T) {
	repo := &fakeStateRepo{}
	svc := setup_test.NewStateServiceWithRepo(repo)

	got, err := svc.GetProviderState(context.Background(), "missing", "pm2")
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}
