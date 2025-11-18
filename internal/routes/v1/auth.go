package v1

import (
	"server-monitoring/internal/handlers"
	"server-monitoring/internal/services"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(r *gin.RouterGroup) {
	userService := services.NewUserService("server_monitoring")
	service := services.NewAuthService(userService)
	handler := handlers.NewAuthHandler(service)

	api := r.Group("/auth")
	{
		api.POST("/login", handler.Login)
	}
}
