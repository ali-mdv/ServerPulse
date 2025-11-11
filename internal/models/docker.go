package models

import "time"

type DockerImage struct {
	ID         string    `json:"id"`
	Repository string    `json:"repository"`
	Tag        string    `json:"tag"`
	Size       string    `json:"size"`
	CreatedAt  time.Time `json:"createdAt"`
}
