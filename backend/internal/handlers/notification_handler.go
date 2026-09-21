package handlers

import (
	"net/http"
	"strconv"

	"server-monitoring/internal/models"
	"server-monitoring/internal/services"
	"server-monitoring/internal/utils"
	"server-monitoring/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type NotificationHandler struct {
	service  services.NotificationService
	servers  services.ServerService
	upgrader websocket.Upgrader
}

func NewNotificationHandler(service services.NotificationService, servers services.ServerService) *NotificationHandler {
	return &NotificationHandler{
		service: service,
		servers: servers,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     checkWebSocketOrigin,
		},
	}
}

// ServeWS upgrades the request and registers the connection with the
// notification hub. Auth is enforced by WSAuthMiddleware before the
// upgrade happens.
func (h *NotificationHandler) ServeWS(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// Upgrade already wrote the HTTP error response.
		return
	}

	client := services.NewNotificationClient(conn, h.service.Hub())
	h.service.Hub().Register(client)
	client.Run()
}

// List returns notifications newest-first, optionally filtered by
// server, severity and read state.
func (h *NotificationHandler) List(c *gin.Context) {
	filter := services.NotificationFilter{
		ServerID:   h.optionalServerID(c),
		Severity:   models.NotificationSeverity(c.Query("severity")),
		UnreadOnly: c.Query("unread") == "true",
	}
	if limit, err := strconv.Atoi(c.Query("limit")); err == nil {
		filter.Limit = limit
	}

	items, err := h.service.List(c.Request.Context(), filter)
	if err != nil {
		respondServerError(c, err)
		return
	}
	if items == nil {
		items = []models.Notification{}
	}

	c.JSON(http.StatusOK, gin.H{"notifications": items})
}

// MarkRead acknowledges a single notification.
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	id, err := bson.ObjectIDFromHex(c.Param("notificationId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification id"})
		return
	}

	if err := h.service.MarkRead(c.Request.Context(), id); err != nil {
		respondServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "notification marked read"})
}

// MarkAllRead acknowledges every unread notification for the server.
func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	updated, err := h.service.MarkAllRead(c.Request.Context(), h.optionalServerID(c))
	if err != nil {
		respondServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"updated": updated})
}

// UnreadCount powers the nav badge.
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	count, err := h.service.UnreadCount(c.Request.Context(), h.optionalServerID(c))
	if err != nil {
		respondServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"unread": count})
}

// optionalServerID resolves ?serverId when present, including the
// "local" sentinel. An absent/empty value means "every server" so the
// panel can show alerts across the fleet.
func (h *NotificationHandler) optionalServerID(c *gin.Context) string {
	id := c.Query("serverId")
	if id == "" || id == "all" {
		return ""
	}
	if id == models.LocalServerName {
		return utils.ServerIDFromQuery(c, h.servers)
	}
	return id
}

// checkWebSocketOrigin mirrors the CORS config: wildcard (or no
// configured origins) allows everything, otherwise the Origin header
// must match. Non-browser clients with no Origin are allowed.
func checkWebSocketOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	cfg := config.Cfg
	if cfg == nil || len(cfg.Origins) == 0 {
		return true
	}
	for _, allowed := range cfg.Origins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}
