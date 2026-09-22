package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"server-monitoring/internal/models"
)

// ServerStaleToleranceFactor multiplies the history poll interval to give
// a server a grace period before it is declared down. Without it, a
// single missed push (scheduling jitter, a slow tick, a briefly blocked
// agent) would flip a server down then online again and spam
// notifications. A factor of 2 means a server must miss two consecutive
// poll windows before it is considered down.
const ServerStaleToleranceFactor = 2.0

// StaleAfter returns the age at which a server's last heartbeat counts as
// stale: the poll interval with the noise tolerance applied.
func StaleAfter(interval time.Duration) time.Duration {
	if interval <= 0 {
		return interval
	}
	return time.Duration(float64(interval) * ServerStaleToleranceFactor)
}

// emitServerTransition raises a server_down / server_up notification for
// an availability change. Callers only pass ServerStatusDown or
// ServerStatusOnline.
func emitServerTransition(ctx context.Context, notifications NotificationService, server models.Server, status models.ServerStatus) {
	if notifications == nil {
		return
	}

	n := models.Notification{
		ServerID: server.HexID(),
		Meta: map[string]any{
			"serverName": server.Name,
			"status":     string(status),
		},
	}
	if status == models.ServerStatusDown {
		n.Type = models.NotificationTypeServerDown
		n.Severity = models.NotificationSeverityCritical
		n.Title = fmt.Sprintf("%s is down", server.Name)
		n.Message = fmt.Sprintf("%s stopped reporting metrics.", server.Name)
	} else {
		n.Type = models.NotificationTypeServerUp
		n.Severity = models.NotificationSeverityInfo
		n.Title = fmt.Sprintf("%s is back online", server.Name)
		n.Message = fmt.Sprintf("%s is reporting metrics again.", server.Name)
	}

	if _, err := notifications.Notify(ctx, n); err != nil {
		log.Printf("notifications: server transition failed: %v", err)
	}
}

// emitProviderTransition raises a notification when a provider's
// availability flips. The first observation (prev == nil) is treated as
// a baseline and stays silent.
func emitProviderTransition(ctx context.Context, notifications NotificationService, serverID, provider string, prev *models.ProviderState, available bool) {
	if notifications == nil || prev == nil || prev.Available == available {
		return
	}

	n := models.Notification{
		ServerID: serverID,
		Meta:     map[string]any{"provider": provider},
	}
	if available {
		n.Type = models.NotificationTypeProviderAvailable
		n.Severity = models.NotificationSeverityInfo
		n.Title = fmt.Sprintf("%s is back online", providerLabel(provider))
		n.Message = fmt.Sprintf("The %s daemon on this server is reachable again.", provider)
	} else {
		n.Type = models.NotificationTypeProviderUnavailable
		n.Severity = models.NotificationSeverityCritical
		n.Title = fmt.Sprintf("%s is unavailable", providerLabel(provider))
		n.Message = fmt.Sprintf("The %s daemon on this server could not be reached.", provider)
	}

	if _, err := notifications.Notify(ctx, n); err != nil {
		log.Printf("notifications: provider transition failed: %v", err)
	}
}

// emitServiceTransitions raises a notification for every managed service
// whose running state changed since the previous observation. It only
// runs while the provider itself is available, so a dead daemon produces
// one provider alert instead of one per service.
func emitServiceTransitions(ctx context.Context, notifications NotificationService, serverID, provider string, prev *models.ProviderState, current []models.ServiceSnapshot) {
	if notifications == nil || prev == nil || !prev.Available {
		return
	}

	before := make(map[string]models.ServiceSnapshot, len(prev.Services))
	for _, svc := range prev.Services {
		before[svc.Meta.ServiceID] = svc
	}

	for _, svc := range current {
		old, ok := before[svc.Meta.ServiceID]
		if !ok || old.Available == svc.Available {
			continue
		}

		n := models.Notification{
			ServerID: serverID,
			Meta: map[string]any{
				"provider":  provider,
				"serviceId": svc.Meta.ServiceID,
				"name":      svc.Meta.Name,
			},
		}
		if svc.Available {
			n.Type = models.NotificationTypeServiceUp
			n.Severity = models.NotificationSeverityInfo
			n.Title = fmt.Sprintf("%s is back up", svc.Meta.Name)
			n.Message = fmt.Sprintf("%s on %s is running again.", svc.Meta.Name, provider)
		} else {
			n.Type = models.NotificationTypeServiceDown
			n.Severity = models.NotificationSeverityCritical
			n.Title = fmt.Sprintf("%s is down", svc.Meta.Name)
			n.Message = fmt.Sprintf("%s (%s) on %s is not running.", svc.Meta.Name, svc.Status, provider)
		}

		if _, err := notifications.Notify(ctx, n); err != nil {
			log.Printf("notifications: service transition failed: %v", err)
		}
	}
}

func providerLabel(provider string) string {
	switch provider {
	case models.ProviderPM2:
		return "PM2"
	case models.ProviderDocker:
		return "Docker"
	default:
		return provider
	}
}
