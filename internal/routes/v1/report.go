package v1

import (
	"server-monitoring/internal/handlers"
	"server-monitoring/internal/middlewares"
	"server-monitoring/internal/services"
	database "server-monitoring/pkg/mongo"

	"github.com/gin-gonic/gin"
)

func RegisterReportRoutes(r *gin.RouterGroup) {
	db := database.GetCollection("server_monitoring", "reports")
	service := services.NewReportService(db)
	handler := handlers.NewReportHandler(service)

	api := r.Group("/reports")
	api.Use(middlewares.AuthMiddleware())

	{
		api.GET("", handler.GetReports)
	}
}
