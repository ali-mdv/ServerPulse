package dtos

import (
	"time"

	"server-monitoring/internal/models"
)

type ServiceSnapshotDTO struct {
	Ts        time.Time           `json:"ts"`
	Meta      models.SnapshotMeta `json:"meta"`
	Status    string              `json:"status,omitempty"`
	Available bool                `json:"available"`
	CPU       *float64            `json:"cpu,omitempty"`
	Memory    *float64            `json:"memory,omitempty"`
	MemUsage  *float64            `json:"memUsage,omitempty"`
	Disk      *float64            `json:"disk,omitempty"`
	NetSent   *float64            `json:"netSent,omitempty"`
	NetRecv   *float64            `json:"netRecv,omitempty"`
}

func ToSnapshotDTO(s models.ServiceSnapshot) ServiceSnapshotDTO {
	dto := ServiceSnapshotDTO{
		Ts:        s.Ts,
		Meta:      s.Meta,
		Status:    s.Status,
		Available: s.Available,
	}
	if s.Meta.ServiceID == models.ProviderStatusSentinelID {
		return dto
	}
	if s.CPU != 0 {
		v := s.CPU
		dto.CPU = &v
	}
	if s.Memory != 0 {
		v := s.Memory
		dto.Memory = &v
	}
	if s.MemUsage != 0 {
		v := s.MemUsage
		dto.MemUsage = &v
	}
	if s.Disk != 0 {
		v := s.Disk
		dto.Disk = &v
	}
	if s.NetSent != 0 {
		v := s.NetSent
		dto.NetSent = &v
	}
	if s.NetRecv != 0 {
		v := s.NetRecv
		dto.NetRecv = &v
	}
	return dto
}

func ToSnapshotDTOs(in []models.ServiceSnapshot) []ServiceSnapshotDTO {
	out := make([]ServiceSnapshotDTO, len(in))
	for i := range in {
		out[i] = ToSnapshotDTO(in[i])
	}
	return out
}

// TrackedServiceDTO mirrors SnapshotMeta but flattens the server
// origin to the top level so list endpoints can scope by server
// without exposing the internal nested-meta shape.
type TrackedServiceDTO struct {
	ServerID  string `json:"serverId"`
	Provider  string `json:"provider"`
	ServiceID string `json:"serviceId"`
	Name      string `json:"name"`
}

func ToTrackedServiceDTO(m models.SnapshotMeta) TrackedServiceDTO {
	return TrackedServiceDTO{
		ServerID:  m.ServerID,
		Provider:  m.Provider,
		ServiceID: m.ServiceID,
		Name:      m.Name,
	}
}

func ToTrackedServiceDTOs(in []models.SnapshotMeta) []TrackedServiceDTO {
	out := make([]TrackedServiceDTO, len(in))
	for i := range in {
		out[i] = ToTrackedServiceDTO(in[i])
	}
	return out
}
