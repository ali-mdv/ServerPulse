package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"server-monitoring/internal/dtos"
	apperrors "server-monitoring/pkg/errors"

	"github.com/zishang520/socket.io/v2/socket"
)

// AgentCommandEvent is the Socket.IO event the backend emits to push a
// command to an agent. The agent must reply via the emit's ack callback.
const AgentCommandEvent = "agent:command"

// DefaultAgentCommandTimeout bounds how long Execute waits for an agent
// ack. Callers may pass a shorter/longer timeout (logs get more headroom).
const DefaultAgentCommandTimeout = 10 * time.Second

// AgentHub tracks one live Socket.IO connection per server and runs
// synchronous request/response commands against them.
//
// A reconnecting agent replaces its previous socket (the old one is
// disconnected), so at most one connection exists per serverID.
type AgentHub struct {
	mu    sync.RWMutex
	conns map[string]*socket.Socket
}

func NewAgentHub() *AgentHub {
	return &AgentHub{conns: make(map[string]*socket.Socket)}
}

// Register binds sock as the live connection for serverID. Any previous
// connection for the same server is disconnected so commands never race
// two sockets for one agent.
func (h *AgentHub) Register(serverID string, sock *socket.Socket) {
	if h == nil || sock == nil || serverID == "" {
		return
	}
	h.mu.Lock()
	if old, ok := h.conns[serverID]; ok && old != nil && old != sock {
		old.Disconnect(true)
	}
	h.conns[serverID] = sock
	h.mu.Unlock()
}

// Unregister removes sock only if it is still the registered connection
// for serverID (a reconnect may have already replaced it).
func (h *AgentHub) Unregister(serverID string, sock *socket.Socket) {
	if h == nil || sock == nil || serverID == "" {
		return
	}
	h.mu.Lock()
	if cur, ok := h.conns[serverID]; ok && cur == sock {
		delete(h.conns, serverID)
	}
	h.mu.Unlock()
}

// IsConnected reports whether serverID currently has a live agent socket.
func (h *AgentHub) IsConnected(serverID string) bool {
	if h == nil {
		return false
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	sock, ok := h.conns[serverID]
	return ok && sock != nil && sock.Connected()
}

// ConnectedServerIDs returns a snapshot of serverIDs with a live socket.
func (h *AgentHub) ConnectedServerIDs() []string {
	if h == nil {
		return nil
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	ids := make([]string, 0, len(h.conns))
	for id, sock := range h.conns {
		if sock != nil && sock.Connected() {
			ids = append(ids, id)
		}
	}
	return ids
}

type agentAck struct {
	args []any
	err  error
}

// Execute sends cmd to the agent for serverID and blocks until the agent
// acks, ctx is done, or the Socket.IO ack timeout fires.
//
// Errors:
//   - 503 agent not connected
//   - 504 ack timeout / ctx deadline
//   - 502 transport-level Socket.IO error
//   - agent-side failure is returned as a non-AppError with the agent's message
//     (handlers map "not found" style messages to 404).
func (h *AgentHub) Execute(
	ctx context.Context,
	serverID string,
	cmd dtos.AgentCommand,
	timeout time.Duration,
) (dtos.AgentCommandResult, error) {
	if h == nil {
		return dtos.AgentCommandResult{}, apperrors.New(503, "agent hub not initialised")
	}
	if timeout <= 0 {
		timeout = DefaultAgentCommandTimeout
	}

	h.mu.RLock()
	sock, ok := h.conns[serverID]
	h.mu.RUnlock()
	if !ok || sock == nil || !sock.Connected() {
		return dtos.AgentCommandResult{}, apperrors.New(503, "agent not connected")
	}

	if cmd.ID == "" {
		cmd.ID = fmt.Sprintf("%s-%d", serverID, time.Now().UnixNano())
	}

	done := make(chan agentAck, 1)
	// Timeout() sets the Socket.IO ack deadline so a dead agent cannot
	// leave this goroutine parked forever even if ctx has no deadline.
	emitErr := sock.Timeout(timeout).Emit(AgentCommandEvent, cmd, func(args []any, err error) {
		done <- agentAck{args: args, err: err}
	})
	if emitErr != nil {
		return dtos.AgentCommandResult{}, apperrors.New(502, fmt.Sprintf("agent command send failed: %s", emitErr))
	}

	select {
	case <-ctx.Done():
		return dtos.AgentCommandResult{}, apperrors.New(504, "agent command cancelled")
	case res := <-done:
		return parseAgentAck(cmd.ID, res)
	}
}

func parseAgentAck(cmdID string, res agentAck) (dtos.AgentCommandResult, error) {
	if res.err != nil {
		msg := res.err.Error()
		if strings.Contains(msg, "timed out") || strings.Contains(msg, "timeout") {
			return dtos.AgentCommandResult{}, apperrors.New(504, "agent command timed out")
		}
		return dtos.AgentCommandResult{}, apperrors.New(502, "agent command failed: "+msg)
	}
	if len(res.args) == 0 {
		return dtos.AgentCommandResult{}, apperrors.New(502, "agent returned an empty ack")
	}

	result, err := coerceCommandResult(res.args[0])
	if err != nil {
		return dtos.AgentCommandResult{}, apperrors.New(502, "agent returned an unreadable ack: "+err.Error())
	}
	if result.ID == "" {
		result.ID = cmdID
	}
	return result, nil
}

// coerceCommandResult normalises the ack payload. Depending on the
// Socket.IO parser, the agent's object arrives either as a typed struct
// (Go client on the same process graph) or as map[string]any.
func coerceCommandResult(v any) (dtos.AgentCommandResult, error) {
	switch t := v.(type) {
	case dtos.AgentCommandResult:
		return t, nil
	case *dtos.AgentCommandResult:
		if t == nil {
			return dtos.AgentCommandResult{}, fmt.Errorf("nil result")
		}
		return *t, nil
	default:
		raw, err := json.Marshal(v)
		if err != nil {
			return dtos.AgentCommandResult{}, err
		}
		var out dtos.AgentCommandResult
		if err := json.Unmarshal(raw, &out); err != nil {
			return dtos.AgentCommandResult{}, err
		}
		return out, nil
	}
}
