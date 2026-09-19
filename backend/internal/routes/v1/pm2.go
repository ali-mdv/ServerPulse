package v1

import (
	"server-monitoring/internal/handlers"
	"server-monitoring/internal/middlewares"
	"server-monitoring/internal/services"

	"github.com/gin-gonic/gin"
)

func RegisterPM2Routes(r *gin.RouterGroup, service services.PM2Service) {
	handler := handlers.NewPM2Handler(service)

	api := r.Group("/pm2")
	api.Use(middlewares.AuthMiddleware())

	{
		api.GET("/services", handler.ProcessList)
		api.GET("/services/:id/inspect", handler.ProcessDetail)
		api.GET("/services/:id/start", handler.StartProcess)
		api.GET("/services/:id/stop", handler.StopProcess)
		api.GET("/services/:id/restart", handler.RestartProcess)
		api.GET("/services/:id/logs", handler.GetProcessLogs)
	}

}
