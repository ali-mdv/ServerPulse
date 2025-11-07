package services

import (
	"context"
	"server-monitoring/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type UserService interface {
	UsersList() ([]models.User, error)
}

type userService struct {
	db *mongo.Collection
}

func NewUserService(db *mongo.Collection) UserService {
	return &userService{db: db}
}

func (s *userService) UsersList() ([]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := s.db.Find(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []models.User
	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
