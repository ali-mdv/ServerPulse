package dtos

import (
	"time"

	"server-monitoring/internal/models"
)

// ServerUsageDTO is the wire shape for the latest system usage of one
// server. It wraps models.SystemUsage unchanged so existing dashboard
// cards render without any client-side changes.
type ServerUsageDTO struct {
	ServerID  string             `json:"serverId"`
	UpdatedAt time.Time          `json:"updatedAt"`
	Usage     models.SystemUsage `json:"usage"`
}

// ToServerUsageDTO maps a ServerUsageState to its DTO.
func ToServerUsageDTO(s models.ServerUsageState) *ServerUsageDTO {
	return &ServerUsageDTO{
		ServerID:  s.ServerID,
		UpdatedAt: s.UpdatedAt,
		Usage:     s.Usage,
	}
}

// ToServerUsageDTOs maps a slice.
func ToServerUsageDTOs(in []models.ServerUsageState) []ServerUsageDTO {
	out := make([]ServerUsageDTO, len(in))
	for i := range in {
		dto := ToServerUsageDTO(in[i])
		out[i] = *dto
	}
	return out
}