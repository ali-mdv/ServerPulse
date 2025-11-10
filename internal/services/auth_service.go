package services

import (
	"server-monitoring/internal/dtos"
	"server-monitoring/pkg/errors"
	jwtutil "server-monitoring/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	GenerateAccessToken(dtos.LoginDTO) (string, error)
}

type authService struct {
	userService UserService
}

func NewAuthService(userService UserService) AuthService {
	return &authService{userService: userService}
}

func (s *authService) GenerateAccessToken(dto dtos.LoginDTO) (string, error) {
	user, err := s.userService.FindUserByEmail(dto.Email)
	if err != nil {
		return "", errors.ErrUnauthorized
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(dto.Password)) != nil {
		return "", errors.ErrUnauthorized
	}

	token, err := jwtutil.GenerateToken(user.ID.Hex())
	if err != nil {
		return "", errors.ErrInternalServer
	}

	return token, nil
}
