package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"server-monitoring/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const historyCollection = "service_history"

type HistoryRepository interface {
	EnsureCollection(ctx context.Context, retention time.Duration) error
	InsertBatch(ctx context.Context, snapshots []models.ServiceSnapshot) error
	ListTrackedServices(ctx context.Context, serverID, provider string) ([]models.SnapshotMeta, error)
	FindSeries(ctx context.Context, serverID, provider, serviceID string, from, to time.Time, bucket time.Duration) ([]models.ServiceSnapshot, error)
	FindAvailability(ctx context.Context, serverID, provider string, from, to time.Time) ([]models.ServiceSnapshot, error)
}

type historyRepository struct {
	Collection *mongo.Collection
}

func NewHistoryRepository(db *mongo.Database) HistoryRepository {
	return &historyRepository{
		Collection: db.Collection(historyCollection),
	}
}

// EnsureCollection creates the time-series collection if it doesn't exist
// and applies the TTL on the time field. Safe to call on every startup:
// "already exists" and a no-op collMod are both swallowed.
func (r *historyRepository) EnsureCollection(ctx context.Context, retention time.Duration) error {
	db := r.Collection.Database()

	ts := options.TimeSeries().
		SetTimeField("ts").
		SetMetaField("meta").
		SetGranularity("minutes")

	createOpts := options.CreateCollection().
		SetTimeSeriesOptions(ts).
		SetExpireAfterSeconds(int64(retention.Seconds()))

	if err := db.CreateCollection(ctx, historyCollection, createOpts); err != nil {
		if !isNamespaceExists(err) {
			return fmt.Errorf("create %s: %w", historyCollection, err)
		}
	}

	collMod := bson.D{
		{Key: "collMod", Value: historyCollection},
		{Key: "expireAfterSeconds", Value: int64(retention.Seconds())},
	}
	if err := db.RunCommand(ctx, collMod).Err(); err != nil {
		if !isNamespaceNotFound(err) {
			return fmt.Errorf("collMod %s: %w", historyCollection, err)
		}
	}
	return nil
}

func isNamespaceExists(err error) bool {
	var cmdErr mongo.CommandError
	if errors.As(err, &cmdErr) {
		return cmdErr.Code == 48 || strings.Contains(cmdErr.Message, "already exists")
	}
	return strings.Contains(err.Error(), "already exists")
}

func isNamespaceNotFound(err error) bool {
	var cmdErr mongo.CommandError
	if errors.As(err, &cmdErr) {
		return cmdErr.Code == 26
	}
	return false
}

func (r *historyRepository) InsertBatch(ctx context.Context, snapshots []models.ServiceSnapshot) error {
	if len(snapshots) == 0 {
		return nil
	}
	docs := make([]interface{}, len(snapshots))
	for i := range snapshots {
		docs[i] = snapshots[i]
	}
	_, err := r.Collection.InsertMany(ctx, docs)
	return err
}

func (r *historyRepository) ListTrackedServices(ctx context.Context, serverID, provider string) ([]models.SnapshotMeta, error) {
	match := bson.D{
		{Key: "meta.serverId", Value: serverID},
		{Key: "meta.provider", Value: provider},
		{Key: "meta.serviceId", Value: bson.D{{Key: "$ne", Value: models.ProviderStatusSentinelID}}},
	}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: match}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "serverId", Value: "$meta.serverId"},
				{Key: "provider", Value: "$meta.provider"},
				{Key: "serviceId", Value: "$meta.serviceId"},
				{Key: "name", Value: "$meta.name"},
			}},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "serverId", Value: "$_id.serverId"},
			{Key: "provider", Value: "$_id.provider"},
			{Key: "serviceId", Value: "$_id.serviceId"},
			{Key: "name", Value: "$_id.name"},
		}}},
	}

	cursor, err := r.Collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var out []models.SnapshotMeta
	for cursor.Next(ctx) {
		var meta models.SnapshotMeta
		if err := cursor.Decode(&meta); err != nil {
			return nil, err
		}
		out = append(out, meta)
	}
	return out, cursor.Err()
}

func (r *historyRepository) FindSeries(ctx context.Context, serverID, provider, serviceID string, from, to time.Time, bucket time.Duration) ([]models.ServiceSnapshot, error) {
	match := bson.D{
		{Key: "meta.serverId", Value: serverID},
		{Key: "meta.provider", Value: provider},
		{Key: "meta.serviceId", Value: serviceID},
		{Key: "ts", Value: bson.D{
			{Key: "$gte", Value: from},
			{Key: "$lte", Value: to},
		}},
	}

	if bucket <= 0 {
		findOpts := options.Find().SetSort(bson.D{{Key: "ts", Value: 1}})
		cursor, err := r.Collection.Find(ctx, match, findOpts)
		if err != nil {
			return nil, err
		}
		defer cursor.Close(ctx)

		var raw []models.ServiceSnapshot
		if err := cursor.All(ctx, &raw); err != nil {
			return nil, err
		}
		return raw, nil
	}

	bucketMs := int64(bucket.Milliseconds())
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: match}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{{Key: "$toDate", Value: bson.D{
				{Key: "$subtract", Value: bson.A{
					bson.D{{Key: "$toLong", Value: "$ts"}},
					bson.D{{Key: "$mod", Value: bson.A{
						bson.D{{Key: "$toLong", Value: "$ts"}},
						bucketMs,
					}}},
				}},
			}}}},
			{Key: "cpu", Value: bson.D{{Key: "$avg", Value: "$cpu"}}},
			{Key: "memory", Value: bson.D{{Key: "$avg", Value: "$memory"}}},
			{Key: "memUsage", Value: bson.D{{Key: "$avg", Value: "$memUsage"}}},
			{Key: "disk", Value: bson.D{{Key: "$avg", Value: "$disk"}}},
			{Key: "netSent", Value: bson.D{{Key: "$avg", Value: "$netSent"}}},
			{Key: "netRecv", Value: bson.D{{Key: "$avg", Value: "$netRecv"}}},
			{Key: "available", Value: bson.D{{Key: "$last", Value: "$available"}}},
			{Key: "status", Value: bson.D{{Key: "$last", Value: "$status"}}},
			{Key: "meta", Value: bson.D{{Key: "$last", Value: "$meta"}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
		{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "ts", Value: "$_id"},
			{Key: "meta", Value: 1},
			{Key: "status", Value: 1},
			{Key: "available", Value: 1},
			{Key: "cpu", Value: 1},
			{Key: "memory", Value: 1},
			{Key: "memUsage", Value: 1},
			{Key: "disk", Value: 1},
			{Key: "netSent", Value: 1},
			{Key: "netRecv", Value: 1},
		}}},
	}

	cursor, err := r.Collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregate series: %w", err)
	}
	defer cursor.Close(ctx)

	var out []models.ServiceSnapshot
	if err := cursor.All(ctx, &out); err != nil {
		return nil, fmt.Errorf("decode series: %w", err)
	}
	return out, nil
}

func (r *historyRepository) FindAvailability(ctx context.Context, serverID, provider string, from, to time.Time) ([]models.ServiceSnapshot, error) {
	return r.FindSeries(ctx, serverID, provider, models.ProviderStatusSentinelID, from, to, 0)
}