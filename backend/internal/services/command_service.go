package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"server-monitoring/internal/dtos"
	"server-monitoring/internal/models"
	apperrors "server-monitoring/pkg/errors"
)

// Provider / action vocabulary shared with the agent command protocol.
const (
	ProviderDocker = "docker"
	ProviderPM2    = "pm2"

	ActionStart   = "start"
	ActionStop    = "stop"
	ActionRestart = "restart"
	ActionLogs    = "logs"
)

// AgentLogsTimeout is more generous than the action timeout because
// log tails can be large.
const AgentLogsTimeout = 20 * time.Second

// CommandService executes start/stop/restart/logs against either the
// local host (in-process Docker/PM2 services) or a remote agent over
// the Socket.IO control channel, chosen by serverID.
type CommandService interface {
	// ExecuteAction runs start|stop|restart for provider on target.
	ExecuteAction(ctx context.Context, serverID, provider, action, target string) error
	// ExecuteLogs returns the last `lines` lines of logs for target.
	ExecuteLogs(ctx context.Context, serverID, provider, target string, lines int) (string, error)
	// IsLocal reports whether serverID refers to the backend's own host.
	IsLocal(ctx context.Context, serverID string) bool
}

type commandService struct {
	hub     *AgentHub
	docker  DockerService
	pm2     PM2Service
	servers ServerService
}

// NewCommandService wires the local executors and the remote hub.
// servers is used to resolve the local server's real hex ID; pass nil
// only in tests that always target remote IDs.
func NewCommandService(
	hub *AgentHub,
	docker DockerService,
	pm2 PM2Service,
	servers ServerService,
) CommandService {
	return &commandService{
		hub:     hub,
		docker:  docker,
		pm2:     pm2,
		servers: servers,
	}
}

func (s *commandService) IsLocal(ctx context.Context, serverID string) bool {
	if serverID == "" || serverID == models.LocalServerName || serverID == LocalServerID {
		return true
	}
	if s.servers == nil {
		return false
	}
	local, err := s.servers.EnsureLocalServer(ctx)
	if err != nil || local == nil {
		return false
	}
	return serverID == local.HexID()
}

func (s *commandService) ExecuteAction(ctx context.Context, serverID, provider, action, target string) error {
	if err := validateCommand(provider, action, target); err != nil {
		return err
	}

	if s.IsLocal(ctx, serverID) {
		return s.executeLocal(provider, action, target, 0)
	}

	cmd := dtos.AgentCommand{
		Provider: provider,
		Action:   action,
		Target:   target,
	}
	result, err := s.hub.Execute(ctx, serverID, cmd, DefaultAgentCommandTimeout)
	if err != nil {
		return err
	}
	return resultToError(result)
}

func (s *commandService) ExecuteLogs(ctx context.Context, serverID, provider, target string, lines int) (string, error) {
	if err := validateCommand(provider, ActionLogs, target); err != nil {
		return "", err
	}
	if lines <= 0 {
		lines = 100
	}

	if s.IsLocal(ctx, serverID) {
		return s.executeLocalLogs(provider, target, lines)
	}

	cmd := dtos.AgentCommand{
		Provider: provider,
		Action:   ActionLogs,
		Target:   target,
		Params:   map[string]any{"lines": lines},
	}
	result, err := s.hub.Execute(ctx, serverID, cmd, AgentLogsTimeout)
	if err != nil {
		return "", err
	}
	if err := resultToError(result); err != nil {
		return "", err
	}
	return logsFromResult(result)
}

func (s *commandService) executeLocal(provider, action, target string, _ int) error {
	switch provider {
	case ProviderDocker:
		if s.docker == nil {
			return apperrors.New(503, "docker service unavailable")
		}
		switch action {
		case ActionStart:
			return s.docker.StartContainer(target)
		case ActionStop:
			return s.docker.StopContainer(target)
		case ActionRestart:
			return s.docker.RestartContainer(target)
		}
		return apperrors.New(400, "unsupported action: "+action)
	case ProviderPM2:
		if s.pm2 == nil {
			return apperrors.New(503, "pm2 service unavailable")
		}
		id, err := parsePM2ID(target)
		if err != nil {
			return err
		}
		switch action {
		case ActionStart:
			return s.pm2.StartPM2ProcessByID(id)
		case ActionStop:
			return s.pm2.StopPM2ProcessByID(id)
		case ActionRestart:
			return s.pm2.RestartPM2ProcessByID(id)
		}
		return apperrors.New(400, "unsupported action: "+action)
	}
	return apperrors.New(400, "unknown provider: "+provider)
}

func (s *commandService) executeLocalLogs(provider, target string, lines int) (string, error) {
	switch provider {
	case ProviderDocker:
		if s.docker == nil {
			return "", apperrors.New(503, "docker service unavailable")
		}
		return s.docker.FetchContainerLogs(target, lines)
	case ProviderPM2:
		if s.pm2 == nil {
			return "", apperrors.New(503, "pm2 service unavailable")
		}
		id, err := parsePM2ID(target)
		if err != nil {
			return "", err
		}
		return s.pm2.FetchContainerLogs(id, lines)
	}
	return "", apperrors.New(400, "unknown provider: "+provider)
}

func validateCommand(provider, action, target string) error {
	if target == "" {
		return apperrors.New(400, "missing target")
	}
	switch provider {
	case ProviderDocker, ProviderPM2:
	default:
		return apperrors.New(400, "unknown provider: "+provider)
	}
	switch action {
	case ActionStart, ActionStop, ActionRestart, ActionLogs:
	default:
		return apperrors.New(400, "unsupported action: "+action)
	}
	return nil
}

func parsePM2ID(target string) (int, error) {
	id, err := strconv.Atoi(strings.TrimSpace(target))
	if err != nil {
		return 0, apperrors.New(400, "invalid PM2 id; expected a numeric value")
	}
	return id, nil
}

// resultToError maps an agent-side failure onto an AppError. A failure
// with no useful code becomes a 500; "not found" style messages become
// 404 so the frontend's existing NotFound branches keep working.
func resultToError(result dtos.AgentCommandResult) error {
	if result.OK {
		return nil
	}
	msg := result.Error
	if msg == "" {
		msg = "agent command failed"
	}
	lower := strings.ToLower(msg)
	if strings.Contains(lower, "not found") ||
		strings.Contains(lower, "no such") ||
		strings.Contains(lower, "no process") {
		return apperrors.New(404, msg)
	}
	if strings.Contains(lower, "unreachable") || strings.Contains(lower, "unavailable") {
		return apperrors.New(503, msg)
	}
	if strings.Contains(lower, "invalid") {
		return apperrors.New(400, msg)
	}
	return apperrors.New(500, msg)
}

func logsFromResult(result dtos.AgentCommandResult) (string, error) {
	if len(result.Data) == 0 {
		return "", nil
	}
	var data dtos.LogsData
	if err := json.Unmarshal(result.Data, &data); err != nil {
		return "", fmt.Errorf("decode logs payload: %w", err)
	}
	return data.Logs, nil
}
