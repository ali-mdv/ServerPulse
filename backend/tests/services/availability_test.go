package services_test

import (
	"testing"
	"time"

	"server-monitoring/internal/services"
)

func TestStaleAfter_AppliesTolerance(t *testing.T) {
	if got := services.StaleAfter(30 * time.Second); got != 60*time.Second {
		t.Fatalf("StaleAfter(30s) = %v, want 60s", got)
	}
	if got := services.StaleAfter(90 * time.Second); got != 3*time.Minute {
		t.Fatalf("StaleAfter(90s) = %v, want 3m", got)
	}
	if got := services.StaleAfter(0); got != 0 {
		t.Fatalf("StaleAfter(0) = %v, want 0", got)
	}
}
