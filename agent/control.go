// ServerPulse agent control channel (Socket.IO).
//
// Connects to the backend's /api/v1/agents/socket.io endpoint and
// executes start/stop/restart/logs commands for Docker and PM2,
// acking each command with an AgentCommandResult.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/zishang520/engine.io-client-go/transports"
	enginatypes "github.com/zishang520/engine.io/v2/types"
	clientsocket "github.com/zishang520/socket.io-client-go/socket"
	serversocket "github.com/zishang520/socket.io/v2/socket"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	dockerclient "github.com/docker/docker/client"
)

// AgentCommand mirrors backend/internal/dtos/agent_command_dto.go.
type AgentCommand struct {
	ID       string         `json:"id"`
	Provider string         `json:"provider"`
	Action   string         `json:"action"`
	Target   string         `json:"target"`
	Params   map[string]any `json:"params,omitempty"`
}

// AgentCommandResult mirrors the backend's result envelope.
type AgentCommandResult struct {
	ID    string          `json:"id"`
	OK    bool            `json:"ok"`
	Error string          `json:"error,omitempty"`
	Data  json.RawMessage `json:"data,omitempty"`
}

// LogsData is the Data payload for action=logs.
type LogsData struct {
	Logs string `json:"logs"`
}

const agentCommandEvent = "agent:command"

// startControlChannel opens a Socket.IO connection that receives
// commands from the backend. It blocks forever, reconnecting as needed
// (the client library handles backoff when reconnection is enabled).
func startControlChannel(baseURL, apiKey string) {
	opts := clientsocket.DefaultOptions()
	opts.SetPath("/api/v1/agents/socket.io")
	opts.SetTransports(enginatypes.NewSet(transports.WebSocket, transports.Polling))
	opts.SetReconnection(true)
	opts.SetAuth(map[string]any{"token": apiKey, "apiKey": apiKey})

	manager := clientsocket.NewManager(strings.TrimRight(baseURL, "/"), opts)
	manager.On("reconnect", func(...any) {
		log.Printf("control channel: reconnected")
	})
	manager.On("reconnect_failed", func(...any) {
		log.Printf("control channel: reconnection failed; will keep retrying")
	})

	sio := manager.Socket("/", opts)

	sio.On("connect", func(args ...any) {
		log.Printf("control channel: connected id=%s", sio.Id())
	})

	sio.On("connect_error", func(args ...any) {
		log.Printf("control channel: connect error: %v", args)
	})

	sio.On("disconnect", func(args ...any) {
		log.Printf("control channel: disconnected: %v", args)
	})

	sio.On(agentCommandEvent, func(args ...any) {
		ack, hasAck := extractAck(args)
		payload, ok := firstPayload(args)
		if !ok {
			log.Printf("control channel: command with no payload: %v", args)
			if hasAck {
				ack([]any{AgentCommandResult{OK: false, Error: "malformed command"}}, nil)
			}
			return
		}

		cmd, err := decodeCommand(payload)
		if err != nil {
			log.Printf("control channel: decode command: %v", err)
			if hasAck {
				ack([]any{AgentCommandResult{OK: false, Error: "malformed command: " + err.Error()}}, nil)
			}
			return
		}

		log.Printf("control channel: command id=%s provider=%s action=%s target=%s",
			cmd.ID, cmd.Provider, cmd.Action, cmd.Target)

		result := executeCommand(cmd)
		if !hasAck {
			log.Printf("control channel: command %s finished but no ack available", cmd.ID)
			return
		}
		ack([]any{result}, nil)
	})

	// Socket() with autoConnect (default) starts connecting; park forever.
	select {}
}

func extractAck(args []any) (func([]any, error), bool) {
	if len(args) == 0 {
		return nil, false
	}
	ack, ok := args[len(args)-1].(serversocket.Ack)
	if !ok {
		return nil, false
	}
	return ack, true
}

func firstPayload(args []any) (any, bool) {
	if len(args) == 0 {
		return nil, false
	}
	// Trailing arg may be the ack — payload is everything before it.
	last := args[len(args)-1]
	if _, isAck := last.(serversocket.Ack); isAck {
		if len(args) == 1 {
			return nil, false
		}
		return args[0], true
	}
	return args[0], true
}

func decodeCommand(payload any) (AgentCommand, error) {
	var cmd AgentCommand
	raw, err := json.Marshal(payload)
	if err != nil {
		return cmd, err
	}
	if err := json.Unmarshal(raw, &cmd); err != nil {
		return cmd, err
	}
	return cmd, nil
}

func executeCommand(cmd AgentCommand) AgentCommandResult {
	result := AgentCommandResult{ID: cmd.ID}
	switch cmd.Provider {
	case "docker":
		result = runDockerCommand(cmd)
	case "pm2":
		result = runPM2Command(cmd)
	default:
		result.Error = "unknown provider: " + cmd.Provider
	}
	result.ID = cmd.ID
	return result
}

// --- Docker control ---

func newDockerClient() (*dockerclient.Client, error) {
	return dockerclient.NewClientWithOpts(
		dockerclient.FromEnv,
		dockerclient.WithAPIVersionNegotiation(),
	)
}

func runDockerCommand(cmd AgentCommand) AgentCommandResult {
	client, err := newDockerClient()
	if err != nil {
		return AgentCommandResult{ID: cmd.ID, OK: false, Error: "docker unavailable: " + err.Error()}
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	id := cmd.Target
	switch cmd.Action {
	case "start":
		err = client.ContainerStart(ctx, id, container.StartOptions{})
	case "stop":
		err = client.ContainerStop(ctx, id, container.StopOptions{})
	case "restart":
		err = client.ContainerRestart(ctx, id, container.StopOptions{})
	case "logs":
		lines := 100
		if cmd.Params != nil {
			if v, ok := cmd.Params["lines"].(float64); ok && v > 0 {
				lines = int(v)
			}
		}
		return dockerLogsResult(cmd, client, ctx, id, lines)
	default:
		return AgentCommandResult{ID: cmd.ID, OK: false, Error: "unsupported action: " + cmd.Action}
	}

	if err != nil {
		return AgentCommandResult{ID: cmd.ID, OK: false, Error: dockerErrorMessage(err)}
	}
	return AgentCommandResult{ID: cmd.ID, OK: true}
}

func dockerLogsResult(cmd AgentCommand, client *dockerclient.Client, ctx context.Context, id string, lines int) AgentCommandResult {
	reader, err := client.ContainerLogs(ctx, id, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Tail:       strconv.Itoa(lines),
	})
	if err != nil {
		return AgentCommandResult{ID: cmd.ID, OK: false, Error: dockerErrorMessage(err)}
	}
	defer reader.Close()

	raw, err := io.ReadAll(reader)
	if err != nil {
		return AgentCommandResult{ID: cmd.ID, OK: false, Error: err.Error()}
	}
	// Docker multiplexes stdout/stderr frames when the container has a
	// TTY=false; strip the 8-byte headers so the dashboard sees plain text.
	text := stripDockerLogFrames(raw)

	data, err := json.Marshal(LogsData{Logs: text})
	if err != nil {
		return AgentCommandResult{ID: cmd.ID, OK: false, Error: err.Error()}
	}
	return AgentCommandResult{ID: cmd.ID, OK: true, Data: data}
}

func dockerErrorMessage(err error) string {
	if cerrdefs.IsNotFound(err) {
		return "not found"
	}
	if dockerclient.IsErrConnectionFailed(err) {
		return "docker unavailable: cannot connect to daemon"
	}
	return err.Error()
}

// stripDockerLogFrames removes Docker's multiplexed stream headers
// (8-byte header: stream|0|0|0|uint32 big-endian length) when present.
func stripDockerLogFrames(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	// Heuristic: multiplexed frames start with stream 0/1/2 and a
	// big-endian length; plain text usually doesn't match for the whole buffer.
	var out bytes.Buffer
	i := 0
	muxed := true
	for i+8 <= len(raw) {
		stream := raw[i]
		if stream > 2 {
			muxed = false
			break
		}
		length := uint32(raw[i+4])<<24 | uint32(raw[i+5])<<16 | uint32(raw[i+6])<<8 | uint32(raw[i+7])
		if length == 0 || i+8+int(length) > len(raw) {
			muxed = false
			break
		}
		out.Write(raw[i+8 : i+8+int(length)])
		i += 8 + int(length)
	}
	if muxed && out.Len() > 0 {
		return out.String()
	}
	return string(raw)
}

// --- PM2 control ---

func runPM2Command(cmd AgentCommand) AgentCommandResult {
	id, err := strconv.Atoi(strings.TrimSpace(cmd.Target))
	if err != nil {
		return AgentCommandResult{ID: cmd.ID, OK: false, Error: "invalid PM2 id; expected a numeric value"}
	}

	switch cmd.Action {
	case "start", "stop", "restart":
		if err := pm2Control(id, cmd.Action); err != nil {
			return AgentCommandResult{ID: cmd.ID, OK: false, Error: err.Error()}
		}
		return AgentCommandResult{ID: cmd.ID, OK: true}
	case "logs":
		lines := 100
		if cmd.Params != nil {
			if v, ok := cmd.Params["lines"].(float64); ok && v > 0 {
				lines = int(v)
			}
		}
		logs, err := pm2TailLogs(id, lines)
		if err != nil {
			return AgentCommandResult{ID: cmd.ID, OK: false, Error: err.Error()}
		}
		data, err := json.Marshal(LogsData{Logs: logs})
		if err != nil {
			return AgentCommandResult{ID: cmd.ID, OK: false, Error: err.Error()}
		}
		return AgentCommandResult{ID: cmd.ID, OK: true, Data: data}
	default:
		return AgentCommandResult{ID: cmd.ID, OK: false, Error: "unsupported action: " + cmd.Action}
	}
}

// pm2Control issues the same axon-rpc methods the backend's PM2 client uses.
func pm2Control(id int, action string) error {
	sock := pm2SocketPath()
	if _, err := os.Stat(sock); err != nil {
		return fmt.Errorf("pm2 daemon unreachable: %s", sock)
	}
	client, err := newPM2RPCClient(sock)
	if err != nil {
		return fmt.Errorf("pm2 daemon unreachable: %w", err)
	}
	defer client.Close()

	var method string
	var args []any
	switch action {
	case "start":
		method, args = "startProcessId", []any{id}
	case "stop":
		method, args = "stopProcessId", []any{id}
	case "restart":
		method, args = "restartProcessId", []any{map[string]any{"id": id}}
	default:
		return fmt.Errorf("unsupported action: %s", action)
	}
	_, err = client.call(method, args)
	return err
}

// pm2TailLogs reads the last `lines` lines from the process's stdout
// and stderr log files, mirroring the backend's TailLogs.
func pm2TailLogs(id int, lines int) (string, error) {
	if lines <= 0 {
		lines = 100
	}
	sock := pm2SocketPath()
	if _, err := os.Stat(sock); err != nil {
		return "", fmt.Errorf("pm2 daemon unreachable: %s", sock)
	}
	client, err := newPM2RPCClient(sock)
	if err != nil {
		return "", fmt.Errorf("pm2 daemon unreachable: %w", err)
	}
	defer client.Close()

	raw, err := client.call("getMonitorData", []any{map[string]any{}})
	if err != nil {
		return "", err
	}
	if raw == nil {
		return "", nil
	}
	var entries []json.RawMessage
	if err := json.Unmarshal(raw, &entries); err != nil {
		return "", fmt.Errorf("decode getMonitorData: %w", err)
	}
	var outLog, errLog string
	found := false
	for _, entry := range entries {
		var meta struct {
			PMID int `json:"pm_id"`
		}
		if err := json.Unmarshal(entry, &meta); err != nil || meta.PMID != id {
			continue
		}
		var wrapped struct {
			PM2Env struct {
				OutLogPath string `json:"pm_out_log_path"`
				ErrLogPath string `json:"pm_err_log_path"`
			} `json:"pm2_env"`
		}
		if err := json.Unmarshal(entry, &wrapped); err != nil {
			return "", fmt.Errorf("decode pm2_env: %w", err)
		}
		outLog, errLog, found = wrapped.PM2Env.OutLogPath, wrapped.PM2Env.ErrLogPath, true
		break
	}
	if !found {
		return "", fmt.Errorf("no process found")
	}

	var b strings.Builder
	if errLog != "" {
		tail, err := tailFile(errLog, lines)
		if err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("tail err log: %w", err)
		}
		b.WriteString(tail)
	}
	if outLog != "" {
		tail, err := tailFile(outLog, lines)
		if err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("tail out log: %w", err)
		}
		if b.Len() > 0 && tail != "" {
			b.WriteByte('\n')
		}
		b.WriteString(tail)
	}

	out := b.String()
	if len(out) > 0 && out[0] == '\n' {
		out = out[1:]
	}
	return out, nil
}

// tailFile returns the last n lines of path (cheap reverse read).
func tailFile(path string, n int) (string, error) {
	if path == "" {
		return "", os.ErrNotExist
	}
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	const maxRead = 64 * 1024
	stat, err := f.Stat()
	if err != nil {
		return "", err
	}
	size := stat.Size()
	if size == 0 {
		return "", nil
	}
	readSize := int64(maxRead)
	if readSize > size {
		readSize = size
	}
	buf := make([]byte, readSize)
	if _, err := f.ReadAt(buf, size-readSize); err != nil {
		return "", err
	}
	if size > readSize && buf[0] != '\n' {
		if idx := bytes.IndexByte(buf, '\n'); idx >= 0 {
			buf = buf[idx+1:]
		}
	}

	lines := splitLinesReverse(buf, n)
	if len(lines) == 0 {
		return "", nil
	}
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}
	return strings.Join(lines, ""), nil
}

func splitLinesReverse(buf []byte, n int) []string {
	ring := make([]string, 0, n)
	start := 0
	for i := 0; i < len(buf); i++ {
		if buf[i] == '\n' {
			line := string(buf[start : i+1])
			if len(ring) < n {
				ring = append(ring, line)
			} else {
				ring = append(ring[1:], line)
			}
			start = i + 1
		}
	}
	if start < len(buf) {
		line := string(buf[start:]) + "\n"
		if len(ring) < n {
			ring = append(ring, line)
		} else {
			ring = append(ring[1:], line)
		}
	}
	return ring
}
