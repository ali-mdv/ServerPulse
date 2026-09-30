package v1

import (
	"server-monitoring/internal/handlers"
	"server-monitoring/internal/middlewares"

	"github.com/gin-gonic/gin"
)

// RegisterAgentRoutes mounts the agent-facing endpoints. These routes
// sit under /v1/agents and are protected by AgentAuthMiddleware
// (Bearer AgentToken) — not AuthMiddleware (user JWT). The dashboard
// itself does not call them; remote agents do, on every cycle.
//
// The Socket.IO control channel is mounted alongside (outside the
// middleware group) because its handshake authenticates via the
// `auth` payload / `token` query, not an Authorization header.
func RegisterAgentRoutes(r *gin.RouterGroup, s *Services) {
	handler := handlers.NewAgentHandler(s.Agent)

	api := r.Group("/agents")
	api.Use(middlewares.AgentAuthMiddleware(s.Servers))

	{
		api.POST("/push", handler.Push)
	}

	if s.AgentSocket != nil {
		r.Any("/agents/socket.io/*any", gin.WrapH(s.AgentSocket.Handler()))
	}
}
