package dtos

import (
	"server-monitoring/internal/models"
)

// AgentProviderDTO is one provider's payload inside an agent push.
// Services is the canonical []models.ServiceSnapshot list so adding a
// new provider on the agent side does not require touching this DTO.
type AgentProviderDTO struct {
	Available bool                     `json:"available"`
	Services  []models.ServiceSnapshot `json:"services,omitempty"`
}

// AgentPushDTO is the cycle payload the remote agent POSTs to
// /v1/agents/push. It carries every piece of data the dashboard needs
// from one server at one tick:
//   - identity: name, host, port, description (server CRUD fields)
//   - usage: models.SystemUsage (wrapped by ServerUsageState)
//   - providers: pm2/docker/... each with their snapshot list
//
// The backend stamps the resolved serverId on every snapshot before
// writing so the agent never has to identify itself beyond the bearer
// token.
type AgentPushDTO struct {
	Name        string                      `binding:"required,min=1,max=64" json:"name"`
	Host        string                      `binding:"required,min=1,max=255" json:"host"`
	Port        int                         `binding:"required,min=1,max=65535" json:"port"`
	Description string                      `binding:"omitempty,max=255" json:"description"`
	Usage       models.SystemUsage          `binding:"required" json:"usage"`
	Providers   map[string]AgentProviderDTO `binding:"omitempty" json:"providers"`
}