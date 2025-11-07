package v1

import (
	"server-monitoring/internal/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(r *gin.RouterGroup) {
	api := r.Group("/users")
	{
		api.GET("", handlers.GetUsers)
	}
}
