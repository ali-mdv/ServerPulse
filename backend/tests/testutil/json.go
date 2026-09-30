package testutil

import (
	"bytes"
	"encoding/json"
	"testing"
)

// MustJSON marshals v into a buffer or fails the test. Shared across
// handler tests.
func MustJSON(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return bytes.NewBuffer(b)
}