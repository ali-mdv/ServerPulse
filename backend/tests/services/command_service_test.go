package services_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"server-monitoring/internal/dtos"
	"server-monitoring/internal/models"
	"server-monitoring/internal/services"
	apperrors "server-monitoring/pkg/errors"

	"github.com/docker/docker/api/types/container"
)

func TestCommandService_LocalDockerStart(t *testing.T) {
	docker := &cmdDockerStub{}
	cmd := services.NewCommandService(nil, docker, nil, nil)

	// Empty serverId short-circuits to local.
	if err := cmd.ExecuteAction(context.Background(), "", "docker", "start", "abc"); err != nil {
		t.Fatalf("ExecuteAction: %v", err)
	}
	if docker.startedID != "abc" {
		t.Fatalf("startedID = %q, want abc", docker.startedID)
	}
}

func TestCommandService_LocalPM2InvalidID(t *testing.T) {
	pm2 := &cmdPM2Stub{}
	cmd := services.NewCommandService(nil, nil, pm2, nil)

	err := cmd.ExecuteAction(context.Background(), "local", "pm2", "start", "not-a-number")
	if err == nil {
		t.Fatal("expected error for non-numeric pm2 id")
	}
	if app, ok := err.(*apperrors.AppError); !ok || app.Code != 400 {
		t.Fatalf("err = %v, want AppError 400", err)
	}
}

func TestCommandService_UnknownProvider(t *testing.T) {
	cmd := services.NewCommandService(nil, nil, nil, nil)
	err := cmd.ExecuteAction(context.Background(), "", "systemd", "start", "x")
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestCommandService_RemoteWithoutAgent503(t *testing.T) {
	hub := services.NewAgentHub()
	cmd := services.NewCommandService(hub, nil, nil, nil)

	err := cmd.ExecuteAction(context.Background(), "remote-server-id", "docker", "start", "abc")
	if err == nil {
		t.Fatal("expected error when agent not connected")
	}
	app, ok := err.(*apperrors.AppError)
	if !ok || app.Code != 503 {
		t.Fatalf("err = %v, want AppError 503", err)
	}
	if !strings.Contains(app.Message, "not connected") {
		t.Fatalf("message = %q", app.Message)
	}
}

func TestCommandService_LocalLogs(t *testing.T) {
	docker := &cmdDockerStub{logs: "line1\nline2\n"}
	cmd := services.NewCommandService(nil, docker, nil, nil)

	got, err := cmd.ExecuteLogs(context.Background(), "", "docker", "abc", 100)
	if err != nil {
		t.Fatalf("ExecuteLogs: %v", err)
	}
	if got != docker.logs {
		t.Fatalf("logs = %q, want %q", got, docker.logs)
	}
}

func TestCommandService_LocalPM2Start(t *testing.T) {
	pm2 := &cmdPM2Stub{}
	cmd := services.NewCommandService(nil, nil, pm2, nil)

	if err := cmd.ExecuteAction(context.Background(), services.LocalServerID, "pm2", "restart", "3"); err != nil {
		t.Fatalf("ExecuteAction: %v", err)
	}
	if pm2.restartedID != 3 {
		t.Fatalf("restartedID = %d, want 3", pm2.restartedID)
	}
}

func TestAgentHub_NotConnected(t *testing.T) {
	hub := services.NewAgentHub()
	if hub.IsConnected("nope") {
		t.Fatal("IsConnected should be false for unknown id")
	}
	_, err := hub.Execute(context.Background(), "nope", dtos.AgentCommand{}, time.Second)
	if err == nil {
		t.Fatal("expected error")
	}
	if app, ok := err.(*apperrors.AppError); !ok || app.Code != 503 {
		t.Fatalf("err = %v, want 503", err)
	}
}

func TestAgentHub_ConnectedIDsEmpty(t *testing.T) {
	hub := services.NewAgentHub()
	if ids := hub.ConnectedServerIDs(); len(ids) != 0 {
		t.Fatalf("ids = %v, want empty", ids)
	}
}

func TestAgentCommandResultJSONRoundTrip(t *testing.T) {
	payload := map[string]any{
		"id":   "cmd-1",
		"ok":   true,
		"data": json.RawMessage(`{"logs":"hello"}`),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	var out dtos.AgentCommandResult
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	if out.ID != "cmd-1" || !out.OK {
		t.Fatalf("out = %+v", out)
	}
	var logs dtos.LogsData
	if err := json.Unmarshal(out.Data, &logs); err != nil {
		t.Fatal(err)
	}
	if logs.Logs != "hello" {
		t.Fatalf("logs = %q", logs.Logs)
	}
}

// --- stubs implementing the real service interfaces ---

type cmdDockerStub struct {
	startedID string
	logs      string
}

func (d *cmdDockerStub) ImagesList(bool) ([]models.DockerImage, error) { return nil, nil }
func (d *cmdDockerStub) ContainersList(bool) ([]models.DockerContainer, error) {
	return nil, nil
}
func (d *cmdDockerStub) InspectContainer(string) (container.InspectResponse, error) {
	return container.InspectResponse{}, nil
}
func (d *cmdDockerStub) StartContainer(id string) error {
	d.startedID = id
	return nil
}
func (d *cmdDockerStub) StopContainer(string) error    { return nil }
func (d *cmdDockerStub) RestartContainer(string) error { return nil }
func (d *cmdDockerStub) FetchContainerLogs(string, int) (string, error) {
	return d.logs, nil
}

type cmdPM2Stub struct {
	restartedID int
	startedID   int
}

func (p *cmdPM2Stub) List() ([]models.PM2Process, error) { return nil, nil }
func (p *cmdPM2Stub) FindPM2ProcessByID(int) (*models.PM2Process, error) {
	return nil, nil
}
func (p *cmdPM2Stub) StartPM2ProcessByID(id int) error {
	p.startedID = id
	return nil
}
func (p *cmdPM2Stub) StopPM2ProcessByID(int) error { return nil }
func (p *cmdPM2Stub) RestartPM2ProcessByID(id int) error {
	p.restartedID = id
	return nil
}
func (p *cmdPM2Stub) FetchContainerLogs(int, int) (string, error) { return "", nil }

var (
	_ services.DockerService = (*cmdDockerStub)(nil)
	_ services.PM2Service    = (*cmdPM2Stub)(nil)
)
