package errors_test

import (
	"errors"
	"testing"

	apperrors "server-monitoring/pkg/errors"
)

func TestNew(t *testing.T) {
	err := apperrors.New(404, "error 404")
	if err.Code != 404 {
		t.Fatalf("code = %d, want 404", err.Code)
	}
	if err.Message != "error 404" {
		t.Fatalf("message = %q", err.Message)
	}
	if err.Error() != "error 404" {
		t.Fatalf("Error() = %q, want %q", err.Error(), "error 404")
	}
}

func TestWrap(t *testing.T) {
	err := apperrors.Wrap(500, "boom", "internal=exploded")
	if err.Code != 500 {
		t.Fatalf("code = %d, want 500", err.Code)
	}
	if err.Message != "boom" {
		t.Fatalf("message = %q", err.Message)
	}
}

func TestIsUnavailable(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"503 app error is unavailable", apperrors.New(503, "down"), true},
		{"500 app error is not unavailable", apperrors.New(500, "boom"), false},
		{"plain error is not unavailable", errors.New("plain"), false},
		{"nil is not unavailable", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := apperrors.IsUnavailable(tc.err); got != tc.want {
				t.Fatalf("IsUnavailable(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestCommonErrors(t *testing.T) {
	// Sanity: every sentinel must be an *AppError with the right code,
	// so the handler switches in handlers/* keep working.
	cases := []struct {
		err  *apperrors.AppError
		code int
	}{
		{apperrors.ErrBadRequest, 400},
		{apperrors.ErrValidation, 422},
		{apperrors.ErrUnauthorized, 401},
		{apperrors.ErrForbidden, 403},
		{apperrors.ErrNotFound, 404},
		{apperrors.ErrConflict, 409},
		{apperrors.ErrInternalServer, 500},
	}
	for _, tc := range cases {
		t.Run(tc.err.Message, func(t *testing.T) {
			if tc.err.Code != tc.code {
				t.Fatalf("%s: code = %d, want %d", tc.err.Message, tc.err.Code, tc.code)
			}
		})
	}
}