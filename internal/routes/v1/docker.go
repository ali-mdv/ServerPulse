package v1

import (
	"server-monitoring/internal/handlers"
	"server-monitoring/internal/middlewares"
	"server-monitoring/internal/services"

	"github.com/gin-gonic/gin"
)

func RegisterDockerRoutes(r *gin.RouterGroup) {
	service := services.NewDockerService()
	handler := handlers.NewDockerHandler(service)

	api := r.Group("/docker")
	api.Use(middlewares.AuthMiddleware())

	{
		api.GET("images", handler.GetDockerImages)
		api.GET("containers", handler.GetDockerContainers)
	}
}
