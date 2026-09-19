package models

import "time"

// ProviderState stores the most recent list of services for a provider.
// It is updated by the scheduler on every tick so the dashboard can read
// the latest known state from MongoDB instead of hitting the host daemons.
type ProviderState struct {
	Provider  string             `bson:"_id" json:"provider"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
	Available bool               `bson:"available" json:"available"`
	Services  []ServiceSnapshot  `bson:"services,omitempty" json:"services,omitempty"`
}

// HostState stores the most recent system usage snapshot.
type HostState struct {
	ID        string      `bson:"_id" json:"id"`
	UpdatedAt time.Time   `bson:"updatedAt" json:"updatedAt"`
	Usage     SystemUsage `bson:"usage" json:"usage"`
}
