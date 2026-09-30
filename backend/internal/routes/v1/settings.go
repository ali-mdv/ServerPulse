package v1

import (
	"server-monitoring/internal/handlers"
	"server-monitoring/internal/middlewares"
	"server-monitoring/internal/services"

	"github.com/gin-gonic/gin"
)

func RegisterSettingsRoutes(r *gin.RouterGroup, settings services.SettingsService) {
	handler := handlers.NewSettingsHandler(settings)

	api := r.Group("/settings")
	api.Use(middlewares.AuthMiddleware())

	{
		api.GET("", handler.Get)
		api.PUT("", handler.Update)
		api.PATCH("", handler.Update)
	}
}
