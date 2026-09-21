package services

import (
	"context"
	"time"

	"server-monitoring/internal/models"
	"server-monitoring/internal/repository"
	database "server-monitoring/pkg/mongo"
)

// LocalServerID is the well-known ID the scheduler uses for the host
// the backend runs on. It is also the _id of the ServerUsageState
// document for the local host and the serverId attached to every
// snapshot the in-process scheduler records.
const LocalServerID = "local"

type StateService interface {
	RecordProviderState(ctx context.Context, serverID, provider string, available bool, services []models.ServiceSnapshot) error
	RecordServerUsage(ctx context.Context, serverID string, usage models.SystemUsage) error
	GetProviderState(ctx context.Context, serverID, provider string) (*models.ProviderState, error)
	GetServerUsage(ctx context.Context, serverID string) (*models.ServerUsageState, error)
}

type stateService struct {
	repo repository.StateRepository
}

func NewStateService(dbName string) StateService {
	db := database.GetDatabase(dbName)
	repo := repository.NewStateRepository(db)
	return &stateService{repo: repo}
}

// NewStateServiceFromRepo is the repo-backed constructor used by the
// test/setup helpers. Not part of the stable API.
func NewStateServiceFromRepo(repo repository.StateRepository) StateService {
	return &stateService{repo: repo}
}

func (s *stateService) RecordProviderState(ctx context.Context, serverID, provider string, available bool, services []models.ServiceSnapshot) error {
	if services == nil {
		services = []models.ServiceSnapshot{}
	}
	// Stamp the server origin on every snapshot so the stored shape
	// matches the time-series meta and downstream readers don't have to
	// guess the source.
	for i := range services {
		services[i].Meta.ServerID = serverID
	}
	return s.repo.UpsertProviderState(ctx, models.ProviderState{
		ServerID:  serverID,
		Provider:  provider,
		UpdatedAt: time.Now().UTC(),
		Available: available,
		Services:  services,
	})
}

func (s *stateService) RecordServerUsage(ctx context.Context, serverID string, usage models.SystemUsage) error {
	return s.repo.UpsertServerUsage(ctx, models.ServerUsageState{
		ServerID:  serverID,
		UpdatedAt: time.Now().UTC(),
		Usage:     usage,
	})
}

func (s *stateService) GetProviderState(ctx context.Context, serverID, provider string) (*models.ProviderState, error) {
	return s.repo.GetProviderState(ctx, serverID, provider)
}

func (s *stateService) GetServerUsage(ctx context.Context, serverID string) (*models.ServerUsageState, error) {
	return s.repo.GetServerUsage(ctx, serverID)
}