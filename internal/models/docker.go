package models

import "time"

type DockerImage struct {
	ID         string    `json:"id"`
	Repository string    `json:"repository"`
	Tag        string    `json:"tag"`
	Size       string    `json:"size"`
	CreatedAt  time.Time `json:"createdAt"`
}

type DockerContainer struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Image     string    `json:"image"`
	Port      string    `json:"port"`
	State     string    `json:"state"`
	UpTime    string    `json:"upTime"`
	CreatedAt time.Time `json:"createdAt"`
}
