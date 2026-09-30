package services

import (
	"context"
	"log"
	"net/http"

	"server-monitoring/internal/models"

	"github.com/zishang520/socket.io/v2/socket"
)

// AgentSocketPath is where the agent-facing Socket.IO endpoint is
// mounted. It mirrors NotificationSocketPath so the nginx /api proxy
// and the agents domain apply unchanged.
const AgentSocketPath = "/api/v1/agents/socket.io"

// agentSocket wraps the Socket.IO server remote agents connect to for
// bidirectional command delivery. Auth is the AgentToken (same identity
// as POST /agents/push), carried in the handshake `auth` payload or the
// `token` query parameter — browsers and many HTTP clients cannot set
// headers on a WebSocket handshake.
type agentSocket struct {
	io      *socket.Server
	servers ServerService
	hub     *AgentHub
}

// NewAgentSocket builds the Socket.IO server used for the control
// channel. It is a separate Engine.IO instance from notifications so
// the two auth domains (user JWT vs AgentToken) never share middleware.
func NewAgentSocket(servers ServerService, hub *AgentHub) *agentSocket {
	opts := socket.DefaultServerOptions()
	opts.SetPath(AgentSocketPath)
	opts.SetServeClient(false)
	// CORS is handled globally by the Gin middleware.
	opts.SetCors(nil)

	io := socket.NewServer(nil, opts)

	io.Use(func(s *socket.Socket, next func(*socket.ExtendedError)) {
		token := agentSocketToken(s)
		if token == "" {
			next(socket.NewExtendedError("unauthorized", nil))
			return
		}
		ctx := context.Background()
		server, err := servers.ResolveByAgentToken(ctx, token)
		if err != nil || server == nil {
			next(socket.NewExtendedError("unauthorized", nil))
			return
		}
		// Stash the resolved server for the connection handler.
		s.SetData(server.HexID())
		next(nil)
	})

	io.On("connection", func(clients ...any) {
		client, ok := clients[0].(*socket.Socket)
		if !ok {
			return
		}
		serverID, _ := client.Data().(string)
		if serverID == "" {
			log.Printf("agent socket: connection without server id, closing")
			client.Disconnect(true)
			return
		}
		log.Printf("agent socket: connected server=%s", serverID)
		hub.Register(serverID, client)
		client.On("disconnect", func(...any) {
			log.Printf("agent socket: disconnected server=%s", serverID)
			hub.Unregister(serverID, client)
		})
	})

	return &agentSocket{io: io, servers: servers, hub: hub}
}

// Handler is the HTTP entry point mounted on the Gin router.
func (a *agentSocket) Handler() http.Handler {
	return a.io.ServeHandler(nil)
}

// Close tears the Socket.IO server down.
func (a *agentSocket) Close() {
	if a == nil || a.io == nil {
		return
	}
	a.io.Close(nil)
}

// agentSocketToken extracts the AgentToken from the handshake. The
// `auth` payload is preferred (not logged in URLs); `token` / `apiKey`
// query params are accepted as fallbacks for simpler clients.
func agentSocketToken(s *socket.Socket) string {
	handshake := s.Handshake()
	if handshake == nil {
		return ""
	}
	if auth, ok := handshake.Auth.(map[string]any); ok {
		for _, key := range []string{"token", "apiKey"} {
			if v, ok := auth[key].(string); ok && v != "" {
				return v
			}
		}
	}
	for _, key := range []string{"token", "apiKey"} {
		if values := handshake.Query[key]; len(values) > 0 && values[0] != "" {
			return values[0]
		}
	}
	return ""
}

// Ensure agentSocket is used even if constructors are refactored later.
var _ = (*agentSocket)(nil)

// ServerHexID is a tiny helper so handlers can read a server's hex ID
// without importing models in every call site.
func ServerHexID(s *models.Server) string {
	if s == nil {
		return ""
	}
	return s.HexID()
}
