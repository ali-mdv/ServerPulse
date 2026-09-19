package services

import (
	"context"
	"log"
	"time"

	"server-monitoring/internal/dtos"
	"server-monitoring/internal/models"
	"server-monitoring/internal/repository"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// AgentService is the ingestion path for remote servers. A successful
// push updates every storage tier the dashboard reads from:
//   - servers: heartbeats lastSeen + status
//   - server_usage: latest SystemUsage per server
//   - provider_state: latest snapshot list per (server, provider)
//   - service_history: append-only snapshots for trend charts
type AgentService interface {
	IngestPush(ctx context.Context, server models.Server, push dtos.AgentPushDTO) error
}

type agentService struct {
	serverRepo repository.ServerRepository
	state      StateService
	history    HistoryService
}

func NewAgentService(serverRepo repository.ServerRepository, state StateService, history HistoryService) AgentService {
	return &agentService{
		serverRepo: serverRepo,
		state:      state,
		history:    history,
	}
}

// IngestPush is the single entry point for an agent cycle. It is
// deliberately tolerant: a single provider failing (or being absent)
// must not abort the whole push — the dashboard should still see the
// usage and any provider that did report back.
func (s *agentService) IngestPush(ctx context.Context, server models.Server, push dtos.AgentPushDTO) error {
	serverID := server.HexID()

	// 1. Refresh server identity fields. We trust the agent on name/
	//    description and on any non-empty host/port it reports, but keep
	//    the existing token/status untouched.
	now := time.Now().UTC()
	if s.identityChanged(server, push) {
		if err := s.touchIdentity(ctx, server.ID, push); err != nil {
			log.Printf("agent: identity update failed for %s: %v", serverID, err)
		}
	}

	// 2. Heartbeat + status.
	if err := s.serverRepo.TouchSeen(ctx, server.ID, models.ServerStatusOnline); err != nil {
		log.Printf("agent: touch seen failed for %s: %v", serverID, err)
	}

	// 3. Latest usage.
	if err := s.state.RecordServerUsage(ctx, serverID, push.Usage); err != nil {
		log.Printf("agent: usage write failed for %s: %v", serverID, err)
	}
	if err := s.history.RecordSystemSnapshot(ctx, serverID, push.Usage); err != nil {
		log.Printf("agent: usage history failed for %s: %v", serverID, err)
	}

	// 4. Per-provider state + history. Unknown providers are accepted so
	//    a forward-compatible agent can ship new data without a backend
	//    change.
	for providerName, payload := range push.Providers {
		services := payload.Services
		for i := range services {
			services[i].Meta.ServerID = serverID
			services[i].Meta.Provider = providerName
			services[i].Ts = now
		}

		if err := s.state.RecordProviderState(ctx, serverID, providerName, payload.Available, services); err != nil {
			log.Printf("agent: provider %s state write failed for %s: %v", providerName, serverID, err)
			continue
		}

		// Best-effort history insert. Snapshot shape is already canonical
		// so no provider-specific conversion is needed.
		if len(services) > 0 {
			if err := s.insertHistory(ctx, services); err != nil {
				log.Printf("agent: provider %s history failed for %s: %v", providerName, serverID, err)
			}
		}
	}

	return nil
}

// insertHistory funnels the provider's snapshots into the time-series
// collection via the history service's generic batch insert.
func (s *agentService) insertHistory(ctx context.Context, snapshots []models.ServiceSnapshot) error {
	if len(snapshots) == 0 {
		return nil
	}
	return s.history.InsertHistoryBatch(ctx, snapshots)
}

// identityChanged reports whether the push carries a different identity
// from the stored server. Empty host/port from the push are treated as
// "no opinion" and do not trigger an update.
func (s *agentService) identityChanged(server models.Server, push dtos.AgentPushDTO) bool {
	if server.Name != push.Name {
		return true
	}
	if server.Description != push.Description {
		return true
	}
	if push.Host != "" && server.Host != push.Host {
		return true
	}
	if push.Port > 0 && server.Port != push.Port {
		return true
	}
	return false
}

// touchIdentity updates the mutable CRUD fields the agent sent, without
// overwriting LastSeen / Status / AgentToken / CreatedAt. Empty host/port
// from the push are ignored so the agent can omit details it does not
// know without clearing previously supplied values.
func (s *agentService) touchIdentity(ctx context.Context, id bson.ObjectID, push dtos.AgentPushDTO) error {
	set := bson.D{
		{Key: "name", Value: push.Name},
		{Key: "updatedAt", Value: time.Now().UTC()},
	}
	if push.Host != "" {
		set = append(set, bson.E{Key: "host", Value: push.Host})
	}
	if push.Port > 0 {
		set = append(set, bson.E{Key: "port", Value: push.Port})
	}
	// Allow description to be intentionally cleared.
	set = append(set, bson.E{Key: "description", Value: push.Description})

	update := bson.D{{Key: "$set", Value: set}}
	if _, err := s.serverRepo.UpdateByID(ctx, id, update); err != nil {
		return err
	}
	return nil
}