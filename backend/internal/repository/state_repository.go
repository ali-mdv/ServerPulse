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
	hostStateCollection     = "host_state"
	hostStateID             = "host"
)

type StateRepository interface {
	UpsertProviderState(ctx context.Context, state models.ProviderState) error
	UpsertHostState(ctx context.Context, state models.HostState) error
	GetProviderState(ctx context.Context, provider string) (*models.ProviderState, error)
	GetHostState(ctx context.Context) (*models.HostState, error)
}

type stateRepository struct {
	providerCollection *mongo.Collection
	hostCollection     *mongo.Collection
}

func NewStateRepository(db *mongo.Database) StateRepository {
	return &stateRepository{
		providerCollection: db.Collection(providerStateCollection),
		hostCollection:     db.Collection(hostStateCollection),
	}
}

func (r *stateRepository) UpsertProviderState(ctx context.Context, state models.ProviderState) error {
	filter := bson.D{{Key: "_id", Value: state.Provider}}
	update := bson.D{{
		Key: "$set",
		Value: bson.D{
			{Key: "updatedAt", Value: state.UpdatedAt},
			{Key: "available", Value: state.Available},
			{Key: "services", Value: state.Services},
		},
	}}
	opts := options.UpdateOne().SetUpsert(true)
	_, err := r.providerCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("upsert provider state %s: %w", state.Provider, err)
	}
	return nil
}

func (r *stateRepository) UpsertHostState(ctx context.Context, state models.HostState) error {
	filter := bson.D{{Key: "_id", Value: hostStateID}}
	update := bson.D{{
		Key: "$set",
		Value: bson.D{
			{Key: "updatedAt", Value: state.UpdatedAt},
			{Key: "usage", Value: state.Usage},
		},
	}}
	opts := options.UpdateOne().SetUpsert(true)
	_, err := r.hostCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("upsert host state: %w", err)
	}
	return nil
}

func (r *stateRepository) GetProviderState(ctx context.Context, provider string) (*models.ProviderState, error) {
	var state models.ProviderState
	err := r.providerCollection.FindOne(ctx, bson.D{{Key: "_id", Value: provider}}).Decode(&state)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("get provider state %s: %w", provider, err)
	}
	return &state, nil
}

func (r *stateRepository) GetHostState(ctx context.Context) (*models.HostState, error) {
	var state models.HostState
	err := r.hostCollection.FindOne(ctx, bson.D{{Key: "_id", Value: hostStateID}}).Decode(&state)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("get host state: %w", err)
	}
	return &state, nil
}
