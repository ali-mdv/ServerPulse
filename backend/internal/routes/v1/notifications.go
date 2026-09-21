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

	// Browsers can't set the Authorization header on a WebSocket
	// handshake, so this route authenticates via ?token= instead.
	r.GET("/notifications/ws", middlewares.WSAuthMiddleware(), handler.ServeWS)
}
