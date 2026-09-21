package v1

import (
	"server-monitoring/internal/handlers"
	"server-monitoring/internal/middlewares"
	"server-monitoring/internal/services"

	"github.com/gin-gonic/gin"
)

func RegisterHistoryRoutes(r *gin.RouterGroup, history services.HistoryService, servers services.ServerService) {
	handler := handlers.NewHistoryHandler(history, servers)

	api := r.Group("/history")
	api.Use(middlewares.AuthMiddleware())

	{
		api.GET("/:provider/services", handler.ListServices)
		api.GET("/:provider/services/:serviceId", handler.GetSeries)
	}
}
