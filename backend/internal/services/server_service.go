package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"server-monitoring/internal/dtos"
	"server-monitoring/internal/models"
	"server-monitoring/internal/repository"
	database "server-monitoring/pkg/mongo"
	apperrors "server-monitoring/pkg/errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	ErrServerNameTaken = apperrors.ErrConflict
	ErrServerNotFound  = apperrors.ErrNotFound
)

type ServerService interface {
	Create(ctx context.Context, dto dtos.CreateServerDTO) (*models.Server, error)
	List(ctx context.Context) ([]models.Server, error)
	GetByID(ctx context.Context, id bson.ObjectID) (*models.Server, error)
	Update(ctx context.Context, id bson.ObjectID, dto dtos.UpdateServerDTO) (*models.Server, error)
	Delete(ctx context.Context, id bson.ObjectID) error
	GenerateAgentToken(ctx context.Context, id bson.ObjectID) (*models.Server, string, error)
	TouchSeen(ctx context.Context, id bson.ObjectID, status models.ServerStatus) error
	ResolveByAgentToken(ctx context.Context, token string) (*models.Server, error)
	EnsureLocalServer(ctx context.Context) (*models.Server, error)
}

type serverService struct {
	repo repository.ServerRepository
}

func NewServerService(dbName string) ServerService {
	db := database.GetDatabase(dbName)
	return &serverService{repo: repository.NewServerRepository(db)}
}

// Create registers a new server (agent) from its display name and an
// optional description. The API key is intentionally left empty; it is
// generated later via GenerateAgentToken so the creation step and the
// credential step can be audited separately.
func (s *serverService) Create(ctx context.Context, dto dtos.CreateServerDTO) (*models.Server, error) {
	if existing, err := s.repo.FindByName(ctx, dto.Name); err != nil {
		return nil, err
	} else if existing != nil {
		return nil, ErrServerNameTaken
	}

	now := time.Now().UTC()
	srv := models.Server{
		Name:        dto.Name,
		Host:        dto.Host,
		Port:        dto.Port,
		Description: dto.Description,
		Status:      models.ServerStatusUnknown,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return s.repo.Insert(ctx, srv)
}

func (s *serverService) List(ctx context.Context) ([]models.Server, error) {
	return s.repo.List(ctx)
}

func (s *serverService) GetByID(ctx context.Context, id bson.ObjectID) (*models.Server, error) {
	srv, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if srv == nil {
		return nil, ErrServerNotFound
	}
	return srv, nil
}

func (s *serverService) Update(ctx context.Context, id bson.ObjectID, dto dtos.UpdateServerDTO) (*models.Server, error) {
	set := bson.D{}
	if dto.Name != nil {
		set = append(set, bson.E{Key: "name", Value: *dto.Name})
	}
	if dto.Host != nil {
		set = append(set, bson.E{Key: "host", Value: *dto.Host})
	}
	if dto.Port != nil {
		set = append(set, bson.E{Key: "port", Value: *dto.Port})
	}
	if dto.Description != nil {
		set = append(set, bson.E{Key: "description", Value: *dto.Description})
	}
	if len(set) == 0 {
		return s.GetByID(ctx, id)
	}
	set = append(set, bson.E{Key: "updatedAt", Value: time.Now().UTC()})

	updated, err := s.repo.UpdateByID(ctx, id, bson.D{{Key: "$set", Value: set}})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrServerNameTaken
		}
		return nil, err
	}
	if updated == nil {
		return nil, ErrServerNotFound
	}
	return updated, nil
}

func (s *serverService) Delete(ctx context.Context, id bson.ObjectID) error {
	srv, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if srv == nil {
		return ErrServerNotFound
	}
	if srv.Name == models.LocalServerName {
		return apperrors.ErrForbidden
	}

	if err := s.repo.DeleteByID(ctx, id); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrServerNotFound
		}
		return err
	}
	return nil
}

// GenerateAgentToken mints a new opaque API key for the server and stores
// it on the record. The token is returned exactly once; subsequent calls
// rotate the key and return the new value, invalidating any previous key.
func (s *serverService) GenerateAgentToken(ctx context.Context, id bson.ObjectID) (*models.Server, string, error) {
	srv, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, "", err
	}
	if srv == nil {
		return nil, "", ErrServerNotFound
	}

	token, err := generateAgentToken()
	if err != nil {
		return nil, "", fmt.Errorf("server: %w", err)
	}

	update := bson.D{{
		Key: "$set",
		Value: bson.D{
			{Key: "agentToken", Value: token},
			{Key: "updatedAt", Value: time.Now().UTC()},
		},
	}}
	updated, err := s.repo.UpdateByID(ctx, id, update)
	if err != nil {
		return nil, "", err
	}
	if updated == nil {
		return nil, "", ErrServerNotFound
	}
	return updated, token, nil
}

func (s *serverService) TouchSeen(ctx context.Context, id bson.ObjectID, status models.ServerStatus) error {
	return s.repo.TouchSeen(ctx, id, status)
}

func (s *serverService) ResolveByAgentToken(ctx context.Context, token string) (*models.Server, error) {
	return s.repo.FindByAgentToken(ctx, token)
}

// EnsureLocalServer makes sure a row exists for the host the backend
// itself runs on. It is called once on startup; subsequent boots reuse
// the same record. The local server's token is left empty on purpose —
// the in-process scheduler writes its data directly, no agent needed.
func (s *serverService) EnsureLocalServer(ctx context.Context) (*models.Server, error) {
	existing, err := s.repo.FindByName(ctx, models.LocalServerName)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}
	now := time.Now().UTC()
	srv := models.Server{
		Name:        models.LocalServerName,
		Host:        "127.0.0.1",
		Port:        0,
		Description: "Backend host (auto-registered)",
		Status:      models.ServerStatusUnknown,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return s.repo.Insert(ctx, srv)
}

func generateAgentToken() (string, error) {
	buf := make([]byte, models.AgentTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}