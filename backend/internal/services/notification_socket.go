package services

import (
	"log"
	"net/http"

	"server-monitoring/internal/models"
	jwtutil "server-monitoring/pkg/jwt"

	"github.com/zishang520/socket.io/v2/socket"
)

// NotificationSocketPath is where the Socket.IO endpoint is mounted. It
// lives under the notifications API so the nginx /api proxy and the
// dashboard's auth model apply unchanged.
const NotificationSocketPath = "/api/v1/notifications/socket.io"

// NotificationEvent is the Socket.IO event name a client subscribes to
// in order to receive alerts.
const NotificationEvent = "notification"

// NotificationEnvelope is the payload pushed over the socket. The
// `type` discriminator lets clients multiplex future message kinds
// (heartbeats, acks, …) without inspecting the notification itself.
type NotificationEnvelope struct {
	Type         string               `json:"type"`
	Notification *models.Notification `json:"notification,omitempty"`
}

// notificationServer wraps the Socket.IO server so the rest of the
// package can broadcast without importing the library directly.
type notificationServer struct {
	io *socket.Server
}

// newNotificationServer builds a Socket.IO server that authenticates
// every handshake with the same JWT the REST API uses. Browsers can't
// set headers on a WebSocket, so the token is read from the Socket.IO
// `auth` payload (preferred) or the `token` query parameter.
func newNotificationServer() *notificationServer {
	opts := socket.DefaultServerOptions()
	opts.SetPath(NotificationSocketPath)
	opts.SetServeClient(false)
	// CORS is handled globally by the Gin middleware; letting Engine.IO
	// add its own headers would duplicate Access-Control-Allow-Origin.
	opts.SetCors(nil)

	io := socket.NewServer(nil, opts)

	io.Use(func(s *socket.Socket, next func(*socket.ExtendedError)) {
		if _, err := jwtutil.ValidateToken(socketToken(s)); err != nil {
			next(socket.NewExtendedError("unauthorized", nil))
			return
		}
		next(nil)
	})

	io.On("connection", func(clients ...any) {
		client, ok := clients[0].(*socket.Socket)
		if !ok {
			return
		}
		log.Printf("notifications: socket connected %s", client.Id())
		client.On("disconnect", func(...any) {
			log.Printf("notifications: socket disconnected %s", client.Id())
		})
	})

	return &notificationServer{io: io}
}

// Handler is the HTTP entry point mounted on the Gin router. The
// Engine.IO server inside routes by the path configured above.
func (n *notificationServer) Handler() http.Handler {
	return n.io.ServeHandler(nil)
}

// Broadcast pushes an alert to every connected client.
func (n *notificationServer) Broadcast(notification models.Notification) {
	if n == nil || n.io == nil {
		return
	}
	n.io.Emit(NotificationEvent, NotificationEnvelope{
		Type:         "notification",
		Notification: &notification,
	})
}

// Close tears the Socket.IO server down.
func (n *notificationServer) Close() {
	if n == nil || n.io == nil {
		return
	}
	n.io.Close(nil)
}

// socketToken extracts the JWT from the Socket.IO handshake. The `auth`
// payload is preferred because it is not logged in URLs.
func socketToken(s *socket.Socket) string {
	handshake := s.Handshake()
	if handshake == nil {
		return ""
	}
	if auth, ok := handshake.Auth.(map[string]any); ok {
		if token, ok := auth["token"].(string); ok && token != "" {
			return token
		}
	}
	if values := handshake.Query["token"]; len(values) > 0 {
		return values[0]
	}
	return ""
}
