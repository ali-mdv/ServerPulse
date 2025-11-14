package services

import (
	"encoding/json"
	"os/exec"
	"server-monitoring/internal/models"
)

type PM2Service interface {
	List() ([]models.PM2Process, error)
}

type pm2Service struct{}

func NewPM2Service() PM2Service {
	return &pm2Service{}
}

func (s *pm2Service) List() ([]models.PM2Process, error) {
	cmd := exec.Command("pm2", "jlist")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var processes []models.PM2Process
	if err := json.Unmarshal(out, &processes); err != nil {
		return nil, err
	}

	return processes, nil
}
