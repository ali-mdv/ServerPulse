package models

import "time"

// SnapshotMeta is the bucketing key for the service_history time-series
// collection. It now carries the server origin so a single provider on
// multiple servers (e.g. two PM2 daemons) can coexist without
// collisions and history queries can be scoped to a specific server.
//
// MongoDB treats missing fields on old rows as empty strings, so the
// pre-server history can still be read back; new writes always include
// the field.
type SnapshotMeta struct {
	ServerID  string `bson:"serverId" json:"serverId"`
	Provider  string `bson:"provider" json:"provider"`
	ServiceID string `bson:"serviceId" json:"serviceId"`
	Name      string `bson:"name" json:"name"`
}

type ServiceSnapshot struct {
	Ts        time.Time    `bson:"ts" json:"ts"`
	Meta      SnapshotMeta `bson:"meta" json:"meta"`
	Status    string       `bson:"status,omitempty" json:"status,omitempty"`
	Available bool         `bson:"available" json:"available"`
	CPU       float64      `bson:"cpu,omitempty" json:"cpu,omitempty"`
	Memory    float64      `bson:"memory,omitempty" json:"memory,omitempty"`
	MemUsage  float64      `bson:"memUsage,omitempty" json:"memUsage,omitempty"`
	Disk      float64      `bson:"disk,omitempty" json:"disk,omitempty"`
	NetSent   float64      `bson:"netSent,omitempty" json:"netSent,omitempty"`
	NetRecv   float64      `bson:"netRecv,omitempty" json:"netRecv,omitempty"`
}

const (
	ProviderPM2    = "pm2"
	ProviderDocker = "docker"
	ProviderSystem = "system"

	ProviderStatusSentinelID   = "__provider__"
	ProviderStatusSentinelName = "__provider__"

	SystemMetricCPU        = "cpu"
	SystemMetricMemory     = "memory"
	SystemMetricDisk       = "disk"
	SystemMetricNetSent    = "network_sent"
	SystemMetricNetRecv    = "network_recv"
)
