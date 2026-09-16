package handlers

import (
	"net/http"
	"server-monitoring/internal/services"

	"github.com/gin-gonic/gin"
)

type reportHandler struct {
	service services.ReportService
}

func NewReportHandler(service services.ReportService) *reportHandler {
	return &reportHandler{service: service}
}

func (h *reportHandler) GetReports(c *gin.Context) {
	reports := h.service.ReportsList()
	c.JSON(http.StatusOK, gin.H{
		"reports": reports,
	})
}
