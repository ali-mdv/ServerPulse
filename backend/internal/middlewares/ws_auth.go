package middlewares

import (
	"net/http"
	"strings"

	jwtutil "server-monitoring/pkg/jwt"

	"github.com/gin-gonic/gin"
)

// WSAuthMiddleware authenticates a WebSocket handshake. Browsers can't
// set the Authorization header on a WebSocket, so the JWT is accepted
// from the `token` query parameter; the Authorization header is still
// honoured for non-browser clients.
func WSAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			if parts := strings.Split(c.GetHeader("Authorization"), " "); len(parts) == 2 && parts[0] == "Bearer" {
				token = parts[1]
			}
		}
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		claims, err := jwtutil.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		c.Set("userID", claims.UserID)
		c.Next()
	}
}
