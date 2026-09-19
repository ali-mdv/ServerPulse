package services

import (
	"context"
	"time"

	"server-monitoring/internal/models"
	"server-monitoring/internal/repository"
	database "server-monitoring/pkg/mongo"
)

const hostStateID = "host"

type StateService interface {
	RecordProviderState(ctx context.Context, provider string, available bool, services []models.ServiceSnapshot) error
	RecordHostState(ctx context.Context, usage models.SystemUsage) error
	GetProviderState(ctx context.Context, provider string) (*models.ProviderState, error)
	GetHostState(ctx context.Context) (*models.HostState, error)
}

type stateService struct {
	repo repository.StateRepository
}

func NewStateService(dbName string) StateService {
	db := database.GetDatabase(dbName)
	repo := repository.NewStateRepository(db)
	return &stateService{repo: repo}
}

func (s *stateService) RecordProviderState(ctx context.Context, provider string, available bool, services []models.ServiceSnapshot) error {
	if services == nil {
		services = []models.ServiceSnapshot{}
	}
	return s.repo.UpsertProviderState(ctx, models.ProviderState{
		Provider:  provider,
		UpdatedAt: time.Now().UTC(),
		Available: available,
		Services:  services,
	})
}

func (s *stateService) RecordHostState(ctx context.Context, usage models.SystemUsage) error {
	return s.repo.UpsertHostState(ctx, models.HostState{
		ID:        hostStateID,
		UpdatedAt: time.Now().UTC(),
		Usage:     usage,
	})
}

func (s *stateService) GetProviderState(ctx context.Context, provider string) (*models.ProviderState, error) {
	return s.repo.GetProviderState(ctx, provider)
}

func (s *stateService) GetHostState(ctx context.Context) (*models.HostState, error) {
	return s.repo.GetHostState(ctx)
}
