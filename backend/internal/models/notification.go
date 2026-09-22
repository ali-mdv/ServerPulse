package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// NotificationSeverity ranks how urgent a notification is. The
// dashboard sorts and colour-codes by this, and future delivery
// channels can use it to decide what deserves a push.
type NotificationSeverity string

const (
	NotificationSeverityInfo     NotificationSeverity = "info"
	NotificationSeverityWarning  NotificationSeverity = "warning"
	NotificationSeverityCritical NotificationSeverity = "critical"
)

// NotificationType identifies what happened, so the UI can render the
// right affordance and readers can filter without parsing Message.
type NotificationType string

const (
	NotificationTypeServerDown          NotificationType = "server_down"
	NotificationTypeServerUp            NotificationType = "server_up"
	NotificationTypeServiceDown         NotificationType = "service_down"
	NotificationTypeServiceUp           NotificationType = "service_up"
	NotificationTypeThresholdExceeded   NotificationType = "threshold_exceeded"
	NotificationTypeProviderUnavailable NotificationType = "provider_unavailable"
	NotificationTypeProviderAvailable   NotificationType = "provider_available"
)

// Notification is an alert raised by the backend when a watched
// condition changes (a server goes down, a service crashes, a usage
// threshold is crossed, a provider daemon becomes unreachable, …).
//
// Notifications are scoped to a Server, exactly like snapshots and
// provider state, so a multi-server dashboard can filter per host.
// `Meta` carries the provider/service/metric details specific to Type
// without forcing every consumer to know them.
type Notification struct {
	ID        bson.ObjectID        `bson:"_id,omitempty" json:"id"`
	ServerID  string               `bson:"serverId" json:"serverId"`
	Type      NotificationType     `bson:"type" json:"type"`
	Severity  NotificationSeverity `bson:"severity" json:"severity"`
	Title     string               `bson:"title" json:"title"`
	Message   string               `bson:"message,omitempty" json:"message,omitempty"`
	Meta      map[string]any       `bson:"meta,omitempty" json:"meta,omitempty"`
	Read      bool                 `bson:"read" json:"read"`
	CreatedAt time.Time            `bson:"createdAt" json:"createdAt"`
}
