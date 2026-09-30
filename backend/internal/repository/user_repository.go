package repository

import (
	"context"
	"fmt"
	"server-monitoring/internal/dtos"
	"server-monitoring/internal/models"
	"server-monitoring/pkg/errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type UserRepository interface {
	UsersList() (*[]models.User, error)
	FindUserByID(string) (*models.User, error)
	FindUserByEmail(string) (*models.User, error)
	CreateUser(dtos.CreateUserDTO) (*models.User, error)
	UpdateUserByID(string, dtos.UpdateUserDTO) (*models.User, error)
	DeleteUserByID(string) error
}

type userRepository struct {
	Collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) UserRepository {
	repo := &userRepository{
		Collection: db.Collection("users"),
	}
	repo.setupIndexes()
	return repo
}

func (r *userRepository) setupIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "email", Value: 1}},
			Options: options.Index().
				SetUnique(true).
				SetName("unique_email"),
		},
	}

	_, err := r.Collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		// You might log this instead of panicking
		panic("failed to create user indexes: " + err.Error())
	}
}

func (r *userRepository) UsersList() (*[]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.Collection.Find(ctx, bson.D{})
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

func (r *userRepository) FindUserByID(userID string) (*models.User, error) {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %v", err)
	}

	var user models.User
	err = r.Collection.FindOne(context.TODO(), bson.D{
		{Key: "_id", Value: objectID},
	}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) FindUserByEmail(userEmail string) (*models.User, error) {
	var user models.User
	err := r.Collection.FindOne(context.TODO(), bson.D{
		{Key: "email", Value: userEmail},
	}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) CreateUser(dto dtos.CreateUserDTO) (*models.User, error) {
	user := models.User{
		ID:        bson.NewObjectID(),
		Password:  dto.Password,
		Email:     dto.Email,
		UpdatedAt: time.Now(),
		CreatedAt: time.Now(),
	}
	_, err := r.Collection.InsertOne(context.TODO(), user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
func (r *userRepository) UpdateUserByID(userID string, data dtos.UpdateUserDTO) (*models.User, error) {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %v", err)
	}

	set := bson.D{}

	if data.Email != nil {
		existing, _ := r.FindUserByEmail(*data.Email)
		if existing != nil && existing.ID != objectID {
			return nil, errors.ErrConflict
		}
		set = append(set, bson.E{Key: "email", Value: *data.Email})
	}
	if data.Password != nil {
		set = append(set, bson.E{Key: "password", Value: *data.Password})
	}

	if len(set) != 0 {
		set = append(set, bson.E{Key: "updatedAt", Value: time.Now()})
		update := bson.D{{Key: "$set", Value: set}}
		if _, err = r.Collection.UpdateByID(context.TODO(), objectID, update); err != nil {
			return nil, err
		}
	}

	return r.FindUserByID(userID)
}

func (r *userRepository) DeleteUserByID(userID string) error {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %v", err)
	}

	res, err := r.Collection.DeleteOne(context.TODO(), bson.D{
		{Key: "_id", Value: objectID},
	})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}
