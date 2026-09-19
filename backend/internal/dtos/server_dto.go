package dtos

import (
	"time"

	"server-monitoring/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// CreateServerDTO is the request body for POST /servers. The backend
// generates the AgentToken and assigns the ObjectID, so the caller only
// supplies identity + connection details.
type CreateServerDTO struct {
	Name        string `binding:"required,min=1,max=64" json:"name"`
	Host        string `binding:"required,min=1,max=255" json:"host"`
	Port        int    `binding:"required,min=1,max=65535" json:"port"`
	Description string `binding:"omitempty,max=255" json:"description"`
}

// UpdateServerDTO is the request body for PUT /servers/:id. Every
// field is a pointer so a partial update (e.g. rename only) leaves the
// rest untouched.
type UpdateServerDTO struct {
	Name        *string `binding:"omitempty,min=1,max=64" json:"name"`
	Host        *string `binding:"omitempty,min=1,max=255" json:"host"`
	Port        *int    `binding:"omitempty,min=1,max=65535" json:"port"`
	Description *string `binding:"omitempty,max=255" json:"description"`
}

// ServerDTO is the response shape for a single server. AgentToken is
// intentionally hidden from this view — it's only returned by the
// dedicated "reveal token" endpoint (TBD) and on the create response.
type ServerDTO struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Host        string             `json:"host"`
	Port        int                `json:"port"`
	Description string             `json:"description,omitempty"`
	Status      models.ServerStatus `json:"status"`
	LastSeen    *time.Time         `json:"lastSeen,omitempty"`
	CreatedAt   time.Time          `json:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt"`
	Usage       *ServerUsageDTO    `json:"usage,omitempty"`
	Providers   map[string]ServerProviderAvailabilityDTO `json:"providers,omitempty"`
}

// ServerListItemDTO is the trimmed shape used in list endpoints. It
// drops the description and embeds the latest usage + provider roll-up
// so the dashboard can render the card without a second round-trip.
type ServerListItemDTO struct {
	ID        string                  `json:"id"`
	Name      string                  `json:"name"`
	Host      string                  `json:"host"`
	Status    models.ServerStatus     `json:"status"`
	LastSeen  *time.Time              `json:"lastSeen,omitempty"`
	CreatedAt time.Time               `json:"createdAt"`
	UpdatedAt time.Time               `json:"updatedAt"`
	Usage     *ServerUsageDTO         `json:"usage,omitempty"`
	Providers map[string]ServerProviderAvailabilityDTO `json:"providers,omitempty"`
}

// ServerProviderAvailabilityDTO is a per-provider health summary nested
// in ServerDTO/ServerListItemDTO. The map key is the provider name
// ("pm2", "docker", ...).
type ServerProviderAvailabilityDTO struct {
	Available bool       `json:"available"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

// ToServerDTO maps a model plus its latest usage/provider roll-ups to
// the wire shape. Any of the roll-ups may be nil; they stay absent from
// the JSON in that case.
func ToServerDTO(s models.Server, usage *models.ServerUsageState, providers map[string]models.ProviderState) ServerDTO {
	dto := ServerDTO{
		ID:          s.ID.Hex(),
		Name:        s.Name,
		Host:        s.Host,
		Port:        s.Port,
		Description: s.Description,
		Status:      s.Status,
		LastSeen:    s.LastSeen,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
	if usage != nil {
		dto.Usage = ToServerUsageDTO(*usage)
	}
	if len(providers) > 0 {
		dto.Providers = ToServerProviderAvailabilityDTOs(providers)
	}
	return dto
}

// ToServerListItemDTO projects a model onto the lighter list view.
func ToServerListItemDTO(s models.Server, usage *models.ServerUsageState, providers map[string]models.ProviderState) ServerListItemDTO {
	dto := ServerListItemDTO{
		ID:        s.ID.Hex(),
		Name:      s.Name,
		Host:      s.Host,
		Status:    s.Status,
		LastSeen:  s.LastSeen,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
	if usage != nil {
		dto.Usage = ToServerUsageDTO(*usage)
	}
	if len(providers) > 0 {
		dto.Providers = ToServerProviderAvailabilityDTOs(providers)
	}
	return dto
}

// ToServerListItemDTOs is a small adapter for repository slices.
func ToServerListItemDTOs(in []models.Server) []ServerListItemDTO {
	out := make([]ServerListItemDTO, len(in))
	for i := range in {
		out[i] = ToServerListItemDTO(in[i], nil, nil)
	}
	return out
}

// ToServerDTOs maps a slice of models.
func ToServerDTOs(in []models.Server) []ServerDTO {
	out := make([]ServerDTO, len(in))
	for i := range in {
		out[i] = ToServerDTO(in[i], nil, nil)
	}
	return out
}

// ParseObjectID parses a hex string into an ObjectID, returning
// apperrors.ErrBadRequest (via the caller) when the input is malformed.
// Centralising the conversion here keeps handlers free of the import.
func ParseObjectID(hex string) (bson.ObjectID, error) {
	return bson.ObjectIDFromHex(hex)
}