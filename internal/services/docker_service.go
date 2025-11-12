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
	ImagesList(all bool) ([]models.DockerImage, error)
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

func (s *dockerService) ImagesList(all bool) ([]models.DockerImage, error) {
	imagesSummary, err := s.client.ImageList(ctx, image.ListOptions{All: all})
	if err != nil {
		return nil, err
	}

	images := make([]models.DockerImage, 0, len(imagesSummary))

	for _, summary := range imagesSummary {

		imageID := summary.ID
		if strings.Contains(imageID, ":") {
			parts := strings.Split(imageID, ":")
			if len(parts) > 1 {
				imageID = parts[1]
			}
		}

		repository := "<none>"
		tag := "<none>"

		if len(summary.RepoTags) > 0 {
			repoTag := summary.RepoTags[0]
			repoParts := strings.Split(repoTag, ":")

			if len(repoParts) >= 2 {
				repository = repoParts[0]
				tag = repoParts[1]
			} else if len(repoParts) == 1 {
				repository = repoParts[0]
			}
		}

		image := models.DockerImage{
			ID:         imageID,
			Repository: repository,
			Tag:        tag,
			Size:       s.humanSize(summary.Size),
			CreatedAt:  time.Unix(summary.Created, 0).UTC(),
		}
		images = append(images, image)
	}

	return images, err
}
