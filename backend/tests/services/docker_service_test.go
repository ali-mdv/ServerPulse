package services_test

import (
	"testing"

	setup_test "server-monitoring/tests/setup"
)

func TestFormatBytes(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{0, "0.00B"},
		{512, "512.00B"}, // < 1000 stays in B
		{1000, "1.00KB"},
		{1024, "1.02KB"},
		{1024 * 1024, "1.05MB"}, // 1024*1024 / 1000^2 = 1.0486
		{1024 * 1024 * 1024, "1.07GB"},
		{1024 * 1024 * 1024 * 1024, "1.10TB"},
	}
	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			got := setup_test.FormatBytes(tc.in)
			if got != tc.want {
				t.Fatalf("FormatBytes(%d) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestFormatBytes_NoExponentBeyondTB(t *testing.T) {
	// Even huge values stay in TB (no PB unit).
	got := setup_test.FormatBytes(1024 * 1024 * 1024 * 1024 * 1024)
	if got[len(got)-2:] != "TB" {
		t.Fatalf("expected TB suffix, got %q", got)
	}
}
