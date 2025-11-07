package handlers

import (
	"net/http"
	"server-monitoring/internal/services"

	"github.com/gin-gonic/gin"
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
