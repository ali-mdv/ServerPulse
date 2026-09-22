package services

import (
	"context"
	"errors"
	"net/http"
	"time"

	"server-monitoring/internal/models"
	"server-monitoring/internal/repository"
	apperrors "server-monitoring/pkg/errors"
	database "server-monitoring/pkg/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const defaultNotificationLimit = 100

// NotificationFilter narrows a notification listing. Zero values mean
// "no constraint", except Limit which falls back to a sane cap.
type NotificationFilter struct {
	ServerID   string
	Severity   models.NotificationSeverity
	UnreadOnly bool
	Limit      int
}

type NotificationService interface {
	Notify(ctx context.Context, n models.Notification) (*models.Notification, error)
	List(ctx context.Context, filter NotificationFilter) ([]models.Notification, error)
	MarkRead(ctx context.Context, id bson.ObjectID) error
	MarkAllRead(ctx context.Context, serverID string) (int64, error)
	UnreadCount(ctx context.Context, serverID string) (int64, error)
	EnsureSchema(ctx context.Context) error
	Handler() http.Handler
	Close()
}

type notificationService struct {
	repo   repository.NotificationRepository
	socket *notificationServer
}

func NewNotificationService(dbName string) NotificationService {
	db := database.GetDatabase(dbName)
	return &notificationService{
		repo:   repository.NewNotificationRepository(db),
		socket: newNotificationServer(),
	}
}

// NewNotificationServiceFromRepo is the repo-backed constructor used by
// the test/setup helpers. Not part of the stable API.
func NewNotificationServiceFromRepo(repo repository.NotificationRepository) NotificationService {
	return &notificationService{repo: repo, socket: newNotificationServer()}
}

// Notify persists the notification then pushes it to every connected
// socket. Callers may leave ID/CreatedAt/Read unset; they are filled in
// here so producers only describe the event.
func (s *notificationService) Notify(ctx context.Context, n models.Notification) (*models.Notification, error) {
	if n.ID.IsZero() {
		n.ID = bson.NewObjectID()
	}
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now().UTC()
	}
	n.Read = false

	err := s.repo.Insert(ctx, n)
	// Broadcast even if persistence failed: a live socket is exactly
	// when a monitoring alert matters most.
	s.socket.Broadcast(n)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (s *notificationService) List(ctx context.Context, filter NotificationFilter) ([]models.Notification, error) {
	if filter.Limit <= 0 {
		filter.Limit = defaultNotificationLimit
	}
	return s.repo.List(ctx, filter.ServerID, filter.Severity, filter.UnreadOnly, filter.Limit)
}

func (s *notificationService) MarkRead(ctx context.Context, id bson.ObjectID) error {
	if err := s.repo.MarkRead(ctx, id); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return apperrors.ErrNotFound
		}
		return err
	}
	return nil
}

func (s *notificationService) MarkAllRead(ctx context.Context, serverID string) (int64, error) {
	return s.repo.MarkAllRead(ctx, serverID)
}

func (s *notificationService) UnreadCount(ctx context.Context, serverID string) (int64, error) {
	return s.repo.CountUnread(ctx, serverID)
}

func (s *notificationService) EnsureSchema(ctx context.Context) error {
	return s.repo.EnsureIndexes(ctx)
}

// Handler exposes the Socket.IO HTTP handler so the router can mount it.
func (s *notificationService) Handler() http.Handler {
	return s.socket.Handler()
}

func (s *notificationService) Close() {
	s.socket.Close()
}
