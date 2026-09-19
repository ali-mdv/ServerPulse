package handlers

import (
	"net/http"
	"server-monitoring/internal/models"
	"server-monitoring/internal/services"

	"github.com/gin-gonic/gin"
)

type stateHandler struct {
	service services.StateService
}

func NewStateHandler(service services.StateService) *stateHandler {
	return &stateHandler{service: service}
}

func (h *stateHandler) GetServicesState(c *gin.Context) {
	ctx := c.Request.Context()

	pm2State, err := h.service.GetProviderState(ctx, "pm2")
	if err != nil {
		respondHistoryError(c, err)
		return
	}
	dockerState, err := h.service.GetProviderState(ctx, "docker")
	if err != nil {
		respondHistoryError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pm2":    emptyIfNil(pm2State),
		"docker": emptyIfNil(dockerState),
	})
}

func (h *stateHandler) GetSystemState(c *gin.Context) {
	ctx := c.Request.Context()

	hostState, err := h.service.GetHostState(ctx)
	if err != nil {
		respondHistoryError(c, err)
		return
	}

	if hostState == nil {
		c.JSON(http.StatusOK, gin.H{
			"systemUsage": nil,
			"updatedAt":   nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"systemUsage": hostState.Usage,
		"updatedAt":   hostState.UpdatedAt,
	})
}

func emptyIfNil(state *models.ProviderState) *models.ProviderState {
	if state == nil {
		return &models.ProviderState{}
	}
	return state
}
