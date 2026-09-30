package handlers

import (
	"net/http"

	"server-monitoring/internal/dtos"
	"server-monitoring/internal/middlewares"
	"server-monitoring/internal/services"

	"github.com/gin-gonic/gin"
)

type AgentHandler struct {
	agent services.AgentService
}

func NewAgentHandler(agent services.AgentService) *AgentHandler {
	return &AgentHandler{agent: agent}
}

// Push is the agent's cycle endpoint. The AgentAuthMiddleware that
// wraps this route resolves the bearer token to a Server and stashes
// it in the gin context; we pick it up here.
//
// The handler always returns 200 once the payload has been accepted
// for processing. Per-provider write errors are logged but do not
// cause a 5xx — the dashboard should still see the usage and any
// provider that did report back, and the agent will retry the missing
// data on the next tick anyway.
func (h *AgentHandler) Push(c *gin.Context) {
	server := middlewares.AgentFromContext(c)
	if server == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "agent middleware did not attach server"})
		return
	}

	var push dtos.AgentPushDTO
	if err := c.ShouldBindJSON(&push); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if err := h.agent.IngestPush(c.Request.Context(), *server, push); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"serverId":  server.HexID(),
		"accepted":  true,
		"providers": len(push.Providers),
	})
}