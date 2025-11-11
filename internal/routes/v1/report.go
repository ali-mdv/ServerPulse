package v1

import (
	"server-monitoring/internal/handlers"
	"server-monitoring/internal/middlewares"
	"server-monitoring/internal/services"

	"github.com/gin-gonic/gin"
)

func RegisterReportRoutes(r *gin.RouterGroup) {
	service := services.NewReportService("server_monitoring")
	handler := handlers.NewReportHandler(service)

	api := r.Group("/reports")
	api.Use(middlewares.AuthMiddleware())

	{
		api.GET("", handler.GetReports)
	}
}
