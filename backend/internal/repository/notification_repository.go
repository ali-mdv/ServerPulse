package repository

import (
	"context"
	"fmt"

	"server-monitoring/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const notificationCollection = "notifications"

type NotificationRepository interface {
	EnsureIndexes(ctx context.Context) error
	Insert(ctx context.Context, n models.Notification) error
	List(ctx context.Context, serverID string, severity models.NotificationSeverity, unreadOnly bool, limit int) ([]models.Notification, error)
	MarkRead(ctx context.Context, id bson.ObjectID) error
	MarkAllRead(ctx context.Context, serverID string) (int64, error)
	CountUnread(ctx context.Context, serverID string) (int64, error)
}

type notificationRepository struct {
	Collection *mongo.Collection
}

func NewNotificationRepository(db *mongo.Database) NotificationRepository {
	return &notificationRepository{Collection: db.Collection(notificationCollection)}
}

// EnsureIndexes creates the lookup index the dashboard polls with:
// newest-first listing scoped to a server. Safe to call on every startup.
func (r *notificationRepository) EnsureIndexes(ctx context.Context) error {
	index := mongo.IndexModel{
		Keys: bson.D{
			{Key: "serverId", Value: 1},
			{Key: "createdAt", Value: -1},
		},
	}
	if _, err := r.Collection.Indexes().CreateOne(ctx, index); err != nil {
		return fmt.Errorf("ensure notification indexes: %w", err)
	}
	return nil
}

func (r *notificationRepository) Insert(ctx context.Context, n models.Notification) error {
	if n.ID.IsZero() {
		n.ID = bson.NewObjectID()
	}
	if _, err := r.Collection.InsertOne(ctx, n); err != nil {
		return fmt.Errorf("insert notification: %w", err)
	}
	return nil
}

func (r *notificationRepository) List(ctx context.Context, serverID string, severity models.NotificationSeverity, unreadOnly bool, limit int) ([]models.Notification, error) {
	filter := notificationFilter(serverID, severity, unreadOnly)
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := r.Collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer cursor.Close(ctx)

	var out []models.Notification
	if err := cursor.All(ctx, &out); err != nil {
		return nil, fmt.Errorf("decode notifications: %w", err)
	}
	return out, nil
}

func (r *notificationRepository) MarkRead(ctx context.Context, id bson.ObjectID) error {
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "read", Value: true}}}}
	res, err := r.Collection.UpdateOne(ctx, bson.D{{Key: "_id", Value: id}}, update)
	if err != nil {
		return fmt.Errorf("mark notification read: %w", err)
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *notificationRepository) MarkAllRead(ctx context.Context, serverID string) (int64, error) {
	filter := bson.D{}
	if serverID != "" {
		filter = append(filter, bson.E{Key: "serverId", Value: serverID})
	}
	filter = append(filter, bson.E{Key: "read", Value: false})

	update := bson.D{{Key: "$set", Value: bson.D{{Key: "read", Value: true}}}}
	res, err := r.Collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, fmt.Errorf("mark notifications read: %w", err)
	}
	return res.ModifiedCount, nil
}

func (r *notificationRepository) CountUnread(ctx context.Context, serverID string) (int64, error) {
	filter := bson.D{{Key: "read", Value: false}}
	if serverID != "" {
		filter = append(filter, bson.E{Key: "serverId", Value: serverID})
	}
	count, err := r.Collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count unread notifications: %w", err)
	}
	return count, nil
}

func notificationFilter(serverID string, severity models.NotificationSeverity, unreadOnly bool) bson.D {
	filter := bson.D{}
	if serverID != "" {
		filter = append(filter, bson.E{Key: "serverId", Value: serverID})
	}
	if severity != "" {
		filter = append(filter, bson.E{Key: "severity", Value: severity})
	}
	if unreadOnly {
		filter = append(filter, bson.E{Key: "read", Value: false})
	}
	return filter
}
