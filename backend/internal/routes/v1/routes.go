package v1

import (
	"net/http"

	"server-monitoring/internal/services"

	"github.com/gin-gonic/gin"
)

type Services struct {
	PM2           services.PM2Service
	Docker        services.DockerService
	System        services.SystemService
	History       services.HistoryService
	State         services.StateService
	Servers       services.ServerService
	Agent         services.AgentService
	Notifications services.NotificationService
	Settings      services.SettingsService
	Commands      services.CommandService
	// AgentSocket is the Socket.IO control-channel server. It is
	// optional (nil in tests) so /agents/push can run without it.
	AgentSocket AgentSocketHandler
}

// AgentSocketHandler is the minimal surface the router needs from the
// agent Socket.IO server (kept as an interface so tests can stub it).
type AgentSocketHandler interface {
	Handler() http.Handler
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
		RegisterHistoryRoutes(api, s.History, s.Servers)
		RegisterStateRoutes(api, s.State, s.Servers)
		RegisterServerRoutes(api, s)
		RegisterServerControlRoutes(api, s)
		RegisterAgentRoutes(api, s)
		RegisterNotificationRoutes(api, s.Notifications, s.Servers)
		RegisterSettingsRoutes(api, s.Settings)
	}
}
