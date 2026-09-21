package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"server-monitoring/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type HistoryScheduler interface {
	Start(ctx context.Context)
}

type historyScheduler struct {
	pm2           PM2Service
	docker        DockerService
	system        SystemService
	history       HistoryService
	state         StateService
	servers       ServerService
	notifications NotificationService
	interval      time.Duration
	serverID      bson.ObjectID
}

func NewHistoryScheduler(pm2 PM2Service, docker DockerService, system SystemService, history HistoryService, state StateService, servers ServerService, notifications NotificationService, interval time.Duration, serverID bson.ObjectID) HistoryScheduler {
	return &historyScheduler{
		pm2:           pm2,
		docker:        docker,
		system:        system,
		history:       history,
		state:         state,
		servers:       servers,
		notifications: notifications,
		interval:      interval,
		serverID:      serverID,
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

	serverID := s.serverID.Hex()

	// Heartbeat the local server row so its status/lastSeen reflect the
	// in-process scheduler activity.
	if err := s.servers.TouchSeen(tickCtx, s.serverID, models.ServerStatusOnline); err != nil {
		log.Printf("history: local heartbeat failed: %v", err)
	}

	prevPM2, _ := s.state.GetProviderState(tickCtx, serverID, models.ProviderPM2)
	if procs, err := s.pm2.List(); err != nil {
		log.Printf("history: pm2 list failed: %v", err)
		_ = s.history.RecordProviderUnavailable(tickCtx, serverID, models.ProviderPM2)
		_ = s.state.RecordProviderState(tickCtx, serverID, models.ProviderPM2, false, nil)
		s.emitProviderTransition(tickCtx, serverID, models.ProviderPM2, prevPM2, false)
	} else {
		snaps := toSnapshots(serverID, procs)
		if err := s.history.RecordPM2Snapshot(tickCtx, serverID, procs); err != nil {
			log.Printf("history: pm2 snapshot write failed: %v", err)
		}
		if err := s.state.RecordProviderState(tickCtx, serverID, models.ProviderPM2, true, snaps); err != nil {
			log.Printf("history: pm2 state write failed: %v", err)
		}
		s.emitProviderTransition(tickCtx, serverID, models.ProviderPM2, prevPM2, true)
		s.emitServiceTransitions(tickCtx, serverID, models.ProviderPM2, prevPM2, snaps)
	}

	prevDocker, _ := s.state.GetProviderState(tickCtx, serverID, models.ProviderDocker)
	if containers, err := s.docker.ContainersList(true); err != nil {
		log.Printf("history: docker list failed: %v", err)
		_ = s.history.RecordProviderUnavailable(tickCtx, serverID, models.ProviderDocker)
		_ = s.state.RecordProviderState(tickCtx, serverID, models.ProviderDocker, false, nil)
		s.emitProviderTransition(tickCtx, serverID, models.ProviderDocker, prevDocker, false)
	} else {
		snaps := toDockerSnapshots(serverID, containers)
		if err := s.history.RecordDockerSnapshot(tickCtx, serverID, containers); err != nil {
			log.Printf("history: docker snapshot write failed: %v", err)
		}
		if err := s.state.RecordProviderState(tickCtx, serverID, models.ProviderDocker, true, snaps); err != nil {
			log.Printf("history: docker state write failed: %v", err)
		}
		s.emitProviderTransition(tickCtx, serverID, models.ProviderDocker, prevDocker, true)
		s.emitServiceTransitions(tickCtx, serverID, models.ProviderDocker, prevDocker, snaps)
	}

	usage := s.system.SystemUsage()
	if err := s.history.RecordSystemSnapshot(tickCtx, serverID, usage); err != nil {
		log.Printf("history: system snapshot write failed: %v", err)
	}
	if err := s.state.RecordServerUsage(tickCtx, serverID, usage); err != nil {
		log.Printf("history: server usage write failed: %v", err)
	}
}

// emitProviderTransition raises a notification when a provider's
// availability flips between ticks. The first observation (prev == nil)
// is treated as a baseline and stays silent.
func (s *historyScheduler) emitProviderTransition(ctx context.Context, serverID, provider string, prev *models.ProviderState, available bool) {
	if s.notifications == nil || prev == nil || prev.Available == available {
		return
	}

	n := models.Notification{
		ServerID: serverID,
		Meta:     map[string]any{"provider": provider},
	}
	if available {
		n.Type = models.NotificationTypeProviderAvailable
		n.Severity = models.NotificationSeverityInfo
		n.Title = fmt.Sprintf("%s is back online", providerLabel(provider))
		n.Message = fmt.Sprintf("The %s daemon on this server is reachable again.", provider)
	} else {
		n.Type = models.NotificationTypeProviderUnavailable
		n.Severity = models.NotificationSeverityCritical
		n.Title = fmt.Sprintf("%s is unavailable", providerLabel(provider))
		n.Message = fmt.Sprintf("The %s daemon on this server could not be reached.", provider)
	}

	if _, err := s.notifications.Notify(ctx, n); err != nil {
		log.Printf("notifications: provider transition failed: %v", err)
	}
}

// emitServiceTransitions raises a notification for every managed service
// whose running state changed since the previous tick. Only runs while
// the provider itself is available, so a dead daemon produces one
// provider alert instead of one per service.
func (s *historyScheduler) emitServiceTransitions(ctx context.Context, serverID, provider string, prev *models.ProviderState, current []models.ServiceSnapshot) {
	if s.notifications == nil || prev == nil || !prev.Available {
		return
	}

	before := make(map[string]models.ServiceSnapshot, len(prev.Services))
	for _, svc := range prev.Services {
		before[svc.Meta.ServiceID] = svc
	}

	for _, svc := range current {
		old, ok := before[svc.Meta.ServiceID]
		if !ok || old.Available == svc.Available {
			continue
		}

		n := models.Notification{
			ServerID: serverID,
			Meta: map[string]any{
				"provider":  provider,
				"serviceId": svc.Meta.ServiceID,
				"name":      svc.Meta.Name,
			},
		}
		if svc.Available {
			n.Type = models.NotificationTypeServiceUp
			n.Severity = models.NotificationSeverityInfo
			n.Title = fmt.Sprintf("%s is back up", svc.Meta.Name)
			n.Message = fmt.Sprintf("%s on %s is running again.", svc.Meta.Name, provider)
		} else {
			n.Type = models.NotificationTypeServiceDown
			n.Severity = models.NotificationSeverityCritical
			n.Title = fmt.Sprintf("%s is down", svc.Meta.Name)
			n.Message = fmt.Sprintf("%s (%s) on %s is not running.", svc.Meta.Name, svc.Status, provider)
		}

		if _, err := s.notifications.Notify(ctx, n); err != nil {
			log.Printf("notifications: service transition failed: %v", err)
		}
	}
}

func providerLabel(provider string) string {
	switch provider {
	case models.ProviderPM2:
		return "PM2"
	case models.ProviderDocker:
		return "Docker"
	default:
		return provider
	}
}

func toSnapshots(serverID string, procs []models.PM2Process) []models.ServiceSnapshot {
	now := time.Now().UTC()
	out := make([]models.ServiceSnapshot, 0, len(procs))
	for _, p := range procs {
		out = append(out, models.ServiceSnapshot{
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
	return out
}

func toDockerSnapshots(serverID string, containers []models.DockerContainer) []models.ServiceSnapshot {
	now := time.Now().UTC()
	out := make([]models.ServiceSnapshot, 0, len(containers))
	for _, c := range containers {
		out = append(out, models.ServiceSnapshot{
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
	return out
}

// SnapshotsFromPM2 is the exported form of toSnapshots. Used by the
// test/setup helpers; production code goes through HistoryScheduler.
func SnapshotsFromPM2(serverID string, procs []models.PM2Process) []models.ServiceSnapshot {
	return toSnapshots(serverID, procs)
}

// SnapshotsFromDocker is the exported form of toDockerSnapshots. Used
// by the test/setup helpers; production code goes through
// HistoryScheduler.
func SnapshotsFromDocker(serverID string, containers []models.DockerContainer) []models.ServiceSnapshot {
	return toDockerSnapshots(serverID, containers)
}