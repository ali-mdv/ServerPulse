package v1

import (
	"server-monitoring/internal/handlers"
	"server-monitoring/internal/middlewares"

	"github.com/gin-gonic/gin"
)

// RegisterServerControlRoutes mounts server-scoped Docker/PM2 control
// under /servers/:serverId/.... These share the /servers group with the
// CRUD routes so AuthMiddleware applies unchanged.
//
// Path shape mirrors the local /docker and /pm2 action routes so the
// frontend can swap the prefix when targeting a remote agent:
//
//	GET /servers/:serverId/docker/containers/:containerId/{start,stop,restart,logs}
//	GET /servers/:serverId/pm2/services/:id/{start,stop,restart,logs}
func RegisterServerControlRoutes(r *gin.RouterGroup, s *Services) {
	handler := handlers.NewServerControlHandler(s.Commands)

	api := r.Group("/servers")
	api.Use(middlewares.AuthMiddleware())

	// Relative to the /servers group — full paths are
	// /servers/:serverId/docker/... and /servers/:serverId/pm2/...
	docker := api.Group("/:serverId/docker/containers/:containerId")
	{
		docker.GET("/start", handler.StartContainer)
		docker.GET("/stop", handler.StopContainer)
		docker.GET("/restart", handler.RestartContainer)
		docker.GET("/logs", handler.GetContainerLogs)
	}

	pm2 := api.Group("/:serverId/pm2/services/:id")
	{
		pm2.GET("/start", handler.StartProcess)
		pm2.GET("/stop", handler.StopProcess)
		pm2.GET("/restart", handler.RestartProcess)
		pm2.GET("/logs", handler.GetProcessLogs)
	}
}
