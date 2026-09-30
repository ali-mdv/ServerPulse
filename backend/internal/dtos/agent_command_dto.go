package dtos

import "encoding/json"

// AgentCommand is a control request the backend sends to a connected
// remote agent over Socket.IO (event `agent:command`). The agent
// executes it against the host's Docker / PM2 daemons and acks with
// an AgentCommandResult.
type AgentCommand struct {
	// ID correlates the command with its result (logging / debugging;
	// the Socket.IO ack already carries the protocol-level correlation).
	ID string `json:"id"`
	// Provider is "docker" or "pm2".
	Provider string `json:"provider"`
	// Action is "start", "stop", "restart", or "logs".
	Action string `json:"action"`
	// Target is the container ID / name (docker) or numeric pm_id as a string (pm2).
	Target string `json:"target"`
	// Params carries optional extras, e.g. {"lines":100} for logs.
	Params map[string]any `json:"params,omitempty"`
}

// AgentCommandResult is the agent's response to one AgentCommand.
// When OK is false, Error holds a human-readable reason suitable for
// surfacing to the dashboard.
type AgentCommandResult struct {
	ID    string          `json:"id"`
	OK    bool            `json:"ok"`
	Error string          `json:"error,omitempty"`
	// Data is provider-specific; for logs it is {"logs":"..."}.
	Data json.RawMessage `json:"data,omitempty"`
}

// LogsData is the Data payload for action=logs.
type LogsData struct {
	Logs string `json:"logs"`
}
