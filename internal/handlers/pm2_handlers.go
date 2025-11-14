package handlers

import (
	"net/http"
	"server-monitoring/internal/services"

	"github.com/gin-gonic/gin"
)

type PM2Handler struct {
	service services.PM2Service
}

func NewPM2Handler(service services.PM2Service) *PM2Handler {
	return &PM2Handler{service: service}
}

func (h *PM2Handler) ProcessList(c *gin.Context) {
	processes, err := h.service.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"processes": processes,
	})
}
