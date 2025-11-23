package services

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"server-monitoring/internal/models"
	"server-monitoring/pkg/errors"
	"strconv"
)

type PM2Service interface {
	List() ([]models.PM2Process, error)
	FindPM2ProcessByID(id int) (*models.PM2Process, error)
	StartPM2ProcessByID(id int) error
	StopPM2ProcessByID(id int) error
	RestartPM2ProcessByID(id int) error
	FetchContainerLogs(id int, lines int) (string, error)
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

func (s *pm2Service) FindPM2ProcessByID(id int) (*models.PM2Process, error) {
	processes, err := s.List()
	if err != nil {
		return nil, err
	}
	for _, p := range processes {
		if p.PMID == id {
			return &p, nil
		}
	}
	return nil, errors.ErrNotFound
}

func (s *pm2Service) StartPM2ProcessByID(id int) error {
	process, err := s.FindPM2ProcessByID(id)
	if err != nil {
		return err
	}
	cmd := exec.Command("pm2", "start", strconv.Itoa(process.PMID))
	_, err = cmd.Output()
	fmt.Printf("Error starting process %d: %v\n", id, err)
	if err != nil {
		return err
	}

	return nil
}

func (s *pm2Service) StopPM2ProcessByID(id int) error {
	process, err := s.FindPM2ProcessByID(id)
	if err != nil {
		return err
	}
	cmd := exec.Command("pm2", "stop", strconv.Itoa(process.PMID))
	_, err = cmd.Output()
	fmt.Printf("Error stopping process %d: %v\n", id, err)
	if err != nil {
		return err
	}

	return nil
}

func (s *pm2Service) RestartPM2ProcessByID(id int) error {
	process, err := s.FindPM2ProcessByID(id)
	if err != nil {
		return err
	}
	cmd := exec.Command("pm2", "restart", strconv.Itoa(process.PMID))
	_, err = cmd.Output()
	fmt.Printf("Error restarting process %d: %v\n", id, err)
	if err != nil {
		return err
	}

	return nil
}

func (s *pm2Service) FetchContainerLogs(id int, lines int) (string, error) {
	process, err := s.FindPM2ProcessByID(id)
	if err != nil {
		return "", err
	}
	cmd := exec.Command(
		"pm2",
		"logs",
		strconv.Itoa(process.PMID),
		"--lines",
		strconv.Itoa(lines),
		"--nostream",
	)

	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error fetching logs process %d: %v\n", id, err)
		return "", err
	}

	return string(out[1:]), nil
}
