package services_test

import (
	"context"
	"testing"

	"server-monitoring/internal/dtos"
	"server-monitoring/internal/models"
	"server-monitoring/internal/services"
	setup_test "server-monitoring/tests/setup"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestAgentService_IngestPush_StampsServerIDOnSnapshots(t *testing.T) {
	repo := &fakeAgentRepo{
		servers: []models.Server{{
			ID:     bson.NewObjectID(),
			Name:   "agent-01",
			Status: models.ServerStatusUnknown,
		}},
	}
	state := &fakeAgentState{}
	history := &fakeAgentHistory{}
	svc := services.NewAgentService(repo, state, history, nil)

	server := models.Server{
		ID:   repo.servers[0].ID,
		Name: "agent-01",
	}
	push := dtos.AgentPushDTO{
		Name:        "agent-01",
		Description: "edge node",
		Usage:       models.SystemUsage{CpuUsage: 42.0},
		Providers: map[string]dtos.AgentProviderDTO{
			"pm2": {
				Available: true,
				Services: []models.ServiceSnapshot{
					{Meta: models.SnapshotMeta{ServiceID: "0", Name: "api"}, CPU: 1.0},
				},
			},
			"docker": {
				Available: true,
				Services: []models.ServiceSnapshot{
					{Meta: models.SnapshotMeta{ServiceID: "abc", Name: "web"}, CPU: 2.0},
				},
			},
		},
	}

	if err := svc.IngestPush(context.Background(), server, push); err != nil {
		t.Fatalf("IngestPush: %v", err)
	}

	if len(state.providerWrites) != 2 {
		t.Fatalf("provider writes = %d, want 2", len(state.providerWrites))
	}
	for _, w := range state.providerWrites {
		for _, snap := range w.Services {
			if snap.Meta.ServerID != server.HexID() {
				t.Fatalf("provider %s: snapshot not stamped, serverID = %q", w.Provider, snap.Meta.ServerID)
			}
			if snap.Meta.Provider != w.Provider {
				t.Fatalf("provider meta mismatch: got %q, want %q", snap.Meta.Provider, w.Provider)
			}
			if snap.Ts.IsZero() {
				t.Fatalf("snapshot ts not set for provider %s", w.Provider)
			}
		}
	}

	if len(state.usageWrites) != 1 {
		t.Fatalf("usage writes = %d, want 1", len(state.usageWrites))
	}
	if state.usageWrites[0].Usage.CpuUsage != 42.0 {
		t.Fatalf("usage cpu = %v", state.usageWrites[0].Usage.CpuUsage)
	}

	if len(history.batches) != 2 {
		t.Fatalf("history batches = %d, want 2", len(history.batches))
	}
}

func TestAgentService_IdentityChanged(t *testing.T) {
	server := models.Server{Name: "a", Description: "old", Host: "10.0.0.1", Port: 80}
	cases := []struct {
		name string
		push dtos.AgentPushDTO
		want bool
	}{
		{"no change", dtos.AgentPushDTO{Name: "a", Description: "old", Host: "10.0.0.1", Port: 80}, false},
		{"name change", dtos.AgentPushDTO{Name: "b", Description: "old", Host: "10.0.0.1", Port: 80}, true},
		{"description change", dtos.AgentPushDTO{Name: "a", Description: "new", Host: "10.0.0.1", Port: 80}, true},
		{"host change", dtos.AgentPushDTO{Name: "a", Description: "old", Host: "10.0.0.2", Port: 80}, true},
		{"port change", dtos.AgentPushDTO{Name: "a", Description: "old", Host: "10.0.0.1", Port: 81}, true},
		{"empty host treated as no-op", dtos.AgentPushDTO{Name: "a", Description: "old", Host: "", Port: 80}, false},
		{"zero port treated as no-op", dtos.AgentPushDTO{Name: "a", Description: "old", Host: "10.0.0.1", Port: 0}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := setup_test.AgentIdentityChanged(server, tc.push); got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAgentService_TouchIdentity_IgnoresEmptyHostPort(t *testing.T) {
	repo := &fakeAgentRepo{}
	id := bson.NewObjectID()
	push := dtos.AgentPushDTO{Name: "n", Host: "", Port: 0, Description: "d"}

	// The fake doesn't decode the bson.D, but the call shouldn't error.
	if err := setup_test.AgentTouchIdentity(context.Background(), repo, id, push); err != nil {
		t.Fatalf("touchIdentity: %v", err)
	}
	if repo.updateCalls != 1 {
		t.Fatalf("update calls = %d, want 1", repo.updateCalls)
	}
}

func TestAgentService_IngestPush_EmitsServiceDownNotification(t *testing.T) {
	serverID := bson.NewObjectID()
	repo := &fakeAgentRepo{servers: []models.Server{{
		ID: serverID, Name: "agent-01", Status: models.ServerStatusOnline,
	}}}
	state := &fakeAgentState{}
	history := &fakeAgentHistory{}
	notifRepo := &fakeNotificationRepo{}
	notifications := services.NewNotificationServiceFromRepo(notifRepo)
	defer notifications.Close()
	svc := services.NewAgentService(repo, state, history, notifications)

	server := models.Server{ID: serverID, Name: "agent-01", Status: models.ServerStatusOnline}
	online := dtos.AgentPushDTO{
		Name: "agent-01",
		Providers: map[string]dtos.AgentProviderDTO{
			"pm2": {Available: true, Services: []models.ServiceSnapshot{
				{Meta: models.SnapshotMeta{ServiceID: "0", Name: "api"}, Status: "online", Available: true},
			}},
		},
	}

	// First observation is the baseline: no notification.
	if err := svc.IngestPush(context.Background(), server, online); err != nil {
		t.Fatalf("first push: %v", err)
	}
	if len(notifRepo.items) != 0 {
		t.Fatalf("baseline push created %d notifications, want 0", len(notifRepo.items))
	}

	offline := dtos.AgentPushDTO{
		Name: "agent-01",
		Providers: map[string]dtos.AgentProviderDTO{
			"pm2": {Available: true, Services: []models.ServiceSnapshot{
				{Meta: models.SnapshotMeta{ServiceID: "0", Name: "api"}, Status: "stopped", Available: false},
			}},
		},
	}
	if err := svc.IngestPush(context.Background(), server, offline); err != nil {
		t.Fatalf("second push: %v", err)
	}

	if len(notifRepo.items) != 1 {
		t.Fatalf("notifications = %d, want 1", len(notifRepo.items))
	}
	if notifRepo.items[0].Type != models.NotificationTypeServiceDown {
		t.Fatalf("type = %q, want service_down", notifRepo.items[0].Type)
	}
	if notifRepo.items[0].ServerID != serverID.Hex() {
		t.Fatalf("serverID = %q", notifRepo.items[0].ServerID)
	}
}

func TestAgentService_IngestPush_EmitsServerUpOnRecovery(t *testing.T) {
	serverID := bson.NewObjectID()
	repo := &fakeAgentRepo{servers: []models.Server{{
		ID: serverID, Name: "agent-01", Status: models.ServerStatusDown,
	}}}
	state := &fakeAgentState{}
	history := &fakeAgentHistory{}
	notifRepo := &fakeNotificationRepo{}
	notifications := services.NewNotificationServiceFromRepo(notifRepo)
	defer notifications.Close()
	svc := services.NewAgentService(repo, state, history, notifications)

	server := models.Server{ID: serverID, Name: "agent-01", Status: models.ServerStatusDown}
	if err := svc.IngestPush(context.Background(), server, dtos.AgentPushDTO{Name: "agent-01"}); err != nil {
		t.Fatalf("IngestPush: %v", err)
	}

	if len(notifRepo.items) != 1 {
		t.Fatalf("notifications = %d, want 1", len(notifRepo.items))
	}
	if notifRepo.items[0].Type != models.NotificationTypeServerUp {
		t.Fatalf("type = %q, want server_up", notifRepo.items[0].Type)
	}
}

func TestAgentService_IngestPush_UpdatesHeartbeat(t *testing.T) {
	repo := &fakeAgentRepo{servers: []models.Server{{
		ID:     bson.NewObjectID(),
		Name:   "agent-01",
		Status: models.ServerStatusUnknown,
	}}}
	state := &fakeAgentState{}
	history := &fakeAgentHistory{}
	svc := services.NewAgentService(repo, state, history, nil)

	server := models.Server{ID: repo.servers[0].ID, Name: "agent-01"}
	if err := svc.IngestPush(context.Background(), server, dtos.AgentPushDTO{
		Name:  "agent-01",
		Usage: models.SystemUsage{},
	}); err != nil {
		t.Fatal(err)
	}
	if repo.servers[0].Status != models.ServerStatusOnline {
		t.Fatalf("status = %q, want online", repo.servers[0].Status)
	}
	if repo.servers[0].LastSeen == nil {
		t.Fatal("LastSeen should be set")
	}
}
