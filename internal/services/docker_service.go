package services

import (
	"context"
	"fmt"
	"log"
	"server-monitoring/internal/models"
	"strings"
	"time"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

var ctx = context.Background()

type DockerService interface {
	ImageList() ([]models.DockerImage, error)
}

type dockerService struct {
	client *client.Client
}

func (s *dockerService) humanSize(size int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	f := float64(size)
	i := 0
	for f >= 1000 && i < len(units)-1 {
		f /= 1000
		i++
	}
	return fmt.Sprintf("%.2f%s", f, units[i])
}

func NewDockerService() DockerService {
	client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Error creating Docker client: %v", err)
	}

	return &dockerService{client: client}
}

func (s *dockerService) ImageList() ([]models.DockerImage, error) {
	imagesSummary, err := s.client.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return nil, err
	}

	images := make([]models.DockerImage, 0, len(imagesSummary))

	for _, summary := range imagesSummary {
		idParts := strings.Split(summary.ID, ":")
		repoParts := strings.Split(summary.RepoTags[0], ":")
		image := models.DockerImage{
			ID:         idParts[1],
			Repository: repoParts[0],
			Tag:        repoParts[1],
			Size:       s.humanSize(summary.Size),
			CreatedAt:  time.Unix(summary.Created, 0).UTC(),
		}
		images = append(images, image)
	}

	return images, err
}
