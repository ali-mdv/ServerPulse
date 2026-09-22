package services

import (
	"context"
	"time"

	"server-monitoring/internal/dtos"
	"server-monitoring/internal/models"
	"server-monitoring/internal/repository"
	apperrors "server-monitoring/pkg/errors"
	database "server-monitoring/pkg/mongo"

	"net/http"
)

// Bounds for the editable duration settings.
const (
	minHistoryPollInterval = 5 * time.Second
	maxHistoryPollInterval = 24 * time.Hour
	minHistoryRetention    = time.Hour
	maxHistoryRetention    = 365 * 24 * time.Hour
)

type SettingsService interface {
	Get(ctx context.Context) (*models.Settings, error)
	Update(ctx context.Context, dto dtos.UpdateSettingsDTO) (*models.Settings, error)
	HistoryPollInterval(ctx context.Context) (time.Duration, error)
	HistoryRetention(ctx context.Context) (time.Duration, error)
}

type settingsService struct {
	repo repository.SettingsRepository
}

func NewSettingsService(dbName string) SettingsService {
	db := database.GetDatabase(dbName)
	return &settingsService{repo: repository.NewSettingsRepository(db)}
}

// NewSettingsServiceFromRepo is the repo-backed constructor used by the
// test/setup helpers. Not part of the stable API.
func NewSettingsServiceFromRepo(repo repository.SettingsRepository) SettingsService {
	return &settingsService{repo: repo}
}

// Get returns the singleton, creating it with defaults on first access.
func (s *settingsService) Get(ctx context.Context) (*models.Settings, error) {
	settings, err := s.repo.Get(ctx)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return s.repo.Ensure(ctx, models.DefaultSettings())
	}
	return settings, nil
}

func (s *settingsService) Update(ctx context.Context, dto dtos.UpdateSettingsDTO) (*models.Settings, error) {
	current, err := s.Get(ctx)
	if err != nil {
		return nil, err
	}

	if dto.Appearance != nil {
		appearance := models.Appearance(*dto.Appearance)
		if !appearance.Valid() {
			return nil, apperrors.New(http.StatusUnprocessableEntity, "appearance must be system, light or dark")
		}
		current.Appearance = appearance
	}

	if dto.HistoryPollIntervalSeconds != nil {
		interval := time.Duration(*dto.HistoryPollIntervalSeconds) * time.Second
		if interval < minHistoryPollInterval || interval > maxHistoryPollInterval {
			return nil, apperrors.New(http.StatusUnprocessableEntity, "historyPollIntervalSeconds out of range")
		}
		current.HistoryPollInterval = interval
	}

	if dto.HistoryRetentionSeconds != nil {
		retention := time.Duration(*dto.HistoryRetentionSeconds) * time.Second
		if retention < minHistoryRetention || retention > maxHistoryRetention {
			return nil, apperrors.New(http.StatusUnprocessableEntity, "historyRetentionSeconds out of range")
		}
		current.HistoryRetention = retention
	}

	current.UpdatedAt = time.Now().UTC()
	return s.repo.Upsert(ctx, *current)
}

// HistoryPollInterval is the convenience accessor the scheduler uses. It
// falls back to the model default if the settings row is unavailable.
func (s *settingsService) HistoryPollInterval(ctx context.Context) (time.Duration, error) {
	settings, err := s.Get(ctx)
	if err != nil {
		return models.DefaultHistoryPollInterval, err
	}
	if settings.HistoryPollInterval <= 0 {
		return models.DefaultHistoryPollInterval, nil
	}
	return settings.HistoryPollInterval, nil
}

// HistoryRetention is the convenience accessor the history service uses.
func (s *settingsService) HistoryRetention(ctx context.Context) (time.Duration, error) {
	settings, err := s.Get(ctx)
	if err != nil {
		return models.DefaultHistoryRetention, err
	}
	if settings.HistoryRetention <= 0 {
		return models.DefaultHistoryRetention, nil
	}
	return settings.HistoryRetention, nil
}
