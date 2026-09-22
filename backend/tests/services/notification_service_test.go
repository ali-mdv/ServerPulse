package services_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"server-monitoring/internal/models"
	"server-monitoring/internal/services"
	jwtutil "server-monitoring/pkg/jwt"
)

func TestNotificationService_NotifyPersistsWithDefaults(t *testing.T) {
	repo := &fakeNotificationRepo{}
	svc := services.NewNotificationServiceFromRepo(repo)

	got, err := svc.Notify(context.Background(), models.Notification{
		ServerID: "s1",
		Type:     models.NotificationTypeServiceDown,
		Severity: models.NotificationSeverityCritical,
		Title:    "api is down",
	})
	if err != nil {
		t.Fatalf("Notify: %v", err)
	}
	if got.ID.IsZero() {
		t.Fatal("expected generated ID")
	}
	if got.CreatedAt.IsZero() {
		t.Fatal("expected generated CreatedAt")
	}
	if got.Read {
		t.Fatal("new notification must start unread")
	}
	if len(repo.items) != 1 {
		t.Fatalf("persisted %d notifications, want 1", len(repo.items))
	}
}

func TestNotificationService_ListFiltersAndLimit(t *testing.T) {
	repo := &fakeNotificationRepo{}
	svc := services.NewNotificationServiceFromRepo(repo)

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if _, err := svc.Notify(ctx, models.Notification{
			ServerID: "s1",
			Severity: models.NotificationSeverityWarning,
			Title:    "warn",
		}); err != nil {
			t.Fatalf("Notify: %v", err)
		}
	}
	if _, err := svc.Notify(ctx, models.Notification{
		ServerID: "s2",
		Severity: models.NotificationSeverityCritical,
		Title:    "crit",
	}); err != nil {
		t.Fatalf("Notify: %v", err)
	}

	items, err := svc.List(ctx, services.NotificationFilter{ServerID: "s1"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("List server s1 = %d, want 3", len(items))
	}

	items, err = svc.List(ctx, services.NotificationFilter{Severity: models.NotificationSeverityCritical})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("List critical = %d, want 1", len(items))
	}

	items, err = svc.List(ctx, services.NotificationFilter{Limit: 2})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("List limit 2 = %d, want 2", len(items))
	}
}

func TestNotificationService_MarkReadAndCount(t *testing.T) {
	repo := &fakeNotificationRepo{}
	svc := services.NewNotificationServiceFromRepo(repo)
	ctx := context.Background()

	n, err := svc.Notify(ctx, models.Notification{ServerID: "s1", Title: "a"})
	if err != nil {
		t.Fatalf("Notify: %v", err)
	}
	if _, err := svc.Notify(ctx, models.Notification{ServerID: "s1", Title: "b"}); err != nil {
		t.Fatalf("Notify: %v", err)
	}

	if count, _ := svc.UnreadCount(ctx, "s1"); count != 2 {
		t.Fatalf("unread = %d, want 2", count)
	}

	if err := svc.MarkRead(ctx, n.ID); err != nil {
		t.Fatalf("MarkRead: %v", err)
	}
	if count, _ := svc.UnreadCount(ctx, "s1"); count != 1 {
		t.Fatalf("unread after MarkRead = %d, want 1", count)
	}

	updated, err := svc.MarkAllRead(ctx, "s1")
	if err != nil {
		t.Fatalf("MarkAllRead: %v", err)
	}
	if updated != 1 {
		t.Fatalf("MarkAllRead updated = %d, want 1", updated)
	}
	if count, _ := svc.UnreadCount(ctx, "s1"); count != 0 {
		t.Fatalf("unread after MarkAllRead = %d, want 0", count)
	}
}

// TestNotificationSocket_HandshakeAuthAndBroadcast drives the raw
// Engine.IO polling protocol end-to-end: handshake, authenticated
// Socket.IO connect, then a broadcast the client reads back. This is the
// Go-side equivalent of what socket.io-client does in the browser.
func TestNotificationSocket_HandshakeAuthAndBroadcast(t *testing.T) {
	repo := &fakeNotificationRepo{}
	svc := services.NewNotificationServiceFromRepo(repo)
	defer svc.Close()

	mux := http.NewServeMux()
	mux.Handle(services.NotificationSocketPath+"/", svc.Handler())
	srv := httptest.NewServer(mux)
	defer srv.Close()

	base := srv.URL + services.NotificationSocketPath + "/"
	client := srv.Client()

	// 1. Engine.IO handshake.
	resp, err := client.Get(base + "?EIO=4&transport=polling")
	if err != nil {
		t.Fatalf("handshake: %v", err)
	}
	body := readBody(t, resp)
	sid := parseSID(t, body)

	// 2. Socket.IO CONNECT carrying the JWT in the auth payload, exactly
	// like socket.io-client's `auth: { token }` option.
	token, err := jwtutil.GenerateToken("test-user")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	connectURL := base + "?EIO=4&transport=polling&sid=" + sid
	connectPayload := `40{"token":"` + token + `"}`
	resp, err = client.Post(connectURL, "text/plain", strings.NewReader(connectPayload))
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	readBody(t, resp)

	// 3. Broadcast and read the event back from the polling stream.
	time.Sleep(50 * time.Millisecond)
	if _, err := svc.Notify(context.Background(), models.Notification{
		ServerID: "s1",
		Severity: models.NotificationSeverityInfo,
		Title:    "hello socket",
	}); err != nil {
		t.Fatalf("Notify: %v", err)
	}

	resp, err = client.Get(connectURL)
	if err != nil {
		t.Fatalf("poll: %v", err)
	}
	msg := readBody(t, resp)
	if !strings.Contains(msg, "hello socket") {
		t.Fatalf("poll payload %q missing notification title", msg)
	}
	if !strings.Contains(msg, "notification") {
		t.Fatalf("poll payload %q missing event name", msg)
	}
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d: %s", resp.StatusCode, raw)
	}
	return string(raw)
}

func parseSID(t *testing.T, body string) string {
	t.Helper()
	// Engine.IO open packet: 0{"sid":"...","upgrades":[...],...}
	idx := strings.Index(body, "{")
	if idx < 0 {
		t.Fatalf("handshake payload %q missing JSON", body)
	}
	var payload struct {
		SID string `json:"sid"`
	}
	if err := json.Unmarshal([]byte(body[idx:]), &payload); err != nil {
		t.Fatalf("decode handshake %q: %v", body, err)
	}
	if payload.SID == "" {
		t.Fatalf("handshake %q missing sid", body)
	}
	return payload.SID
}
