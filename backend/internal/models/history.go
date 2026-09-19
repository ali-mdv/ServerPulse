package models

import "time"

type SnapshotMeta struct {
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
