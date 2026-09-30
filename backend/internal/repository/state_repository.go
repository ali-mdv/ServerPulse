package repository

import (
	"context"
	"fmt"

	"server-monitoring/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	providerStateCollection = "provider_state"
	serverUsageCollection   = "server_usage"
)

type StateRepository interface {
	UpsertProviderState(ctx context.Context, state models.ProviderState) error
	UpsertServerUsage(ctx context.Context, state models.ServerUsageState) error
	GetProviderState(ctx context.Context, serverID, provider string) (*models.ProviderState, error)
	GetServerUsage(ctx context.Context, serverID string) (*models.ServerUsageState, error)
}

type stateRepository struct {
	providerCollection *mongo.Collection
	serverCollection   *mongo.Collection
}

func NewStateRepository(db *mongo.Database) StateRepository {
	return &stateRepository{
		providerCollection: db.Collection(providerStateCollection),
		serverCollection:   db.Collection(serverUsageCollection),
	}
}

func (r *stateRepository) UpsertProviderState(ctx context.Context, state models.ProviderState) error {
	filter := bson.D{
		{Key: "serverId", Value: state.ServerID},
		{Key: "provider", Value: state.Provider},
	}
	update := bson.D{{
		Key: "$set",
		Value: bson.D{
			{Key: "serverId", Value: state.ServerID},
			{Key: "provider", Value: state.Provider},
			{Key: "updatedAt", Value: state.UpdatedAt},
			{Key: "available", Value: state.Available},
			{Key: "services", Value: state.Services},
		},
	}}
	opts := options.UpdateOne().SetUpsert(true)
	if _, err := r.providerCollection.UpdateOne(ctx, filter, update, opts); err != nil {
		return fmt.Errorf("upsert provider state %s/%s: %w", state.ServerID, state.Provider, err)
	}
	return nil
}

func (r *stateRepository) UpsertServerUsage(ctx context.Context, state models.ServerUsageState) error {
	filter := bson.D{{Key: "_id", Value: state.ServerID}}
	update := bson.D{{
		Key: "$set",
		Value: bson.D{
			{Key: "updatedAt", Value: state.UpdatedAt},
			{Key: "usage", Value: state.Usage},
		},
	}}
	opts := options.UpdateOne().SetUpsert(true)
	if _, err := r.serverCollection.UpdateOne(ctx, filter, update, opts); err != nil {
		return fmt.Errorf("upsert server usage %s: %w", state.ServerID, err)
	}
	return nil
}

func (r *stateRepository) GetProviderState(ctx context.Context, serverID, provider string) (*models.ProviderState, error) {
	var state models.ProviderState
	err := r.providerCollection.FindOne(ctx, bson.D{
		{Key: "serverId", Value: serverID},
		{Key: "provider", Value: provider},
	}).Decode(&state)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("get provider state %s/%s: %w", serverID, provider, err)
	}
	return &state, nil
}

func (r *stateRepository) GetServerUsage(ctx context.Context, serverID string) (*models.ServerUsageState, error) {
	var state models.ServerUsageState
	err := r.serverCollection.FindOne(ctx, bson.D{{Key: "_id", Value: serverID}}).Decode(&state)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("get server usage %s: %w", serverID, err)
	}
	return &state, nil
}