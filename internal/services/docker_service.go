package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"server-monitoring/internal/models"
	"server-monitoring/pkg/errors"
	"strconv"
	"strings"
	"sync"
	"time"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

var ctx = context.Background()

type DockerService interface {
	ImagesList(all bool) ([]models.DockerImage, error)
	ContainersList(all bool) ([]models.DockerContainer, error)
	InspectContainer(id string) (container.InspectResponse, error)
	StartContainer(id string) error
	StopContainer(id string) error
	RestartContainer(id string) error
	FetchContainerLogs(id string, lines int) (string, error)
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
			Size:       summary.Size,
			CreatedAt:  time.Unix(summary.Created, 0).UTC(),
		}
		images = append(images, image)
	}

	return images, err
}

func (s *dockerService) ContainerStats(containerID string) *models.DockerContainerUsage {
	stats, err := s.client.ContainerStats(ctx, containerID, false)
	if err != nil {
		fmt.Printf("Error getting stats for %s: %v\n", containerID[:12], err)
		return nil
	}
	defer stats.Body.Close()

	var statsJson container.StatsResponse
	if err := json.NewDecoder(stats.Body).Decode(&statsJson); err != nil && err != io.EOF {
		fmt.Printf("Decode error for %s: %v\n", containerID[:12], err)
		return nil
	}

	// Calculate CPU percentage
	cpuDelta := float64(statsJson.CPUStats.CPUUsage.TotalUsage - statsJson.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(statsJson.CPUStats.SystemUsage - statsJson.PreCPUStats.SystemUsage)
	cpuPercent := 0.0
	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * float64(len(statsJson.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}

	memUsage := statsJson.MemoryStats.Usage
	memLimit := statsJson.MemoryStats.Limit
	memPercent := float64(memUsage) / float64(memLimit) * 100.0

	return &models.DockerContainerUsage{
		CpuPercent: cpuPercent,
		MemPercent: memPercent,
		MemUsage:   memUsage,
	}
}

func (s *dockerService) ContainersList(all bool) ([]models.DockerContainer, error) {
	containersSummary, err := s.client.ContainerList(ctx, container.ListOptions{All: all})
	if err != nil {
		return nil, err
	}

	containersInfo := make([]models.DockerContainer, len(containersSummary))
	var wg sync.WaitGroup
	wg.Add(len(containersSummary))

	for i, summary := range containersSummary {
		go func(i int, summary container.Summary) {
			defer wg.Done()

			containerID := strings.TrimPrefix(summary.ID, "sha256:")
			containerName := "<none>"
			if len(summary.Names) > 0 {
				containerName = strings.TrimLeft(summary.Names[0], "/")
			}

			port := "<none>"
			if len(summary.Ports) > 0 {
				p := summary.Ports[0]
				port = fmt.Sprintf("%d:%d", p.PublicPort, p.PrivatePort)
			}

			info := models.DockerContainer{
				ID:        containerID,
				Name:      containerName,
				Image:     summary.Image,
				Port:      port,
				State:     summary.State,
				UpTime:    summary.Status,
				CreatedAt: time.Unix(summary.Created, 0).UTC(),
			}

			// Run stats only for running containers
			if summary.State == "running" {
				if usage := s.ContainerStats(summary.ID); usage != nil {
					info.Usage = *usage
				}
			}

			containersInfo[i] = info
		}(i, summary)
	}

	wg.Wait()
	return containersInfo, nil
}

func (s *dockerService) InspectContainer(id string) (container.InspectResponse, error) {
	result, err := s.client.ContainerInspect(ctx, id)

	if err != nil {
		if cerrdefs.IsNotFound(err) {
			return container.InspectResponse{}, errors.New(http.StatusNotFound, "container with this id does not exists")
		}
		fmt.Printf("Error inspecting container %s: %v\n", id[:12], err)
		return container.InspectResponse{}, err
	}
	return result, nil
}

func (s *dockerService) StartContainer(id string) error {
	err := s.client.ContainerStart(ctx, id, container.StartOptions{})
	if err != nil {
		if cerrdefs.IsNotFound(err) {
			return errors.ErrNotFound
		}
		fmt.Printf("Error starting container %s: %v\n", id[:12], err)
		return err
	}
	return nil
}

func (s *dockerService) StopContainer(id string) error {
	err := s.client.ContainerStop(ctx, id, container.StopOptions{})
	if err != nil {
		if cerrdefs.IsNotFound(err) {
			return errors.ErrNotFound
		}
		fmt.Printf("Error stopping container %s: %v\n", id[:12], err)
		return err
	}
	return nil
}

func (s *dockerService) RestartContainer(id string) error {
	err := s.client.ContainerRestart(ctx, id, container.StopOptions{})
	if err != nil {
		if cerrdefs.IsNotFound(err) {
			return errors.ErrNotFound
		}
		fmt.Printf("Error restarting container %s: %v\n", id[:12], err)
		return err
	}
	return nil
}

func (s *dockerService) FetchContainerLogs(id string, lines int) (string, error) {

	reader, err := s.client.ContainerLogs(ctx, id, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Tail:       strconv.Itoa(lines),
	})

	if err != nil {
		if cerrdefs.IsNotFound(err) {
			return "", errors.ErrNotFound
		}
		fmt.Printf("Error fetching logs container %s: %v\n", id[:12], err)
		return "", err
	}

	defer reader.Close()

	logs, err := io.ReadAll(reader)
	if err != nil {
		fmt.Printf("Error fetching logs container %s: %v\n", id[:12], err)
		return "", err
	}

	return string(logs), nil
}
