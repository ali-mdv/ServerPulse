package handlers

import (
	"net/http"
	"server-monitoring/internal/services"
	"server-monitoring/pkg/errors"

	"github.com/gin-gonic/gin"
)

type dockerHandler struct {
	service services.DockerService
}

func NewDockerHandler(service services.DockerService) *dockerHandler {
	return &dockerHandler{service: service}
}

func (h *dockerHandler) GetDockerImages(c *gin.Context) {
	images, err := h.service.ImageList()
	if err != nil {
		switch e := err.(type) {
		case *errors.AppError:
			c.JSON(e.Code, gin.H{
				"error": e.Message,
			})
		default:
			c.JSON(errors.ErrInternalServer.Code, gin.H{
				"error": errors.ErrInternalServer.Message,
			})
		}

	}
	c.JSON(http.StatusOK, gin.H{
		"images": images,
	})
}
