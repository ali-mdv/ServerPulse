package handlers

import (
	stderrors "errors"
	"net"
	"net/http"
	"strings"

	dockerclient "github.com/docker/docker/client"
	"server-monitoring/internal/services"
	"server-monitoring/pkg/errors"

	"github.com/gin-gonic/gin"
)

type dockerHandler struct {
	service services.DockerService
}

// isDockerUnavailable reports whether the docker daemon (or its socket) is
// unreachable -- a degraded, not fatal, condition for polling endpoints.
func isDockerUnavailable(err error) bool {
	if dockerclient.IsErrConnectionFailed(err) {
		return true
	}
	var opErr *net.OpError
	if stderrors.As(err, &opErr) {
		return true
	}
	return strings.Contains(err.Error(), "Cannot connect to the Docker daemon")
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
		// Polling endpoint: degrade instead of failing. The frontend hides
		// the Docker section when the daemon is not reachable.
		if isDockerUnavailable(err) {
			c.JSON(http.StatusOK, gin.H{
				"containers": []any{},
				"available":  false,
			})
			return
		}
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
		"available":  true,
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
				"error": errors.ErrInternalServer.Message,
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
				"error": errors.ErrInternalServer.Message,
			})
		}
		return

	}
	c.JSON(http.StatusOK, gin.H{
		"message": "container successfully started",
	})
}

func (h *dockerHandler) StopContainer(c *gin.Context) {
	containerId, _ := c.Params.Get("containerId")
	err := h.service.StopContainer(containerId)
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
		"message": "container successfully stopped",
	})
}

func (h *dockerHandler) RestartContainer(c *gin.Context) {
	containerId, _ := c.Params.Get("containerId")
	err := h.service.RestartContainer(containerId)
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
		"message": "container successfully restarted",
	})
}

func (h *dockerHandler) GetContainerLogs(c *gin.Context) {
	containerId, _ := c.Params.Get("containerId")
	logs, err := h.service.FetchContainerLogs(containerId, 100)
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
		"logs": logs,
	})
}
