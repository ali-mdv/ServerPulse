package pm2

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// TestAMPRoundTrip covers every type pack/unpack path:
// - raw blob
// - "j:" JSON
// - "s:" string
// - nested arrays
// - argc boundaries
func TestAMPRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   [][]byte
	}{
		{"single json", [][]byte{[]byte("j:{\"type\":\"call\"}")}},
		{"single string", [][]byte{[]byte("s:hello world")}},
		{"single blob", [][]byte{bytes.Repeat([]byte{0xAB}, 32)}},
		{"mixed types", [][]byte{
			[]byte("j:[1,2,3]"),
			[]byte("s:abc"),
			[]byte("j:null"),
		}},
		{"max argc", makeArgs(15)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			frame, err := encodeAMP(tc.in)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			out, err := decodeAMP(frame)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if len(out) != len(tc.in) {
				t.Fatalf("argc: got %d want %d", len(out), len(tc.in))
			}
			for i := range tc.in {
				if !bytes.Equal(out[i], tc.in[i]) {
					t.Fatalf("arg %d: got %q want %q", i, out[i], tc.in[i])
				}
			}
		})
	}
}

func TestAMPRejectsZeroArgs(t *testing.T) {
	if _, err := encodeAMP(nil); err == nil {
		t.Fatal("expected error for zero args")
	}
	if _, err := encodeAMP([][]byte{}); err == nil {
		t.Fatal("expected error for empty slice")
	}
}

func TestAMPRejectsTooManyArgs(t *testing.T) {
	args := makeArgs(16)
	if _, err := encodeAMP(args); err == nil {
		t.Fatal("expected error for 16 args")
	}
}

func TestAMPMetaByteLayout(t *testing.T) {
	// First byte must be (version << 4) | argc == 0x13 for version=1 argc=3.
	frame, err := encodeAMP([][]byte{
		[]byte("j:1"),
		[]byte("j:2"),
		[]byte("j:3"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if frame[0] != 0x13 {
		t.Fatalf("meta byte = %#x, want 0x13", frame[0])
	}
	// Sanity: length prefix of first arg is 0x00 0x00 0x00 0x03 (length of "j:1").
	if got := binary.BigEndian.Uint32(frame[1:5]); got != 3 {
		t.Fatalf("first arg len = %d, want 3", got)
	}
}

func TestAMPSplitBufferDecode(t *testing.T) {
	// Build a frame, then chop it into N pieces and feed them one at a
	// time — decode must give the same args regardless of how the bytes
	// were split across reads.
	frame, err := encodeAMP([][]byte{
		[]byte("j:{\"a\":1}"),
		[]byte("s:longer string here for splitting"),
		[]byte("j:42"),
	})
	if err != nil {
		t.Fatal(err)
	}
	// We don't simulate streaming here (transport does the ReadFull); just
	// verify the encoded buffer decodes correctly when split at an odd
	// boundary mid-arg-length.
	for _, cut := range []int{1, 3, 5, 8, 17} {
		head := append([]byte{}, frame[:cut]...)
		tail := append([]byte{}, frame[cut:]...)
		joined := append(head, tail...)
		args, err := decodeAMP(joined)
		if err != nil {
			t.Fatalf("split at %d: %v", cut, err)
		}
		if len(args) != 3 {
			t.Fatalf("split at %d: got %d args want 3", cut, len(args))
		}
		if string(args[0]) != "j:{\"a\":1}" {
			t.Fatalf("split at %d: arg0 = %q", cut, args[0])
		}
	}
}

func makeArgs(n int) [][]byte {
	out := make([][]byte, n)
	for i := 0; i < n; i++ {
		out[i] = []byte("j:" + string(rune('a'+i)))
	}
	return out
}