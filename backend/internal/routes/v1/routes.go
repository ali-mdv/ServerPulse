package v1

import (
	"server-monitoring/internal/services"

	"github.com/gin-gonic/gin"
)

type Services struct {
	PM2     services.PM2Service
	Docker  services.DockerService
	System  services.SystemService
	History services.HistoryService
	State   services.StateService
	Servers services.ServerService
	Agent   services.AgentService
}

func RegisterV1Routes(r *gin.RouterGroup, s *Services) {
	api := r.Group("/v1")

	{
		RegisterAuthRoutes(api)
		RegisterUserRoutes(api)
		RegisterDockerRoutes(api, s.Docker)
		RegisterReportRoutes(api)
		RegisterSystemRoutes(api, s.System)
		RegisterPM2Routes(api, s.PM2)
		RegisterHistoryRoutes(api, s.History)
		RegisterStateRoutes(api, s.State)
		RegisterServerRoutes(api, s)
		RegisterAgentRoutes(api, s)
	}
}