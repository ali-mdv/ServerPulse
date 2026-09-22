package dtos

import (
	"time"

	"server-monitoring/internal/models"
)

// SettingsDTO is the wire shape for the settings singleton. Durations
// are exposed in seconds so the client never has to know Go's duration
// encoding.
type SettingsDTO struct {
	Appearance                 string    `json:"appearance"`
	HistoryPollIntervalSeconds int64     `json:"historyPollIntervalSeconds"`
	HistoryRetentionSeconds    int64     `json:"historyRetentionSeconds"`
	UpdatedAt                  time.Time `json:"updatedAt"`
}

// UpdateSettingsDTO is a partial update: nil fields are left untouched.
type UpdateSettingsDTO struct {
	Appearance                 *string `json:"appearance"`
	HistoryPollIntervalSeconds *int64  `json:"historyPollIntervalSeconds"`
	HistoryRetentionSeconds    *int64  `json:"historyRetentionSeconds"`
}

func ToSettingsDTO(s models.Settings) SettingsDTO {
	return SettingsDTO{
		Appearance:                 string(s.Appearance),
		HistoryPollIntervalSeconds: int64(s.HistoryPollInterval.Seconds()),
		HistoryRetentionSeconds:    int64(s.HistoryRetention.Seconds()),
		UpdatedAt:                  s.UpdatedAt,
	}
}
