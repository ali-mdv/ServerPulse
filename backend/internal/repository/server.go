package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"server-monitoring/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const serverCollection = "servers"

type ServerRepository interface {
	Insert(ctx context.Context, s models.Server) (*models.Server, error)
	FindByID(ctx context.Context, id bson.ObjectID) (*models.Server, error)
	FindByName(ctx context.Context, name string) (*models.Server, error)
	FindByAgentToken(ctx context.Context, token string) (*models.Server, error)
	List(ctx context.Context) ([]models.Server, error)
	UpdateByID(ctx context.Context, id bson.ObjectID, update bson.D) (*models.Server, error)
	DeleteByID(ctx context.Context, id bson.ObjectID) error
	TouchSeen(ctx context.Context, id bson.ObjectID, status models.ServerStatus) error
}

type serverRepository struct {
	Collection *mongo.Collection
}

func NewServerRepository(db *mongo.Database) ServerRepository {
	return &serverRepository{Collection: db.Collection(serverCollection)}
}

func (r *serverRepository) Insert(ctx context.Context, s models.Server) (*models.Server, error) {
	if s.ID.IsZero() {
		s.ID = bson.NewObjectID()
	}
	if _, err := r.Collection.InsertOne(ctx, s); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, fmt.Errorf("server: %w", err)
		}
		return nil, fmt.Errorf("insert server: %w", err)
	}
	return &s, nil
}

func (r *serverRepository) FindByID(ctx context.Context, id bson.ObjectID) (*models.Server, error) {
	var s models.Server
	if err := r.Collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&s); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("find server by id: %w", err)
	}
	return &s, nil
}

func (r *serverRepository) FindByName(ctx context.Context, name string) (*models.Server, error) {
	var s models.Server
	if err := r.Collection.FindOne(ctx, bson.D{{Key: "name", Value: name}}).Decode(&s); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("find server by name: %w", err)
	}
	return &s, nil
}

func (r *serverRepository) FindByAgentToken(ctx context.Context, token string) (*models.Server, error) {
	if token == "" {
		return nil, nil
	}
	var s models.Server
	if err := r.Collection.FindOne(ctx, bson.D{{Key: "agentToken", Value: token}}).Decode(&s); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("find server by token: %w", err)
	}
	return &s, nil
}

func (r *serverRepository) List(ctx context.Context) ([]models.Server, error) {
	cursor, err := r.Collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("list servers: %w", err)
	}
	defer cursor.Close(ctx)

	var out []models.Server
	if err := cursor.All(ctx, &out); err != nil {
		return nil, fmt.Errorf("decode servers: %w", err)
	}
	return out, nil
}

func (r *serverRepository) UpdateByID(ctx context.Context, id bson.ObjectID, update bson.D) (*models.Server, error) {
	if len(update) == 0 {
		return r.FindByID(ctx, id)
	}
	if _, err := r.Collection.UpdateByID(ctx, id, update); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, fmt.Errorf("server: %w", err)
		}
		return nil, fmt.Errorf("update server: %w", err)
	}
	return r.FindByID(ctx, id)
}

func (r *serverRepository) DeleteByID(ctx context.Context, id bson.ObjectID) error {
	res, err := r.Collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		return fmt.Errorf("delete server: %w", err)
	}
	if res.DeletedCount == 0 {
		return fmt.Errorf("delete server: not found")
	}
	return nil
}

// TouchSeen stamps LastSeen + Status without touching any other field.
// Called by the agent ingestion path so a successful push refreshes the
// heartbeat that powers ServerStatusOnline/ServerStatusDown decisions.
func (r *serverRepository) TouchSeen(ctx context.Context, id bson.ObjectID, status models.ServerStatus) error {
	update := bson.D{{
		Key: "$set",
		Value: bson.D{
			{Key: "lastSeen", Value: time.Now().UTC()},
			{Key: "status", Value: status},
		},
	}}
	if _, err := r.Collection.UpdateByID(ctx, id, update); err != nil {
		return fmt.Errorf("touch server: %w", err)
	}
	return nil
}