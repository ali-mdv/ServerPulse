package services

import (
	"encoding/json"
	"fmt"
	"log"

	"server-monitoring/internal/models"
	pm2pkg "server-monitoring/internal/services/pm2"
	"server-monitoring/pkg/errors"
)

type PM2Service interface {
	List() ([]models.PM2Process, error)
	FindPM2ProcessByID(id int) (*models.PM2Process, error)
	StartPM2ProcessByID(id int) error
	StopPM2ProcessByID(id int) error
	RestartPM2ProcessByID(id int) error
	FetchContainerLogs(id int, lines int) (string, error)
}

type pm2Service struct {
	client  *pm2pkg.Client
	dialErr error // surfaced to callers when the client never came up
}

// NewPM2Service connects to the host PM2 daemon via its unix socket.
// The socket path is the value of PM2_SOCKET_PATH (e.g. /root/.pm2/rpc.sock
// when /root/.pm2 is bind-mounted from the host).
func NewPM2Service(socketPath string) PM2Service {
	c, err := pm2pkg.Dial(socketPath)
	if err != nil {
		log.Printf("pm2 service: dial %s failed: %v", socketPath, err)
		return &pm2Service{client: nil, dialErr: err}
	}
	return &pm2Service{client: c}
}

// NewPM2ServiceFromDialErr builds a PM2Service that surfaces the given
// dial error as a 503 on every call. Used by the test/setup helpers;
// production callers should use NewPM2Service.
func NewPM2ServiceFromDialErr(err error) PM2Service {
	return &pm2Service{client: nil, dialErr: err}
}

func (s *pm2Service) notReady() error {
	if s.dialErr != nil {
		return errors.New(503, fmt.Sprintf("pm2 daemon unreachable: %s", s.dialErr))
	}
	return errors.New(503, "pm2 daemon unreachable")
}

func (s *pm2Service) List() ([]models.PM2Process, error) {
	if s.client == nil {
		return nil, s.notReady()
	}
	raw, err := s.client.GetMonitorData()
	if err != nil {
		return nil, err
	}
	procs := make([]models.PM2Process, 0, len(raw))
	for _, r := range raw {
		var p models.PM2Process
		if err := json.Unmarshal(r, &p); err != nil {
			return nil, fmt.Errorf("pm2: decode process: %w", err)
		}
		procs = append(procs, p)
	}
	return procs, nil
}

func (s *pm2Service) FindPM2ProcessByID(id int) (*models.PM2Process, error) {
	if s.client == nil {
		return nil, s.notReady()
	}
	procs, err := s.List()
	if err != nil {
		return nil, err
	}
	for _, p := range procs {
		if p.PMID == id {
			return &p, nil
		}
	}
	return nil, errors.ErrNotFound
}

func (s *pm2Service) StartPM2ProcessByID(id int) error {
	if s.client == nil {
		return s.notReady()
	}
	return s.client.StartProcessID(id)
}

func (s *pm2Service) StopPM2ProcessByID(id int) error {
	if s.client == nil {
		return s.notReady()
	}
	return s.client.StopProcessID(id)
}

func (s *pm2Service) RestartPM2ProcessByID(id int) error {
	if s.client == nil {
		return s.notReady()
	}
	return s.client.RestartProcessID(id)
}

func (s *pm2Service) FetchContainerLogs(id int, lines int) (string, error) {
	if s.client == nil {
		return "", s.notReady()
	}
	out, err := s.client.TailLogs(id, lines)
	if err != nil {
		// Mirror pm2-cli behaviour: missing log file is not fatal.
		if err.Error() == "no process found" {
			return "", errors.ErrNotFound
		}
		return "", err
	}
	return out, nil
}
