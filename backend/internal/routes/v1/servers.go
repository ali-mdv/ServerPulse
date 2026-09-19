package v1

import (
	"server-monitoring/internal/handlers"
	"server-monitoring/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterServerRoutes(r *gin.RouterGroup, s *Services) {
	handler := handlers.NewServerHandler(s.Servers, s.State)

	api := r.Group("/servers")
	api.Use(middlewares.AuthMiddleware())

	{
		api.GET("", handler.List)
		api.POST("", handler.Create)
		api.GET("/:serverId", handler.Get)
		api.PUT("/:serverId", handler.Update)
		api.DELETE("/:serverId", handler.Delete)
	}
}