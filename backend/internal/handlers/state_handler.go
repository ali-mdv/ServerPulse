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

// serverIDOrLocal resolves the server scope for state endpoints. It
// accepts ?serverId=... and falls back to the local-server sentinel so
// the dashboard can keep polling without knowing about multi-server.
func serverIDOrLocal(c *gin.Context) string {
	if id := c.Query("serverId"); id != "" {
		return id
	}
	return services.LocalServerID
}

func (h *stateHandler) GetServicesState(c *gin.Context) {
	ctx := c.Request.Context()
	serverID := serverIDOrLocal(c)

	pm2State, err := h.service.GetProviderState(ctx, serverID, "pm2")
	if err != nil {
		respondHistoryError(c, err)
		return
	}
	dockerState, err := h.service.GetProviderState(ctx, serverID, "docker")
	if err != nil {
		respondHistoryError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"serverId": serverID,
		"pm2":      emptyIfNil(pm2State),
		"docker":   emptyIfNil(dockerState),
	})
}

func (h *stateHandler) GetSystemState(c *gin.Context) {
	ctx := c.Request.Context()
	serverID := serverIDOrLocal(c)

	usageState, err := h.service.GetServerUsage(ctx, serverID)
	if err != nil {
		respondHistoryError(c, err)
		return
	}

	if usageState == nil {
		c.JSON(http.StatusOK, gin.H{
			"serverId":     serverID,
			"systemUsage":  nil,
			"updatedAt":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"serverId":    usageState.ServerID,
		"systemUsage": usageState.Usage,
		"updatedAt":   usageState.UpdatedAt,
	})
}

func emptyIfNil(state *models.ProviderState) *models.ProviderState {
	if state == nil {
		return &models.ProviderState{}
	}
	return state
}