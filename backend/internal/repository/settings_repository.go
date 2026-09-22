package repository

import (
	"context"
	"errors"
	"fmt"

	"server-monitoring/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const settingsCollection = "settings"

type SettingsRepository interface {
	Get(ctx context.Context) (*models.Settings, error)
	Ensure(ctx context.Context, defaults models.Settings) (*models.Settings, error)
	Upsert(ctx context.Context, settings models.Settings) (*models.Settings, error)
}

type settingsRepository struct {
	Collection *mongo.Collection
}

func NewSettingsRepository(db *mongo.Database) SettingsRepository {
	return &settingsRepository{Collection: db.Collection(settingsCollection)}
}

func (r *settingsRepository) Get(ctx context.Context) (*models.Settings, error) {
	var settings models.Settings
	err := r.Collection.FindOne(ctx, singletonFilter()).Decode(&settings)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("get settings: %w", err)
	}
	return &settings, nil
}

// Ensure inserts the defaults only when the singleton is missing. The
// $setOnInsert with a fixed _id makes this atomic, so concurrent
// startups cannot create a second document.
func (r *settingsRepository) Ensure(ctx context.Context, defaults models.Settings) (*models.Settings, error) {
	defaults.ID = models.SettingsSingletonID
	update := bson.D{{Key: "$setOnInsert", Value: settingsDoc(defaults)}}
	opts := options.UpdateOne().SetUpsert(true)
	if _, err := r.Collection.UpdateOne(ctx, singletonFilter(), update, opts); err != nil {
		return nil, fmt.Errorf("ensure settings: %w", err)
	}
	return r.Get(ctx)
}

// Upsert overwrites the singleton with the supplied values.
func (r *settingsRepository) Upsert(ctx context.Context, settings models.Settings) (*models.Settings, error) {
	settings.ID = models.SettingsSingletonID
	update := bson.D{{Key: "$set", Value: settingsDoc(settings)}}
	opts := options.UpdateOne().SetUpsert(true)
	if _, err := r.Collection.UpdateOne(ctx, singletonFilter(), update, opts); err != nil {
		return nil, fmt.Errorf("upsert settings: %w", err)
	}
	return &settings, nil
}

func singletonFilter() bson.D {
	return bson.D{{Key: "_id", Value: models.SettingsSingletonID}}
}

func settingsDoc(settings models.Settings) bson.D {
	return bson.D{
		{Key: "appearance", Value: settings.Appearance},
		{Key: "historyPollInterval", Value: settings.HistoryPollInterval},
		{Key: "historyRetention", Value: settings.HistoryRetention},
		{Key: "updatedAt", Value: settings.UpdatedAt},
	}
}
