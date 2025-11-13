package v1

import (
	"server-monitoring/internal/handlers"
	"server-monitoring/internal/middlewares"
	"server-monitoring/internal/services"

	"github.com/gin-gonic/gin"
)

func RegisterSystemRoutes(r *gin.RouterGroup) {
	service := services.NewSystemService()
	handler := handlers.NewSystemHandler(service)

	api := r.Group("/system")
	api.Use(middlewares.AuthMiddleware())

	{
		api.GET("/usage", handler.GetSystemUsage)
	}
}
