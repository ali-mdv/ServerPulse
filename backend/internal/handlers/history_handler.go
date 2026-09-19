package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"server-monitoring/internal/dtos"
	"server-monitoring/internal/services"
	"server-monitoring/pkg/errors"

	"github.com/gin-gonic/gin"
)

type historyHandler struct {
	service services.HistoryService
}

func NewHistoryHandler(service services.HistoryService) *historyHandler {
	return &historyHandler{service: service}
}

const (
	maxRangeDuration = 7 * 24 * time.Hour
	defaultRange     = time.Hour
)

func (h *historyHandler) ListServices(c *gin.Context) {
	provider := c.Param("provider")
	if !validProvider(provider) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider must be 'pm2' or 'docker'"})
		return
	}

	services, err := h.service.ListTrackedServices(c.Request.Context(), provider)
	if err != nil {
		respondHistoryError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"provider": provider,
		"services": dtos.ToTrackedServiceDTOs(services),
	})
}

func (h *historyHandler) GetSeries(c *gin.Context) {
	provider := c.Param("provider")
	serviceID := c.Param("serviceId")
	if !validProvider(provider) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider must be 'pm2' or 'docker'"})
		return
	}
	if strings.TrimSpace(serviceID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "serviceId is required"})
		return
	}

	from, to, bucket, err := parseRange(c, defaultRange)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	points, err := h.service.FindSeries(c.Request.Context(), provider, serviceID, from, to, bucket)
	if err != nil {
		respondHistoryError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"provider":  provider,
		"serviceId": serviceID,
		"from":      from,
		"to":        to,
		"bucket":    bucket.String(),
		"points":    dtos.ToSnapshotDTOs(points),
	})
}

func validProvider(p string) bool {
	return p == "pm2" || p == "docker" || p == "system"
}

func parseRange(c *gin.Context, def time.Duration) (time.Time, time.Time, time.Duration, error) {
	now := time.Now().UTC()
	to := now
	from := now.Add(-def)

	if v := c.Query("to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return time.Time{}, time.Time{}, 0, errInvalidTime("to")
		}
		to = t.UTC()
	}
	if v := c.Query("from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return time.Time{}, time.Time{}, 0, errInvalidTime("from")
		}
		from = t.UTC()
	}
	if !from.Before(to) {
		return time.Time{}, time.Time{}, 0, errInvalidTime("from must be before to")
	}
	if to.Sub(from) > maxRangeDuration {
		return time.Time{}, time.Time{}, 0, errInvalidTime("range exceeds 7 days")
	}

	var bucket time.Duration
	if v := c.Query("bucket"); v != "" {
		secs, err := strconv.ParseInt(v, 10, 64)
		if err != nil || secs <= 0 {
			return time.Time{}, time.Time{}, 0, errInvalidTime("bucket must be a positive integer (seconds)")
		}
		bucket = time.Duration(secs) * time.Second
	}

	return from, to, bucket, nil
}

type errInvalidTime string

func (e errInvalidTime) Error() string { return "invalid " + string(e) }

func respondHistoryError(c *gin.Context, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		c.JSON(appErr.Code, gin.H{"error": appErr.Message})
		return
	}
	c.JSON(errors.ErrInternalServer.Code, gin.H{"error": errors.ErrInternalServer.Message})
}
