package handlers

import (
	"net/http"
	"server-monitoring/internal/services"
	"server-monitoring/pkg/errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PM2Handler struct {
	service services.PM2Service
}

func NewPM2Handler(service services.PM2Service) *PM2Handler {
	return &PM2Handler{service: service}
}

func (h *PM2Handler) ProcessList(c *gin.Context) {
	processes, err := h.service.List()
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
		"processes": processes,
	})
}

func (h *PM2Handler) ProcessDetail(c *gin.Context) {
	idStr := c.Param("id")
	pmID, err := strconv.ParseInt(idStr, 10, 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid PM2 id; expected a numeric value",
		})
		return
	}

	process, err := h.service.FindPM2ProcessByID(int(pmID))
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
		"process": process,
	})
}

func (h *PM2Handler) StartProcess(c *gin.Context) {
	idStr := c.Param("id")
	pmID, err := strconv.ParseInt(idStr, 10, 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid PM2 id; expected a numeric value",
		})
		return
	}

	err = h.service.StartPM2ProcessByID(int(pmID))
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
		"message": "process successfully started",
	})
}

func (h *PM2Handler) StopProcess(c *gin.Context) {
	idStr := c.Param("id")
	pmID, err := strconv.ParseInt(idStr, 10, 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid PM2 id; expected a numeric value",
		})
		return
	}

	err = h.service.StopPM2ProcessByID(int(pmID))
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
		"message": "process successfully stopped",
	})
}

func (h *PM2Handler) RestartProcess(c *gin.Context) {
	idStr := c.Param("id")
	pmID, err := strconv.ParseInt(idStr, 10, 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid PM2 id; expected a numeric value",
		})
		return
	}

	err = h.service.RestartPM2ProcessByID(int(pmID))
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
		"message": "process successfully restarted",
	})
}
