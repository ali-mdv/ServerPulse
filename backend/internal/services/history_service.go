package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"server-monitoring/internal/models"
	"server-monitoring/internal/repository"
	database "server-monitoring/pkg/mongo"
)

type HistoryService interface {
	EnsureSchema(ctx context.Context) error
	RecordPM2Snapshot(ctx context.Context, serverID string, procs []models.PM2Process) error
	RecordDockerSnapshot(ctx context.Context, serverID string, containers []models.DockerContainer) error
	RecordSystemSnapshot(ctx context.Context, serverID string, usage models.SystemUsage) error
	RecordProviderUnavailable(ctx context.Context, serverID, provider string) error
	InsertHistoryBatch(ctx context.Context, snapshots []models.ServiceSnapshot) error
	ListTrackedServices(ctx context.Context, serverID, provider string) ([]models.SnapshotMeta, error)
	FindSeries(ctx context.Context, serverID, provider, serviceID string, from, to time.Time, bucket time.Duration) ([]models.ServiceSnapshot, error)
}

type historyService struct {
	repo      repository.HistoryRepository
	settings  SettingsService
	retention time.Duration
}

// NewHistoryService builds the production service. Retention now comes
// from the settings singleton, read on every EnsureSchema call, so a
// settings change updates the TTL without a restart.
func NewHistoryService(dbName string, settings SettingsService) HistoryService {
	db := database.GetDatabase(dbName)
	repo := repository.NewHistoryRepository(db)
	return &historyService{repo: repo, settings: settings}
}

// NewHistoryServiceFromRepo is the repo-backed constructor used by the
// test/setup helpers. Not part of the stable API.
func NewHistoryServiceFromRepo(repo repository.HistoryRepository, retention time.Duration) HistoryService {
	return &historyService{repo: repo, retention: retention}
}

// NewHistoryServiceWithSettings is the repo+settings-backed constructor
// used by the test/setup helpers. Not part of the stable API.
func NewHistoryServiceWithSettings(repo repository.HistoryRepository, settings SettingsService) HistoryService {
	return &historyService{repo: repo, settings: settings}
}

func (s *historyService) EnsureSchema(ctx context.Context) error {
	return s.repo.EnsureCollection(ctx, s.currentRetention(ctx))
}

// currentRetention resolves the retention to apply: the settings
// singleton when configured, else the static value, else the default.
func (s *historyService) currentRetention(ctx context.Context) time.Duration {
	if s.settings != nil {
		if retention, err := s.settings.HistoryRetention(ctx); err == nil && retention > 0 {
			return retention
		}
	}
	if s.retention > 0 {
		return s.retention
	}
	return models.DefaultHistoryRetention
}

func (s *historyService) RecordPM2Snapshot(ctx context.Context, serverID string, procs []models.PM2Process) error {
	now := time.Now().UTC()
	snapshots := make([]models.ServiceSnapshot, 0, len(procs))
	for _, p := range procs {
		snapshots = append(snapshots, models.ServiceSnapshot{
			Ts: now,
			Meta: models.SnapshotMeta{
				ServerID:  serverID,
				Provider:  models.ProviderPM2,
				ServiceID: fmt.Sprintf("%d", p.PMID),
				Name:      p.Name,
			},
			Status:    p.PM2Env.Status,
			Available: p.PM2Env.Status == "online",
			CPU:       p.Monit.CPU,
			Memory:    p.Monit.Memory,
		})
	}
	if err := s.repo.InsertBatch(ctx, snapshots); err != nil {
		log.Printf("history: insert pm2 batch failed: %v", err)
		return err
	}
	return nil
}

func (s *historyService) RecordDockerSnapshot(ctx context.Context, serverID string, containers []models.DockerContainer) error {
	now := time.Now().UTC()
	snapshots := make([]models.ServiceSnapshot, 0, len(containers))
	for _, c := range containers {
		snapshots = append(snapshots, models.ServiceSnapshot{
			Ts: now,
			Meta: models.SnapshotMeta{
				ServerID:  serverID,
				Provider:  models.ProviderDocker,
				ServiceID: c.ID,
				Name:      c.Name,
			},
			Status:    string(c.State),
			Available: c.State == "running",
			CPU:       c.Usage.CpuPercent,
			Memory:    c.Usage.MemPercent,
			MemUsage:  float64(c.Usage.MemUsage),
		})
	}
	if err := s.repo.InsertBatch(ctx, snapshots); err != nil {
		log.Printf("history: insert docker batch failed: %v", err)
		return err
	}
	return nil
}

func (s *historyService) RecordProviderUnavailable(ctx context.Context, serverID, provider string) error {
	snap := models.ServiceSnapshot{
		Ts: time.Now().UTC(),
		Meta: models.SnapshotMeta{
			ServerID:  serverID,
			Provider:  provider,
			ServiceID: models.ProviderStatusSentinelID,
			Name:      models.ProviderStatusSentinelName,
		},
		Available: false,
	}
	if err := s.repo.InsertBatch(ctx, []models.ServiceSnapshot{snap}); err != nil {
		log.Printf("history: insert %s unavailable sentinel failed: %v", provider, err)
		return err
	}
	return nil
}

// InsertHistoryBatch is the provider-agnostic write path used by the
// agent ingestion service. Callers are responsible for stamping
// Meta.ServerID, Meta.Provider and Ts on each snapshot.
func (s *historyService) InsertHistoryBatch(ctx context.Context, snapshots []models.ServiceSnapshot) error {
	return s.repo.InsertBatch(ctx, snapshots)
}

func (s *historyService) RecordSystemSnapshot(ctx context.Context, serverID string, usage models.SystemUsage) error {
	now := time.Now().UTC()
	snapshots := []models.ServiceSnapshot{
		{
			Ts: now,
			Meta: models.SnapshotMeta{
				ServerID:  serverID,
				Provider:  models.ProviderSystem,
				ServiceID: models.SystemMetricCPU,
				Name:      "CPU",
			},
			Available: true,
			CPU:       usage.CpuUsage,
		},
		{
			Ts: now,
			Meta: models.SnapshotMeta{
				ServerID:  serverID,
				Provider:  models.ProviderSystem,
				ServiceID: models.SystemMetricMemory,
				Name:      "Memory",
			},
			Available: true,
			Memory:    usage.MemUsage.UsedPercent,
			MemUsage:  float64(usage.MemUsage.Used),
		},
		{
			Ts: now,
			Meta: models.SnapshotMeta{
				ServerID:  serverID,
				Provider:  models.ProviderSystem,
				ServiceID: models.SystemMetricDisk,
				Name:      "Disk",
			},
			Available: true,
			Disk:      usage.DiskUsage.UsedPercent,
		},
		{
			Ts: now,
			Meta: models.SnapshotMeta{
				ServerID:  serverID,
				Provider:  models.ProviderSystem,
				ServiceID: models.SystemMetricNetSent,
				Name:      "Network Sent",
			},
			Available: true,
			NetSent:   float64(usage.NetIO.Send),
		},
		{
			Ts: now,
			Meta: models.SnapshotMeta{
				ServerID:  serverID,
				Provider:  models.ProviderSystem,
				ServiceID: models.SystemMetricNetRecv,
				Name:      "Network Received",
			},
			Available: true,
			NetRecv:   float64(usage.NetIO.Received),
		},
	}
	if err := s.repo.InsertBatch(ctx, snapshots); err != nil {
		log.Printf("history: insert system batch failed: %v", err)
		return err
	}
	return nil
}

func (s *historyService) ListTrackedServices(ctx context.Context, serverID, provider string) ([]models.SnapshotMeta, error) {
	return s.repo.ListTrackedServices(ctx, serverID, provider)
}

func (s *historyService) FindSeries(ctx context.Context, serverID, provider, serviceID string, from, to time.Time, bucket time.Duration) ([]models.ServiceSnapshot, error) {
	if bucket < 0 {
		bucket = 0
	}
	return s.repo.FindSeries(ctx, serverID, provider, serviceID, from, to, bucket)
}