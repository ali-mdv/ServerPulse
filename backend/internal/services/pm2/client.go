package pm2

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

// Client is a high-level wrapper around the PM2 daemon's RPC API.
// It hides the AMP/axon-rpc plumbing and exposes the operations used
// by the existing PM2Service interface.
type Client struct {
	rpc *rpcClient
}

// Dial opens a connection to a PM2 daemon listening on socketPath
// (typically $PM2_HOME/rpc.sock, default $HOME/.pm2/rpc.sock).
func Dial(socketPath string) (*Client, error) {
	if socketPath == "" {
		return nil, errors.New("pm2: empty socket path")
	}
	tr, err := newTransport(socketPath)
	if err != nil {
		return nil, err
	}
	return &Client{rpc: newRPCClient(tr)}, nil
}

// Close shuts down the underlying socket connection.
func (c *Client) Close() error {
	if c == nil || c.rpc == nil || c.rpc.tr == nil {
		return nil
	}
	return c.rpc.tr.Close()
}

// GetMonitorData mirrors PM2's `God.getMonitorData` — returns every
// process with its full pm2_env, monit stats, and log paths. Each item
// is left as raw JSON so callers can decode into their own model.
func (c *Client) GetMonitorData() ([]json.RawMessage, error) {
	raw, err := c.rpc.call("getMonitorData", []any{map[string]any{}})
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return []json.RawMessage{}, nil
	}
	var procs []json.RawMessage
	if err := json.Unmarshal(raw, &procs); err != nil {
		return nil, fmt.Errorf("pm2: decode getMonitorData: %w", err)
	}
	return procs, nil
}

// StartProcessId mirrors PM2's `God.startProcessId` (resume a stopped process).
func (c *Client) StartProcessID(id int) error {
	_, err := c.rpc.call("startProcessId", []any{id})
	return err
}

// StopProcessId mirrors PM2's `God.stopProcessId`.
func (c *Client) StopProcessID(id int) error {
	_, err := c.rpc.call("stopProcessId", []any{id})
	return err
}

// RestartProcessId mirrors PM2's `God.restartProcessId` (stop then start).
func (c *Client) RestartProcessID(id int) error {
	_, err := c.rpc.call("restartProcessId", []any{map[string]any{"id": id}})
	return err
}

// findProcess returns the monitor-data entry whose pm2_env.pm_id matches id.
func (c *Client) findProcess(id int) (json.RawMessage, error) {
	procs, err := c.GetMonitorData()
	if err != nil {
		return nil, err
	}
	for _, p := range procs {
		var meta struct {
			PMID int `json:"pm_id"`
		}
		if err := json.Unmarshal(p, &meta); err != nil {
			continue
		}
		if meta.PMID == id {
			return p, nil
		}
	}
	return nil, os.ErrNotExist
}

// FindProcessByID is a convenience for callers that just need the entry.
func (c *Client) FindProcessByID(id int) (json.RawMessage, error) {
	return c.findProcess(id)
}

// TailLogs reads the last `lines` lines from both the process's stdout
// and stderr log files (as configured by PM2) and returns them combined,
// stderr-first then stdout (PM2's CLI ordering).
//
// This mirrors `pm2 logs <id> --lines N --nostream` without needing the
// pm2 binary inside the container.
func (c *Client) TailLogs(id int, lines int) (string, error) {
	if lines <= 0 {
		lines = 100
	}
	proc, err := c.findProcess(id)
	if err != nil {
		return "", err
	}
	var wrapped struct {
		PM2Env struct {
			OutLogPath string `json:"pm_out_log_path"`
			ErrLogPath string `json:"pm_err_log_path"`
		} `json:"pm2_env"`
	}
	if err := json.Unmarshal(proc, &wrapped); err != nil {
		return "", fmt.Errorf("pm2: decode pm2_env: %w", err)
	}

	var b strings.Builder
	if wrapped.PM2Env.ErrLogPath != "" {
		tail, err := tailFile(wrapped.PM2Env.ErrLogPath, lines)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("pm2: tail err log: %w", err)
		}
		b.WriteString(tail)
	}
	if wrapped.PM2Env.OutLogPath != "" {
		tail, err := tailFile(wrapped.PM2Env.OutLogPath, lines)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("pm2: tail out log: %w", err)
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

// tailFile returns the last `n` lines of path, or "" if the file is
// missing. Implemented with a small reverse-seek so it stays cheap for
// large log files. Returns os.ErrNotExist if the path is empty or missing.
func tailFile(path string, n int) (string, error) {
	if path == "" {
		return "", os.ErrNotExist
	}
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", os.ErrNotExist
		}
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

	// Trim a half-line on the left when we sliced mid-line.
	if size > readSize && buf[0] != '\n' {
		if idx := bytes.IndexByte(buf, '\n'); idx >= 0 {
			buf = buf[idx+1:]
		}
	}

	lines := splitLinesReverse(buf, n)
	if len(lines) == 0 {
		return "", nil
	}
	// Reverse to chronological order.
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}
	return strings.Join(lines, ""), nil
}

// splitLinesReverse returns the last n lines from buf, preserving
// trailing newlines on each line. buf is treated as a text stream.
func splitLinesReverse(buf []byte, n int) []string {
	scanner := bufio.NewScanner(bytes.NewReader(buf))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	ring := make([]string, 0, n)
	for scanner.Scan() {
		line := scanner.Text() + "\n"
		if len(ring) < n {
			ring = append(ring, line)
		} else {
			ring = append(ring[1:], line)
		}
	}
	return ring
}