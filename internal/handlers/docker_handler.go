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
	images, err := h.service.ImagesList(true)
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
		return

	}
	c.JSON(http.StatusOK, gin.H{
		"images": images,
	})
}

func (h *dockerHandler) GetDockerContainers(c *gin.Context) {
	containers, err := h.service.ContainersList(true)
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
		return

	}
	c.JSON(http.StatusOK, gin.H{
		"containers": containers,
	})
}

func (h *dockerHandler) GetContainerInfo(c *gin.Context) {
	containerId, _ := c.Params.Get("containerId")
	info, err := h.service.InspectContainer(containerId)
	if err != nil {
		switch e := err.(type) {
		case *errors.AppError:
			c.JSON(e.Code, gin.H{
				"error": e.Message,
			})
		default:
			c.JSON(errors.ErrInternalServer.Code, gin.H{
				"error": err.Error(),
			})
		}
		return

	}
	c.JSON(http.StatusOK, gin.H{
		"info": info,
	})
}

func (h *dockerHandler) StartContainer(c *gin.Context) {
	containerId, _ := c.Params.Get("containerId")
	err := h.service.StartContainer(containerId)
	if err != nil {
		switch e := err.(type) {
		case *errors.AppError:
			c.JSON(e.Code, gin.H{
				"error": e.Message,
			})
		default:
			c.JSON(errors.ErrInternalServer.Code, gin.H{
				"error": err.Error(),
			})
		}
		return

	}
	c.JSON(http.StatusOK, gin.H{
		"message": "container successfully start",
	})
}
