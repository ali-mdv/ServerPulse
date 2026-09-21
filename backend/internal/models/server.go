package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// ServerStatus reflects whether the backend has seen recent traffic
// from a registered server's agent. It is recomputed on every push by
// comparing the snapshot's timestamp against the freshness window.
type ServerStatus string

const (
	ServerStatusOnline  ServerStatus = "online"
	ServerStatusDown    ServerStatus = "down"
	ServerStatusUnknown ServerStatus = "unknown"
)

// AgentTokenBytes is the byte length of the opaque token generated for
// each server. The agent presents it on every push so the backend can
// authenticate the request and resolve the server it belongs to. Tokens
// are never serialised in normal API responses (json:"-") and are only
// revealed once at creation time.
const AgentTokenBytes = 32

// LocalServerName is the canonical name of the local host. The
// scheduler seeds a single Server record with this name on startup so
// the host where the backend runs is treated like any other server.
const LocalServerName = "local"

// Server is a registered host whose metrics the dashboard tracks.
//
// Two-way communication model: the backend exposes a push endpoint that
// the remote agent calls with its AgentToken in `Authorization: Bearer`.
// The token resolves to a single Server, which is what the rest of the
// system uses to scope snapshots, history, and provider state. The
// backend itself owns the local server — its agent loop runs in-process
// via the existing scheduler.
type Server struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string        `bson:"name" json:"name"`
	Host        string        `bson:"host" json:"host"`
	Port        int           `bson:"port" json:"port"`
	Description string        `bson:"description,omitempty" json:"description,omitempty"`

	Status     ServerStatus `bson:"status" json:"status"`
	LastSeen   *time.Time   `bson:"lastSeen,omitempty" json:"lastSeen,omitempty"`
	AgentToken string       `bson:"agentToken" json:"-"`

	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
}

// HexID returns the canonical string form used by every related type's
// ServerID field. Callers should always go through this so the format
// stays in one place.
func (s *Server) HexID() string {
	if s == nil {
		return ""
	}
	return s.ID.Hex()
}