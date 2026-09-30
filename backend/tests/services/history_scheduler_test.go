package services_test

import (
	"testing"
	"time"

	"server-monitoring/internal/models"
	setup_test "server-monitoring/tests/setup"

	"github.com/docker/docker/api/types/container"
)

func TestToSnapshots_StampsServerIDAndProvider(t *testing.T) {
	procs := []models.PM2Process{
		{Name: "api", PMID: 0, PM2Env: models.PM2Env{Status: "online"}, Monit: models.PM2Monit{CPU: 5, Memory: 1024}},
		{Name: "worker", PMID: 7, PM2Env: models.PM2Env{Status: "stopped"}},
	}
	got := setup_test.SnapshotsFromPM2Processes("host-1", procs)
	if len(got) != 2 {
		t.Fatalf("len = %d", len(got))
	}
	for i, s := range got {
		if s.Meta.ServerID != "host-1" {
			t.Fatalf("snapshot %d: serverID = %q", i, s.Meta.ServerID)
		}
		if s.Meta.Provider != models.ProviderPM2 {
			t.Fatalf("snapshot %d: provider = %q", i, s.Meta.Provider)
		}
		if s.Meta.ServiceID == "" {
			t.Fatalf("snapshot %d: serviceId empty", i)
		}
		if s.Ts.IsZero() || s.Ts.After(time.Now()) {
			t.Fatalf("snapshot %d: bad ts = %v", i, s.Ts)
		}
	}
	if got[0].Available != true || got[0].CPU != 5 || got[0].Memory != 1024 {
		t.Fatalf("online snapshot bad: %+v", got[0])
	}
	if got[1].Available != false {
		t.Fatalf("stopped snapshot should be unavailable: %+v", got[1])
	}
}

func TestToDockerSnapshots_StampsServerIDAndProvider(t *testing.T) {
	containers := []models.DockerContainer{
		{ID: "c1", Name: "web", State: container.StateRunning, Usage: models.DockerContainerUsage{CpuPercent: 3, MemPercent: 4, MemUsage: 9999}},
		{ID: "c2", Name: "db", State: container.StateExited},
	}
	got := setup_test.SnapshotsFromDockerContainers("host-1", containers)
	if len(got) != 2 {
		t.Fatalf("len = %d", len(got))
	}
	for i, s := range got {
		if s.Meta.ServerID != "host-1" {
			t.Fatalf("snapshot %d: serverID = %q", i, s.Meta.ServerID)
		}
		if s.Meta.Provider != models.ProviderDocker {
			t.Fatalf("snapshot %d: provider = %q", i, s.Meta.Provider)
		}
	}
	if got[0].Status != "running" || !got[0].Available {
		t.Fatalf("running snapshot bad: %+v", got[0])
	}
	if got[1].Status != "exited" || got[1].Available {
		t.Fatalf("exited snapshot bad: %+v", got[1])
	}
}
