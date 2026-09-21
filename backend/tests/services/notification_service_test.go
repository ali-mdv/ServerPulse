package services_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"server-monitoring/internal/models"
	"server-monitoring/internal/services"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/v2/bson"
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

func TestNotificationHub_BroadcastsToConnectedClient(t *testing.T) {
	hub := services.NewNotificationHub()
	defer hub.Close()

	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		client := services.NewNotificationClient(conn, hub)
		hub.Register(client)
		client.Run()
	}))
	defer srv.Close()

	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	// Register happens in the handler goroutine; give it a beat to land
	// in the hub's client set before broadcasting.
	time.Sleep(50 * time.Millisecond)

	hub.Broadcast(models.Notification{
		ID:       bson.NewObjectID(),
		ServerID: "s1",
		Severity: models.NotificationSeverityInfo,
		Title:    "hello socket",
	})

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(msg), "hello socket") {
		t.Fatalf("payload %q missing notification title", msg)
	}
	if !strings.Contains(string(msg), `"type":"notification"`) {
		t.Fatalf("payload %q missing envelope type", msg)
	}
}
