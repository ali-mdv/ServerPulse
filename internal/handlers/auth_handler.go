package handlers

import (
	"net/http"
	"server-monitoring/internal/dtos"
	"server-monitoring/internal/services"

	"github.com/gin-gonic/gin"

	apperrors "server-monitoring/pkg/errors"
)

type authHandler struct {
	service services.AuthService
}

func NewAuthHandler(service services.AuthService) *authHandler {
	return &authHandler{service: service}
}

func (h *authHandler) Login(c *gin.Context) {
	var dto dtos.LoginDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": err.Error(),
		})
		return
	}

	token, err := h.service.GenerateAccessToken(dto)
	if err != nil {
		switch e := err.(type) {
		case *apperrors.AppError:
			c.JSON(e.Code, gin.H{
				"message": e.Message,
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Internal server error",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}
