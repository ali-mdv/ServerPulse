package middlewares

import (
	"errors"
	"net/http"
	"strings"

	"server-monitoring/internal/models"
	"server-monitoring/internal/services"

	"github.com/gin-gonic/gin"
)

// agentContextKey is the gin.Context key under which the resolved
// Server is stored by AgentAuthMiddleware.
const agentContextKey = "agentServer"

// AgentFromContext returns the Server attached by AgentAuthMiddleware.
// Returns nil when the middleware did not run on this route.
func AgentFromContext(c *gin.Context) *models.Server {
	if v, ok := c.Get(agentContextKey); ok {
		if s, ok := v.(*models.Server); ok {
			return s
		}
	}
	return nil
}

// AgentAuthMiddleware validates the Bearer AgentToken presented by a
// remote agent and attaches the resolved Server to the request context.
//
// It must be installed *without* AuthMiddleware on the same route: the
// agent identity is the token, not a user JWT.
func AgentAuthMiddleware(serverSvc services.ServerService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			unauthorized(c, "missing Authorization header")
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			unauthorized(c, "Authorization must be Bearer <token>")
			return
		}

		token := strings.TrimSpace(parts[1])
		if token == "" {
			unauthorized(c, "empty bearer token")
			return
		}

		server, err := serverSvc.ResolveByAgentToken(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "agent auth: lookup failed"})
			return
		}
		if server == nil {
			unauthorized(c, "unknown agent token")
			return
		}

		c.Set(agentContextKey, server)
		c.Next()
	}
}

func unauthorized(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": msg})
}

// ErrNoAgentServer is returned by handlers that require AgentAuthMiddleware
// but ran without it (defensive — should never happen in normal flow).
var ErrNoAgentServer = errors.New("agent: no server in context")