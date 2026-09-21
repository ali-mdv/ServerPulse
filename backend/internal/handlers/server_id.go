package handlers

import (
	"context"

	"server-monitoring/internal/models"
	"server-monitoring/internal/services"

	"github.com/gin-gonic/gin"
)

// serverIDFromQuery returns the ?serverId value from the request, resolving
// the sentinel "local" name to the actual local server's hex ID so that the
// in-process scheduler's data (written under the real row's ID) is found.
// If the local server cannot be resolved, it falls back to the legacy
// "local" sentinel so existing data isn't orphaned.
func serverIDFromQuery(c *gin.Context, serverSvc services.ServerService) string {
	id := c.Query("serverId")
	if id != "" && id != models.LocalServerName {
		return id
	}
	return resolveLocalServerID(c.Request.Context(), serverSvc)
}

func resolveLocalServerID(ctx context.Context, serverSvc services.ServerService) string {
	if serverSvc == nil {
		return services.LocalServerID
	}
	srv, err := serverSvc.EnsureLocalServer(ctx)
	if err != nil || srv == nil {
		return services.LocalServerID
	}
	return srv.HexID()
}
