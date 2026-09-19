package handlers

import (
	"context"
	"net/http"

	"server-monitoring/internal/dtos"
	"server-monitoring/internal/models"
	"server-monitoring/internal/services"
	apperrors "server-monitoring/pkg/errors"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ServerHandler struct {
	serverSvc services.ServerService
	stateSvc  services.StateService
}

func NewServerHandler(serverSvc services.ServerService, stateSvc services.StateService) *ServerHandler {
	return &ServerHandler{serverSvc: serverSvc, stateSvc: stateSvc}
}

// List returns every registered server with its latest usage + provider
// roll-up so the dashboard can render the cards without follow-ups.
func (h *ServerHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	servers, err := h.serverSvc.List(ctx)
	if err != nil {
		respondServerError(c, err)
		return
	}

	items := make([]dtos.ServerListItemDTO, 0, len(servers))
	for _, s := range servers {
		items = append(items, h.toListItem(ctx, s))
	}

	c.JSON(http.StatusOK, gin.H{"servers": items})
}

// Get returns a single server with the same roll-up as List.
func (h *ServerHandler) Get(c *gin.Context) {
	id, err := bson.ObjectIDFromHex(c.Param("serverId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid server id"})
		return
	}
	ctx := c.Request.Context()
	srv, err := h.serverSvc.GetByID(ctx, id)
	if err != nil {
		respondServerError(c, err)
		return
	}

	dto := dtos.ToServerDTO(*srv, nil, nil)
	if usage, err := h.stateSvc.GetServerUsage(ctx, srv.HexID()); err == nil && usage != nil {
		dto.Usage = dtos.ToServerUsageDTO(*usage)
	}
	if providers, err := h.collectProviders(ctx, srv.HexID()); err == nil && len(providers) > 0 {
		dto.Providers = dtos.ToServerProviderAvailabilityDTOs(providers)
	}

	c.JSON(http.StatusOK, gin.H{"server": dto})
}

// Create registers a new server and returns the freshly minted token
// exactly once. Subsequent reads (Get/List) never expose the token.
func (h *ServerHandler) Create(c *gin.Context) {
	var body dtos.CreateServerDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	srv, err := h.serverSvc.Create(c.Request.Context(), body)
	if err != nil {
		respondServerError(c, err)
		return
	}

	dto := dtos.ToServerDTO(*srv, nil, nil)
	c.JSON(http.StatusCreated, gin.H{
		"server":     dto,
		"agentToken": srv.AgentToken,
	})
}

// Update mutates a server's editable fields. The agent token is never
// returned and cannot be changed through this endpoint.
func (h *ServerHandler) Update(c *gin.Context) {
	id, err := bson.ObjectIDFromHex(c.Param("serverId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid server id"})
		return
	}

	var body dtos.UpdateServerDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	srv, err := h.serverSvc.Update(c.Request.Context(), id, body)
	if err != nil {
		respondServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"server": dtos.ToServerDTO(*srv, nil, nil)})
}

// Delete removes a server. State/history rows are left in place so the
// dashboard can still render past charts for deleted hosts.
func (h *ServerHandler) Delete(c *gin.Context) {
	id, err := bson.ObjectIDFromHex(c.Param("serverId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid server id"})
		return
	}
	if err := h.serverSvc.Delete(c.Request.Context(), id); err != nil {
		respondServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "server deleted"})
}

// toListItem projects a server model plus its latest usage + provider
// roll-up onto the lighter list view used by List and Get.
func (h *ServerHandler) toListItem(ctx context.Context, s models.Server) dtos.ServerListItemDTO {
	dto := dtos.ServerListItemDTO{
		ID:        s.ID.Hex(),
		Name:      s.Name,
		Host:      s.Host,
		Status:    s.Status,
		LastSeen:  s.LastSeen,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}

	if usage, err := h.stateSvc.GetServerUsage(ctx, s.HexID()); err == nil && usage != nil {
		dto.Usage = dtos.ToServerUsageDTO(*usage)
	}
	if providers, err := h.collectProviders(ctx, s.HexID()); err == nil && len(providers) > 0 {
		dto.Providers = dtos.ToServerProviderAvailabilityDTOs(providers)
	}
	return dto
}

// collectProviders fetches the known-provider states for a server.
func (h *ServerHandler) collectProviders(ctx context.Context, serverID string) (map[string]models.ProviderState, error) {
	out := map[string]models.ProviderState{}
	for _, p := range []string{models.ProviderPM2, models.ProviderDocker} {
		state, err := h.stateSvc.GetProviderState(ctx, serverID, p)
		if err != nil {
			return nil, err
		}
		if state != nil {
			out[p] = *state
		}
	}
	return out, nil
}

func respondServerError(c *gin.Context, err error) {
	if appErr, ok := err.(*apperrors.AppError); ok {
		c.JSON(appErr.Code, gin.H{"error": appErr.Message})
		return
	}
	c.JSON(apperrors.ErrInternalServer.Code, gin.H{"error": apperrors.ErrInternalServer.Message})
}