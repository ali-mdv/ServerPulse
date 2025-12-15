package handlers

import (
	"errors"
	"net/http"
	"server-monitoring/internal/dtos"
	"server-monitoring/internal/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type userHandler struct {
	service services.UserService
}

func NewUserHandler(service services.UserService) *userHandler {
	return &userHandler{service: service}
}

func (h *userHandler) GetUsers(c *gin.Context) {
	users, err := h.service.UsersList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"users": users,
	})
}

func (h *userHandler) GetUserByID(c *gin.Context) {
	userID := c.Param("userID")

	user, err := h.service.FindUserByID(userID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "user not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

func (h *userHandler) CreateUser(c *gin.Context) {
	var body dtos.CreateUserDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := h.service.CreateUser(body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user": user,
	})
}

func (h *userHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("userID")

	var body dtos.UpdateUserDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := h.service.UpdateUserByID(userID, body)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "user not found",
			})
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}
