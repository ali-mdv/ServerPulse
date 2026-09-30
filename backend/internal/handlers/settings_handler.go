package handlers

import (
	"net/http"

	"server-monitoring/internal/dtos"
	"server-monitoring/internal/services"

	"github.com/gin-gonic/gin"
)

type SettingsHandler struct {
	service services.SettingsService
}

func NewSettingsHandler(service services.SettingsService) *SettingsHandler {
	return &SettingsHandler{service: service}
}

// Get returns the application settings singleton, creating it with
// defaults on first read.
func (h *SettingsHandler) Get(c *gin.Context) {
	settings, err := h.service.Get(c.Request.Context())
	if err != nil {
		respondServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"settings": dtos.ToSettingsDTO(*settings)})
}

// Update applies a partial update to the settings singleton.
func (h *SettingsHandler) Update(c *gin.Context) {
	var body dtos.UpdateSettingsDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	settings, err := h.service.Update(c.Request.Context(), body)
	if err != nil {
		respondServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"settings": dtos.ToSettingsDTO(*settings)})
}
