package v1

import (
	"server-monitoring/internal/handlers"
	"server-monitoring/internal/services"
	database "server-monitoring/pkg/mongo"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(r *gin.RouterGroup) {
	db := database.GetCollection("server_monitoring", "users")
	userService := services.NewUserService(db)
	service := services.NewAuthService(userService)
	handler := handlers.NewAuthHandler(service)

	api := r.Group("/auth")
	{
		api.POST("/login", handler.Login)
	}
}
