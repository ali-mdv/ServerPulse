package services

import (
	"fmt"
	"server-monitoring/internal/dtos"
	"server-monitoring/internal/models"
	"server-monitoring/internal/repository"
	database "server-monitoring/pkg/mongo"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	UsersList() (*[]models.User, error)
	CreateUser(dtos.CreateUserDTO) (*models.User, error)
	FindUserByID(string) (*models.User, error)
	FindUserByEmail(string) (*models.User, error)
	UpdateUserByID(string, dtos.UpdateUserDTO) (*models.User, error)
	GenerateHash(string) (*string, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(dbName string) UserService {
	db := database.GetDatabase(dbName)
	userRepo := repository.NewUserRepository(db)
	return &userService{repo: userRepo}
}

func (s *userService) UsersList() (*[]models.User, error) {
	return s.repo.UsersList()
}

func (s *userService) FindUserByID(userID string) (*models.User, error) {
	return s.repo.FindUserByID(userID)
}

func (s *userService) FindUserByEmail(userEmail string) (*models.User, error) {
	return s.repo.FindUserByEmail(userEmail)
}

func (s *userService) CreateUser(dto dtos.CreateUserDTO) (*models.User, error) {
	hashedPassword, err := s.GenerateHash(dto.Password)
	if err != nil {
		return nil, err
	}
	dto.Password = *hashedPassword

	return s.repo.CreateUser(dto)
}

func (s *userService) UpdateUserByID(userID string, dto dtos.UpdateUserDTO) (*models.User, error) {
	if dto.Password != nil {
		hashedPassword, err := s.GenerateHash(*dto.Password)
		if err != nil {
			return nil, err
		}
		dto.Password = hashedPassword
	}

	_, err := s.repo.UpdateUserByID(userID, dto)
	if err != nil {
		return nil, err
	}

	user, _ := s.FindUserByID(userID)
	return user, nil
}

func (s *userService) GenerateHash(password string) (*string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	hash := string(hashedPassword)
	return &hash, nil
}
