package dtos

import (
	"time"

	"server-monitoring/internal/models"
)

// ProviderStateDTO is the per-server provider state shape. It carries
// the full snapshot list (mirroring the existing pm2/docker endpoints)
// so the dashboard can render the section without a separate round-trip.
type ProviderStateDTO struct {
	ServerID  string                `json:"serverId"`
	Provider  string                `json:"provider"`
	UpdatedAt time.Time             `json:"updatedAt"`
	Available bool                  `json:"available"`
	Services  []ServiceSnapshotDTO  `json:"services,omitempty"`
}

// ToProviderStateDTO maps a ProviderState model to the wire shape. The
// nested ServiceSnapshot conversion reuses the existing history
// converter so semantics stay identical.
func ToProviderStateDTO(p models.ProviderState) ProviderStateDTO {
	return ProviderStateDTO{
		ServerID:  p.ServerID,
		Provider:  p.Provider,
		UpdatedAt: p.UpdatedAt,
		Available: p.Available,
		Services:  ToSnapshotDTOs(p.Services),
	}
}

// ToProviderStateDTOs maps a slice of models.
func ToProviderStateDTOs(in []models.ProviderState) []ProviderStateDTO {
	out := make([]ProviderStateDTO, len(in))
	for i := range in {
		out[i] = ToProviderStateDTO(in[i])
	}
	return out
}

// ToServerProviderAvailabilityDTOs rolls up provider health into the
// nested map shape used inside ServerDTO/ServerListItemDTO. Each entry
// only carries availability + freshness — the snapshot list is dropped
// to keep the list view small.
func ToServerProviderAvailabilityDTOs(providers map[string]models.ProviderState) map[string]ServerProviderAvailabilityDTO {
	if len(providers) == 0 {
		return nil
	}
	out := make(map[string]ServerProviderAvailabilityDTO, len(providers))
	for k, p := range providers {
		entry := ServerProviderAvailabilityDTO{Available: p.Available}
		if !p.UpdatedAt.IsZero() {
			ts := p.UpdatedAt
			entry.UpdatedAt = &ts
		}
		out[k] = entry
	}
	return out
}