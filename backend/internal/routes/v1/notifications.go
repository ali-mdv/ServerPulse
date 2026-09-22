package v1

import (
	"server-monitoring/internal/handlers"
	"server-monitoring/internal/middlewares"
	"server-monitoring/internal/services"

	"github.com/gin-gonic/gin"
)

func RegisterNotificationRoutes(r *gin.RouterGroup, notifications services.NotificationService, servers services.ServerService) {
	handler := handlers.NewNotificationHandler(notifications, servers)

	api := r.Group("/notifications")
	api.Use(middlewares.AuthMiddleware())

	{
		api.GET("", handler.List)
		api.GET("/unread-count", handler.UnreadCount)
		api.POST("/read-all", handler.MarkAllRead)
		api.POST("/:notificationId/read", handler.MarkRead)
	}

	// Socket.IO endpoint. Engine.IO's handshake authenticates via the
	// socket `auth` payload / `token` query, so it bypasses the header
	// auth used by the REST routes above.
	r.Any("/notifications/socket.io/*any", gin.WrapH(notifications.Handler()))
}
