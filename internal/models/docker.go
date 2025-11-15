package models

import (
	"time"

	"github.com/docker/docker/api/types/container"
)

type DockerContainerUsage struct {
	CpuPercent float64 `json:"cpuPercent"`
	MemPercent float64 `json:"memPercent"`
	MemUsage   uint64  `json:"memUsage"`
}

type DockerImage struct {
	ID         string    `json:"id"`
	Repository string    `json:"repository"`
	Tag        string    `json:"tag"`
	Size       int64     `json:"size"`
	CreatedAt  time.Time `json:"createdAt"`
}

type DockerContainer struct {
	ID        string                   `json:"id"`
	Name      string                   `json:"name"`
	Image     string                   `json:"image"`
	Port      string                   `json:"port"`
	State     container.ContainerState `json:"state"`
	UpTime    string                   `json:"upTime"`
	Usage     DockerContainerUsage     `json:"usage,omitzero"`
	CreatedAt time.Time                `json:"createdAt"`
}
