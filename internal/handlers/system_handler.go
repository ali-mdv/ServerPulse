package handlers

import (
	"net/http"
	"server-monitoring/internal/services"

	"github.com/gin-gonic/gin"
)

type systemHandler struct {
	service services.SystemService
}

func NewSystemHandler(service services.SystemService) *systemHandler {
	return &systemHandler{service: service}
}

func (h *systemHandler) GetSystemUsage(c *gin.Context) {
	systemUsage := h.service.SystemUsage()
	c.JSON(http.StatusOK, gin.H{
		"systemUsage": systemUsage,
	})
}
