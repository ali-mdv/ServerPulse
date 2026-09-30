package handlers

import (
	"net/http"
	"strconv"

	"server-monitoring/internal/models"
	"server-monitoring/internal/services"
	"server-monitoring/internal/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type NotificationHandler struct {
	service services.NotificationService
	servers services.ServerService
}

func NewNotificationHandler(service services.NotificationService, servers services.ServerService) *NotificationHandler {
	return &NotificationHandler{service: service, servers: servers}
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
