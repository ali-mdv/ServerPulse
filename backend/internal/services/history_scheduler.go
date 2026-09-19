package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"server-monitoring/internal/models"
)

type HistoryScheduler interface {
	Start(ctx context.Context)
}

type historyScheduler struct {
	pm2      PM2Service
	docker   DockerService
	system   SystemService
	history  HistoryService
	state    StateService
	interval time.Duration
}

func NewHistoryScheduler(pm2 PM2Service, docker DockerService, system SystemService, history HistoryService, state StateService, interval time.Duration) HistoryScheduler {
	return &historyScheduler{
		pm2:      pm2,
		docker:   docker,
		system:   system,
		history:  history,
		state:    state,
		interval: interval,
	}
}

func (s *historyScheduler) Start(ctx context.Context) {
	if err := s.history.EnsureSchema(ctx); err != nil {
		log.Printf("history: ensure schema failed: %v", err)
	}

	s.tick(ctx)

	go s.run(ctx)
}

func (s *historyScheduler) run(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

func (s *historyScheduler) tick(ctx context.Context) {
	tickCtx, cancel := context.WithTimeout(ctx, s.interval)
	defer cancel()

	if procs, err := s.pm2.List(); err != nil {
		log.Printf("history: pm2 list failed: %v", err)
		_ = s.history.RecordProviderUnavailable(tickCtx, "pm2")
		_ = s.state.RecordProviderState(tickCtx, "pm2", false, nil)
	} else {
		if err := s.history.RecordPM2Snapshot(tickCtx, procs); err != nil {
			log.Printf("history: pm2 snapshot write failed: %v", err)
		}
		if err := s.state.RecordProviderState(tickCtx, "pm2", true, toSnapshots(procs)); err != nil {
			log.Printf("history: pm2 state write failed: %v", err)
		}
	}

	if containers, err := s.docker.ContainersList(true); err != nil {
		log.Printf("history: docker list failed: %v", err)
		_ = s.history.RecordProviderUnavailable(tickCtx, "docker")
		_ = s.state.RecordProviderState(tickCtx, "docker", false, nil)
	} else {
		if err := s.history.RecordDockerSnapshot(tickCtx, containers); err != nil {
			log.Printf("history: docker snapshot write failed: %v", err)
		}
		if err := s.state.RecordProviderState(tickCtx, "docker", true, toDockerSnapshots(containers)); err != nil {
			log.Printf("history: docker state write failed: %v", err)
		}
	}

	usage := s.system.SystemUsage()
	if err := s.history.RecordSystemSnapshot(tickCtx, usage); err != nil {
		log.Printf("history: system snapshot write failed: %v", err)
	}
	if err := s.state.RecordHostState(tickCtx, usage); err != nil {
		log.Printf("history: host state write failed: %v", err)
	}
}

func toSnapshots(procs []models.PM2Process) []models.ServiceSnapshot {
	now := time.Now().UTC()
	out := make([]models.ServiceSnapshot, 0, len(procs))
	for _, p := range procs {
		out = append(out, models.ServiceSnapshot{
			Ts: now,
			Meta: models.SnapshotMeta{
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
	return out
}

func toDockerSnapshots(containers []models.DockerContainer) []models.ServiceSnapshot {
	now := time.Now().UTC()
	out := make([]models.ServiceSnapshot, 0, len(containers))
	for _, c := range containers {
		out = append(out, models.ServiceSnapshot{
			Ts: now,
			Meta: models.SnapshotMeta{
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
	return out
}
