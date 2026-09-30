package models

import "time"

// ServerUsageState stores the most recent system usage snapshot for a
// single server. It replaces the previous singleton HostState: the same
// shape, just keyed by server ID instead of a hard-coded "host" string.
//
// The repository upserts one row per server, so each Server has at
// most one ServerUsageState document. `_id` is the server's hex ID so
// lookups by server are O(1).
type ServerUsageState struct {
	ServerID  string      `bson:"_id" json:"serverId"`
	UpdatedAt time.Time   `bson:"updatedAt" json:"updatedAt"`
	Usage     SystemUsage `bson:"usage" json:"usage"`
}

// ProviderState stores the most recent list of services for a provider
// on a single server. Each (server, provider) pair has at most one row;
// the repository enforces uniqueness via a compound index on
// {serverId, provider} and filters on both fields for upserts/reads.
type ProviderState struct {
	ServerID  string            `bson:"serverId" json:"serverId"`
	Provider  string            `bson:"provider" json:"provider"`
	UpdatedAt time.Time         `bson:"updatedAt" json:"updatedAt"`
	Available bool              `bson:"available" json:"available"`
	Services  []ServiceSnapshot `bson:"services,omitempty" json:"services,omitempty"`
}