package v1

import (
	"server-monitoring/internal/handlers"
	"server-monitoring/internal/services"
	database "server-monitoring/pkg/mongo"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(r *gin.RouterGroup) {
	db := database.GetCollection("server_monitoring", "users")
	service := services.NewUserService(db)
	handler := handlers.NewUserHandler(service)

	api := r.Group("/users")
	{
		api.GET("", handler.GetUsers)
	}
}
