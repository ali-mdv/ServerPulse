package v1

import (
	"server-monitoring/internal/handlers"
	"server-monitoring/internal/middlewares"
	"server-monitoring/internal/services"

	"github.com/gin-gonic/gin"
)

func RegisterStateRoutes(r *gin.RouterGroup, state services.StateService, servers services.ServerService) {
	handler := handlers.NewStateHandler(state, servers)

	api := r.Group("/state")
	api.Use(middlewares.AuthMiddleware())

	{
		api.GET("/services", handler.GetServicesState)
		api.GET("/system", handler.GetSystemState)
	}
}
