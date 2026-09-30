// ServerPulse remote agent (Go).
//
// Collects system usage and optional PM2 provider state, then pushes them to
// a ServerPulse backend using an API key.
//
// Workflow:
//  1. Create a server/agent in the backend (POST /api/v1/servers) with a
//     name and optional description.
//  2. Generate an API key for that server (POST
//     /api/v1/servers/<id>/api-key).
//  3. Run this agent with SERVER_PULSE_API_KEY set to the revealed key.
//
// Required env:
//
//	SERVER_PULSE_URL      Base URL of the backend, e.g. http://localhost:8080
//	SERVER_PULSE_API_KEY  API key revealed by POST /servers/<id>/api-key
//
// Optional env:
//
//	SERVER_PULSE_INTERVAL Push interval in seconds (default: 30)
//	SERVER_PULSE_NAME     Agent/server name (default: hostname)
//	SERVER_PULSE_HOST     Host address to report (default: resolved hostname)
//	SERVER_PULSE_PORT     Port to report (default: 0)
//	SERVER_PULSE_DESCRIPTION  Optional description pushed with identity
//
// Dependencies:
//
//	go mod tidy
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	gopsutilnet "github.com/shirou/gopsutil/v4/net"
)

// Usage matches backend/internal/models.SystemUsage.
type Usage struct {
	UsedPercent float64 `json:"percent"`
	Used        uint64  `json:"used"`
	Total       uint64  `json:"total"`
}

type NetworkUsage struct {
	Send     uint64 `json:"sent"`
	Received uint64 `json:"received"`
}

type SystemUsage struct {
	MemUsage  Usage        `json:"memUsage"`
	CpuUsage  float64      `json:"cpuUsage"`
	DiskUsage Usage        `json:"diskUsage"`
	NetIO     NetworkUsage `json:"netIO"`
}

type SnapshotMeta struct {
	ServerID  string `json:"serverId"`
	Provider  string `json:"provider"`
	ServiceID string `json:"serviceId"`
	Name      string `json:"name"`
}

type ServiceSnapshot struct {
	Ts        time.Time    `json:"ts"`
	Meta      SnapshotMeta `json:"meta"`
	Status    string       `json:"status"`
	Available bool         `json:"available"`
	CPU       float64      `json:"cpu"`
	Memory    float64      `json:"memory"`
	MemUsage  float64      `json:"memUsage,omitempty"`
}

type AgentProvider struct {
	Available bool              `json:"available"`
	Services  []ServiceSnapshot `json:"services,omitempty"`
}

type PushPayload struct {
	Name        string                   `json:"name"`
	Host        string                   `json:"host,omitempty"`
	Port        int                      `json:"port,omitempty"`
	Description string                   `json:"description,omitempty"`
	Usage       SystemUsage              `json:"usage"`
	Providers   map[string]AgentProvider `json:"providers,omitempty"`
}

type netSampler struct {
	lastSent     uint64
	lastReceived uint64
	lastTime     time.Time
}

func (s *netSampler) sample() NetworkUsage {
	ioCounters, err := gopsutilnet.IOCounters(false)
	if err != nil || len(ioCounters) == 0 {
		return NetworkUsage{}
	}

	now := time.Now()
	c := ioCounters[0]
	usage := NetworkUsage{}

	if !s.lastTime.IsZero() {
		interval := now.Sub(s.lastTime).Seconds()
		if interval > 0 {
			usage.Send = uint64(float64(c.BytesSent-s.lastSent) / interval)
			usage.Received = uint64(float64(c.BytesRecv-s.lastReceived) / interval)
		}
	}

	s.lastSent = c.BytesSent
	s.lastReceived = c.BytesRecv
	s.lastTime = now
	return usage
}

func collectSystem(ns *netSampler) SystemUsage {
	vmStat, _ := mem.VirtualMemory()
	diskStat, _ := disk.Usage("/")
	cpuPercent, _ := cpu.Percent(time.Second, false)

	cpuVal := 0.0
	if len(cpuPercent) > 0 {
		cpuVal = cpuPercent[0]
	}

	return SystemUsage{
		MemUsage: Usage{
			UsedPercent: vmStat.UsedPercent,
			Used:        vmStat.Used,
			Total:       vmStat.Total,
		},
		CpuUsage: cpuVal,
		DiskUsage: Usage{
			UsedPercent: diskStat.UsedPercent,
			Used:        diskStat.Used,
			Total:       diskStat.Total,
		},
		NetIO: ns.sample(),
	}
}

// --- PM2 provider (axon-rpc over the PM2 unix socket) ---

const (
	ampVersion = 1
	ampMaxArgs = 15
)

func encodeAMP(args [][]byte) ([]byte, error) {
	n := len(args)
	if n == 0 {
		return nil, errors.New("amp: refuse to encode zero-arg message")
	}
	if n > ampMaxArgs {
		return nil, fmt.Errorf("amp: too many args (%d > %d)", n, ampMaxArgs)
	}
	total := 1
	for _, a := range args {
		total += 4 + len(a)
	}
	buf := make([]byte, total)
	buf[0] = (ampVersion << 4) | byte(n)
	off := 1
	for _, a := range args {
		binary.BigEndian.PutUint32(buf[off:], uint32(len(a)))
		off += 4
		copy(buf[off:], a)
		off += len(a)
	}
	return buf, nil
}

func decodeAMP(buf []byte) ([][]byte, error) {
	if len(buf) == 0 {
		return nil, errors.New("amp: empty frame")
	}
	meta := buf[0]
	version := meta >> 4
	argc := int(meta & 0x0f)
	if version != ampVersion {
		return nil, fmt.Errorf("amp: unsupported version %d", version)
	}
	if argc > ampMaxArgs {
		return nil, fmt.Errorf("amp: invalid argc %d", argc)
	}
	args := make([][]byte, argc)
	off := 1
	for i := 0; i < argc; i++ {
		if off+4 > len(buf) {
			return nil, fmt.Errorf("amp: short read on arg %d length", i)
		}
		l := binary.BigEndian.Uint32(buf[off:])
		off += 4
		if off+int(l) > len(buf) {
			return nil, fmt.Errorf("amp: short read on arg %d body", i)
		}
		args[i] = make([]byte, l)
		copy(args[i], buf[off:off+int(l)])
		off += int(l)
	}
	return args, nil
}

func packJSON(payload []byte) []byte {
	out := make([]byte, 0, len(payload)+2)
	out = append(out, 'j', ':')
	out = append(out, payload...)
	return out
}

func packString(s string) []byte {
	out := make([]byte, 0, len(s)+2)
	out = append(out, 's', ':')
	out = append(out, s...)
	return out
}

func unpackJSON(buf []byte) ([]byte, error) {
	if len(buf) < 2 || buf[0] != 'j' || buf[1] != ':' {
		return nil, fmt.Errorf("amp-msg: not a json-typed argument (got %q)", string(buf[:min(2, len(buf))]))
	}
	return buf[2:], nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type pm2CallRequest struct {
	Type   string `json:"type"`
	Method string `json:"method"`
	Args   []any  `json:"args"`
}

type pm2Reply struct {
	Args  []json.RawMessage `json:"args"`
	Error string            `json:"error"`
}

type pm2RPCClient struct {
	conn   net.Conn
	idCtr  atomic.Uint32
	closed bool
}

func newPM2RPCClient(socketPath string) (*pm2RPCClient, error) {
	c, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("pm2: dial %s: %w", socketPath, err)
	}
	return &pm2RPCClient{conn: c}, nil
}

func (c *pm2RPCClient) nextID() string {
	n := c.idCtr.Add(1)
	return strconv.Itoa(os.Getpid()) + ":" + strconv.FormatUint(uint64(n), 10)
}

func (c *pm2RPCClient) call(method string, args []any) (json.RawMessage, error) {
	if args == nil {
		args = []any{}
	}
	req := pm2CallRequest{Type: "call", Method: method, Args: args}
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("pm2 rpc: marshal %s: %w", method, err)
	}
	frame, err := encodeAMP([][]byte{packJSON(payload), packString(c.nextID())})
	if err != nil {
		return nil, fmt.Errorf("pm2 rpc: encode %s: %w", method, err)
	}
	if _, err := c.conn.Write(frame); err != nil {
		return nil, fmt.Errorf("pm2 rpc: send %s: %w", method, err)
	}

	rawArgs, err := c.recvFrame()
	if err != nil {
		return nil, fmt.Errorf("pm2 rpc: recv %s: %w", method, err)
	}
	if len(rawArgs) != 2 {
		return nil, fmt.Errorf("pm2 rpc: %s: expected 2 reply args, got %d", method, len(rawArgs))
	}
	body, err := unpackJSON(rawArgs[0])
	if err != nil {
		return nil, fmt.Errorf("pm2 rpc: %s: %w", method, err)
	}
	var r pm2Reply
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("pm2 rpc: %s: bad reply json: %w", method, err)
	}
	if r.Error != "" {
		return nil, fmt.Errorf("pm2 rpc %s: %s", method, r.Error)
	}
	if len(r.Args) == 0 {
		return nil, nil
	}
	return r.Args[0], nil
}

func (c *pm2RPCClient) recvFrame() ([][]byte, error) {
	var meta [1]byte
	if _, err := io.ReadFull(c.conn, meta[:]); err != nil {
		return nil, fmt.Errorf("pm2: read meta: %w", err)
	}
	argc := int(meta[0] & 0x0f)
	if argc > ampMaxArgs {
		return nil, fmt.Errorf("pm2: bad argc %d in reply meta %#x", argc, meta[0])
	}
	args := make([][]byte, argc)
	var lenBuf [4]byte
	for i := 0; i < argc; i++ {
		if _, err := io.ReadFull(c.conn, lenBuf[:]); err != nil {
			return nil, fmt.Errorf("pm2: read arg %d length: %w", i, err)
		}
		l := binary.BigEndian.Uint32(lenBuf[:])
		args[i] = make([]byte, l)
		if l == 0 {
			continue
		}
		if _, err := io.ReadFull(c.conn, args[i]); err != nil {
			return nil, fmt.Errorf("pm2: read arg %d body: %w", i, err)
		}
	}
	return args, nil
}

func (c *pm2RPCClient) Close() error {
	if c.closed {
		return nil
	}
	c.closed = true
	return c.conn.Close()
}

type pm2Process struct {
	Name   string `json:"name"`
	PMID   int    `json:"pm_id"`
	PM2Env struct {
		Status string `json:"status"`
	} `json:"pm2_env"`
	Monit struct {
		CPU    float64 `json:"cpu"`
		Memory float64 `json:"memory"`
	} `json:"monit"`
}

func pm2SocketPath() string {
	if p := os.Getenv("PM2_SOCKET_PATH"); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	if home == "" {
		home = "."
	}
	return filepath.Join(home, ".pm2", "rpc.sock")
}

func collectPM2() *AgentProvider {
	sock := pm2SocketPath()
	if _, err := os.Stat(sock); err != nil {
		return nil
	}

	client, err := newPM2RPCClient(sock)
	if err != nil {
		return &AgentProvider{Available: false, Services: []ServiceSnapshot{}}
	}
	defer client.Close()

	raw, err := client.call("getMonitorData", []any{map[string]any{}})
	if err != nil {
		return &AgentProvider{Available: false, Services: []ServiceSnapshot{}}
	}
	if raw == nil {
		return &AgentProvider{Available: true, Services: []ServiceSnapshot{}}
	}

	var procs []pm2Process
	if err := json.Unmarshal(raw, &procs); err != nil {
		return &AgentProvider{Available: false, Services: []ServiceSnapshot{}}
	}

	now := time.Now().UTC()
	services := make([]ServiceSnapshot, 0, len(procs))
	for _, p := range procs {
		services = append(services, ServiceSnapshot{
			Ts: now,
			Meta: SnapshotMeta{
				Provider:  "pm2",
				ServiceID: strconv.Itoa(p.PMID),
				Name:      p.Name,
			},
			Status:    p.PM2Env.Status,
			Available: p.PM2Env.Status == "online",
			CPU:       p.Monit.CPU,
			Memory:    p.Monit.Memory,
		})
	}
	return &AgentProvider{Available: true, Services: services}
}

func buildPayload(name, host string, port int, description string, ns *netSampler) PushPayload {
	providers := map[string]AgentProvider{}
	if pm2 := collectPM2(); pm2 != nil {
		providers["pm2"] = *pm2
	}

	return PushPayload{
		Name:        name,
		Host:        host,
		Port:        port,
		Description: description,
		Usage:       collectSystem(ns),
		Providers:   providers,
	}
}

func push(baseURL, apiKey string, payload PushPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	url := strings.TrimRight(baseURL, "/") + "/api/v1/agents/push"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	log.Printf("push accepted: %d %s", resp.StatusCode, string(respBody))
	return nil
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("Error: %s is required", key)
	}
	return v
}

func getenv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func main() {
	baseURL := requireEnv("SERVER_PULSE_URL")
	apiKey := requireEnv("SERVER_PULSE_API_KEY")

	intervalSec, _ := strconv.Atoi(getenv("SERVER_PULSE_INTERVAL", "30"))
	if intervalSec <= 0 {
		intervalSec = 30
	}
	interval := time.Duration(intervalSec) * time.Second

	hostname, _ := os.Hostname()
	name := os.Getenv("SERVER_PULSE_NAME")
	if name == "" {
		name = hostname
	}
	host := os.Getenv("SERVER_PULSE_HOST")
	if host == "" {
		host = hostname
	}
	port, _ := strconv.Atoi(os.Getenv("SERVER_PULSE_PORT"))
	description := os.Getenv("SERVER_PULSE_DESCRIPTION")

	log.Printf("ServerPulse agent starting: name=%s url=%s interval=%s", name, baseURL, interval)

	// Control channel: receive start/stop/restart/logs from the backend
	// over Socket.IO. Runs forever (with reconnect); push loop continues
	// independently below.
	go startControlChannel(baseURL, apiKey)

	ns := &netSampler{}
	for {
		start := time.Now()
		payload := buildPayload(name, host, port, description, ns)
		if err := push(baseURL, apiKey, payload); err != nil {
			log.Printf("push error: %v", err)
		}

		elapsed := time.Since(start)
		sleepFor := interval - elapsed
		if sleepFor < time.Second {
			sleepFor = time.Second
		}
		time.Sleep(sleepFor)
	}
}
