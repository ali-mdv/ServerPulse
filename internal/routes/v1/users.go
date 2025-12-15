package v1

import (
	"server-monitoring/internal/handlers"
	"server-monitoring/internal/middlewares"
	"server-monitoring/internal/services"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(r *gin.RouterGroup) {
	service := services.NewUserService("server_monitoring")
	handler := handlers.NewUserHandler(service)

	api := r.Group("/users")
	api.Use(middlewares.AuthMiddleware())

	{
		api.GET("", handler.GetUsers)
		api.POST("", handler.CreateUser)
		api.GET("/:userID", handler.GetUserByID)
		api.PUT("/:userID", handler.UpdateUser)
	}
}
