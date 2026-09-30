package handlers

import (
	"net/http"
	"strconv"

	"server-monitoring/internal/services"
	apperrors "server-monitoring/pkg/errors"

	"github.com/gin-gonic/gin"
)

// ServerControlHandler exposes server-scoped Docker/PM2 control under
// /servers/:serverId/.... Local targets short-circuit to the in-process
// services; remote targets are forwarded to the agent over Socket.IO.
//
// Actions are GET to match the existing /docker and /pm2 routes so the
// frontend can reuse the same call shape.
type ServerControlHandler struct {
	commands services.CommandService
}

func NewServerControlHandler(commands services.CommandService) *ServerControlHandler {
	return &ServerControlHandler{commands: commands}
}

// --- Docker ---

func (h *ServerControlHandler) StartContainer(c *gin.Context) {
	h.runAction(c, services.ProviderDocker, services.ActionStart, "container successfully started")
}

func (h *ServerControlHandler) StopContainer(c *gin.Context) {
	h.runAction(c, services.ProviderDocker, services.ActionStop, "container successfully stopped")
}

func (h *ServerControlHandler) RestartContainer(c *gin.Context) {
	h.runAction(c, services.ProviderDocker, services.ActionRestart, "container successfully restarted")
}

func (h *ServerControlHandler) GetContainerLogs(c *gin.Context) {
	h.runLogs(c, services.ProviderDocker)
}

// --- PM2 ---

func (h *ServerControlHandler) StartProcess(c *gin.Context) {
	h.runAction(c, services.ProviderPM2, services.ActionStart, "process successfully started")
}

func (h *ServerControlHandler) StopProcess(c *gin.Context) {
	h.runAction(c, services.ProviderPM2, services.ActionStop, "process successfully stopped")
}

func (h *ServerControlHandler) RestartProcess(c *gin.Context) {
	h.runAction(c, services.ProviderPM2, services.ActionRestart, "process successfully restarted")
}

func (h *ServerControlHandler) GetProcessLogs(c *gin.Context) {
	// Validate numeric id early so the local and remote paths share the
	// same 400 as the unscoped PM2 routes.
	idStr := c.Param("id")
	if _, err := strconv.ParseInt(idStr, 10, 0); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid PM2 id; expected a numeric value",
		})
		return
	}
	h.runLogs(c, services.ProviderPM2)
}

// --- helpers ---

func (h *ServerControlHandler) runAction(c *gin.Context, provider, action, successMessage string) {
	serverID := c.Param("serverId")
	target := actionTarget(c, provider)

	if err := h.commands.ExecuteAction(c.Request.Context(), serverID, provider, action, target); err != nil {
		writeCommandError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": successMessage})
}

func (h *ServerControlHandler) runLogs(c *gin.Context, provider string) {
	serverID := c.Param("serverId")
	target := actionTarget(c, provider)

	logs, err := h.commands.ExecuteLogs(c.Request.Context(), serverID, provider, target, 100)
	if err != nil {
		writeCommandError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"logs": logs})
}

func actionTarget(c *gin.Context, provider string) string {
	if provider == services.ProviderDocker {
		target, _ := c.Params.Get("containerId")
		return target
	}
	return c.Param("id")
}

func writeCommandError(c *gin.Context, err error) {
	if e, ok := err.(*apperrors.AppError); ok {
		c.JSON(e.Code, gin.H{"error": e.Message})
		return
	}
	c.JSON(apperrors.ErrInternalServer.Code, gin.H{
		"error": apperrors.ErrInternalServer.Message,
	})
}
