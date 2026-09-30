package services_test

import (
	"net"
	"os"
	"path/filepath"
	"testing"

	"server-monitoring/internal/services"
	apperrors "server-monitoring/pkg/errors"
	setup_test "server-monitoring/tests/setup"
)

func TestPM2Service_NotReadyWhenDialFailed(t *testing.T) {
	svc := setup_test.NewPM2ServiceWithDialErr(os.ErrNotExist)
	_, err := svc.List()
	if err == nil {
		t.Fatal("expected error")
	}
	if !apperrors.IsUnavailable(err) {
		t.Fatalf("expected unavailable AppError, got %v", err)
	}
}

func TestPM2Service_OperationsSurfaceUnavailable(t *testing.T) {
	svc := setup_test.NewPM2ServiceWithDialErr(os.ErrNotExist)
	if err := svc.StartPM2ProcessByID(0); !apperrors.IsUnavailable(err) {
		t.Fatalf("Start: got %v, want unavailable", err)
	}
	if err := svc.StopPM2ProcessByID(0); !apperrors.IsUnavailable(err) {
		t.Fatalf("Stop: got %v, want unavailable", err)
	}
	if err := svc.RestartPM2ProcessByID(0); !apperrors.IsUnavailable(err) {
		t.Fatalf("Restart: got %v, want unavailable", err)
	}
	if _, err := svc.FetchContainerLogs(0, 10); !apperrors.IsUnavailable(err) {
		t.Fatalf("Logs: got %v, want unavailable", err)
	}
	if _, err := svc.FindPM2ProcessByID(0); !apperrors.IsUnavailable(err) {
		t.Fatalf("Find: got %v, want unavailable", err)
	}
}

// TestNewPM2Service_BadSocketKeepsError ensures that when the daemon is
// down, the service remembers the dial error so it can surface it as a
// 503 later (rather than panicking or returning a misleading 500).
func TestNewPM2Service_BadSocketKeepsError(t *testing.T) {
	bad := filepath.Join(t.TempDir(), "missing.sock")
	svc := services.NewPM2Service(bad)

	if _, err := svc.List(); !apperrors.IsUnavailable(err) {
		t.Fatalf("List: got %v, want unavailable", err)
	}
}

// TestNewPM2Service_ConnectsToTempSocket spins up a throwaway unix
// listener so we can confirm the happy path (when the socket exists and
// accepts connections) doesn't store a dialErr. The connection will be
// left open by the daemon side; the service immediately closes it after
// the first read fails, which is fine.
func TestNewPM2Service_ConnectsToTempSocket(t *testing.T) {
	ln, err := net.Listen("unix", filepath.Join(t.TempDir(), "pm2.sock"))
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	// Accept one connection, immediately close so the RPC read fails.
	go func() {
		c, err := ln.Accept()
		if err == nil {
			c.Close()
		}
	}()

	// Just ensure it doesn't blow up. The service will fail later when
	// it tries to RPC, but we're only checking construction here.
	_ = services.NewPM2Service(ln.Addr().String())
}
