package services

import (
	"context"
	"fmt"
	"server-monitoring/internal/dtos"
	"server-monitoring/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	UsersList() (*[]models.User, error)
	CreateUser(dtos.CreateUserDTO) (*models.User, error)
	FindUserByID(string) (*models.User, error)
	UpdateUserByID(string, dtos.UpdateUserDTO) (*models.User, error)
}

type userService struct {
	db *mongo.Collection
}

func NewUserService(db *mongo.Collection) UserService {
	return &userService{db: db}
}

func (s *userService) UsersList() (*[]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := s.db.Find(ctx, bson.D{})
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

	return &users, nil
}

func (s *userService) FindUserByID(userID string) (*models.User, error) {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %v", err)
	}

	var user models.User
	err = s.db.FindOne(context.TODO(), bson.D{
		{Key: "_id", Value: objectID},
	}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *userService) CreateUser(dto dtos.CreateUserDTO) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := models.User{
		ID:        bson.NewObjectID(),
		Username:  dto.Username,
		Password:  string(hash),
		Email:     dto.Email,
		UpdatedAt: time.Now(),
		CreatedAt: time.Now(),
	}
	_, err = s.db.InsertOne(context.TODO(), user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *userService) UpdateUserByID(userID string, dto dtos.UpdateUserDTO) (*models.User, error) {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %v", err)
	}

	update := bson.D{}

	if dto.Username != nil {
		update = append(update, bson.E{Key: "$set", Value: bson.D{{Key: "username", Value: *dto.Username}}})
	}
	if dto.Email != nil {
		update = append(update, bson.E{Key: "$set", Value: bson.D{{Key: "email", Value: *dto.Email}}})
	}
	if dto.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*dto.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		update = append(update, bson.E{Key: "$set", Value: bson.D{{Key: "password", Value: string(hashedPassword)}}})
	}

	if len(update) != 0 {
		_, err = s.db.UpdateByID(context.TODO(), objectID, update)
		if err != nil {
			return nil, err
		}
	}

	user, _ := s.FindUserByID(userID)
	return user, nil
}
