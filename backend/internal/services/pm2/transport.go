// Package pm2 talks to a PM2 daemon over its local unix socket using the
// axon-rpc protocol (AMP wire format + JSON codec). It is a drop-in for
// `exec.Command("pm2", ...)` so the backend container can stay minimal
// (distroless, no Node) while still reaching PM2 running on the host.
package pm2

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
)

// ampVersion is the protocol version byte for `visionmedia/node-amp`.
const ampVersion = 1

// ampMaxArgs is the maximum number of arguments per AMP message (4 bits).
const ampMaxArgs = 15

// encodeAMP packs the given argument buffers into a single AMP frame.
//
// Wire format (visionmedia/node-amp @ 0.3.1):
//
//	[0]                     : meta byte  = (version << 4) | argc
//	for each arg i in 0..argc:
//	  [off..off+4) (uint32 BE) : length of arg i
//	  [off..off+len]           : bytes of arg i
//
// Each arg is a *packed* buffer (see packJSON / packString), not a raw
// user value. Callers pack first, then pass the packed buffers here.
func encodeAMP(args [][]byte) ([]byte, error) {
	n := len(args)
	if n == 0 {
		return nil, errors.New("amp: refuse to encode zero-arg message")
	}
	if n > ampMaxArgs {
		return nil, fmt.Errorf("amp: too many args (%d > %d)", n, ampMaxArgs)
	}

	total := 1
	for _, a := range args {
		total += 4 + len(a)
	}

	buf := make([]byte, total)
	buf[0] = (ampVersion << 4) | byte(n)

	off := 1
	for _, a := range args {
		binary.BigEndian.PutUint32(buf[off:], uint32(len(a)))
		off += 4
		copy(buf[off:], a)
		off += len(a)
	}
	return buf, nil
}

// decodeAMP parses one AMP frame and returns the raw argument buffers.
func decodeAMP(buf []byte) ([][]byte, error) {
	if len(buf) == 0 {
		return nil, errors.New("amp: empty frame")
	}
	meta := buf[0]
	version := meta >> 4
	argc := int(meta & 0x0f)
	if version != ampVersion {
		return nil, fmt.Errorf("amp: unsupported version %d", version)
	}
	if argc > ampMaxArgs {
		return nil, fmt.Errorf("amp: invalid argc %d", argc)
	}

	args := make([][]byte, argc)
	off := 1
	for i := 0; i < argc; i++ {
		if off+4 > len(buf) {
			return nil, fmt.Errorf("amp: short read on arg %d length", i)
		}
		l := binary.BigEndian.Uint32(buf[off:])
		off += 4
		if off+int(l) > len(buf) {
			return nil, fmt.Errorf("amp: short read on arg %d body", i)
		}
		args[i] = make([]byte, l)
		copy(args[i], buf[off:off+int(l)])
		off += int(l)
	}
	return args, nil
}

// packJSON prefixes a JSON payload with "j:" so it round-trips through
// amp-message's `unpack` (which detects the prefix and JSON-parses).
func packJSON(payload []byte) []byte {
	out := make([]byte, 0, len(payload)+2)
	out = append(out, 'j', ':')
	out = append(out, payload...)
	return out
}

// packString prefixes a UTF-8 string with "s:" so it round-trips through
// amp-message's `unpack` (which detects the prefix and returns the bytes).
func packString(s string) []byte {
	out := make([]byte, 0, len(s)+2)
	out = append(out, 's', ':')
	out = append(out, s...)
	return out
}

// unpackJSON strips the "j:" prefix and returns the JSON payload.
func unpackJSON(buf []byte) ([]byte, error) {
	if len(buf) < 2 || buf[0] != 'j' || buf[1] != ':' {
		return nil, fmt.Errorf("amp-msg: not a json-typed argument (got %q)", prefix(buf, 2))
	}
	return buf[2:], nil
}

func prefix(b []byte, n int) []byte {
	if len(b) < n {
		n = len(b)
	}
	return b[:n]
}

// transport is a thread-safe single connection to one PM2 daemon socket.
// Each request gets a synchronous reply; the daemon keeps request/reply
// order per connection, so a mutex serialises writes and reads.
type transport struct {
	conn net.Conn
	mu   sync.Mutex
}

func newTransport(socketPath string) (*transport, error) {
	c, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("pm2: dial %s: %w", socketPath, err)
	}
	return &transport{conn: c}, nil
}

// Close releases the underlying socket.
func (t *transport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.conn == nil {
		return nil
	}
	err := t.conn.Close()
	t.conn = nil
	return err
}

// send writes a fully-encoded AMP frame to the socket.
func (t *transport) send(frame []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.conn == nil {
		return errors.New("pm2: transport is closed")
	}
	_, err := t.conn.Write(frame)
	return err
}

// recv reads exactly one AMP frame from the socket. Uses io.ReadFull so
// partial reads across syscalls are reassembled correctly.
func (t *transport) recv() ([][]byte, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.conn == nil {
		return nil, errors.New("pm2: transport is closed")
	}

	var meta [1]byte
	if _, err := io.ReadFull(t.conn, meta[:]); err != nil {
		return nil, fmt.Errorf("pm2: read meta: %w", err)
	}
	argc := int(meta[0] & 0x0f)
	if argc > ampMaxArgs {
		return nil, fmt.Errorf("pm2: bad argc %d in reply meta %#x", argc, meta[0])
	}

	args := make([][]byte, argc)
	var lenBuf [4]byte
	for i := 0; i < argc; i++ {
		if _, err := io.ReadFull(t.conn, lenBuf[:]); err != nil {
			return nil, fmt.Errorf("pm2: read arg %d length: %w", i, err)
		}
		l := binary.BigEndian.Uint32(lenBuf[:])
		args[i] = make([]byte, l)
		if l == 0 {
			continue
		}
		if _, err := io.ReadFull(t.conn, args[i]); err != nil {
			return nil, fmt.Errorf("pm2: read arg %d body: %w", i, err)
		}
	}
	return args, nil
}