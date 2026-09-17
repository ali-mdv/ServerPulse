package pm2

import (
	"encoding/json"
	"io"
	"net"
	"sync"
	"testing"
)

// fakeDaemon is a minimal axon-rpc server used to exercise Client
// without needing a real PM2 daemon on the host.
type fakeDaemon struct {
	ln       net.Listener
	handler  func(method string, args []any) (any, error)
	mu       sync.Mutex
	received []callRequest
}

func startFakeDaemon(t *testing.T) *fakeDaemon {
	t.Helper()
	ln, err := net.Listen("unix", "\x00fake-pm2-rpc")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	d := &fakeDaemon{ln: ln}
	d.handler = func(method string, args []any) (any, error) {
		return map[string]any{"echo": method, "args": args}, nil
	}
	go d.serve()
	return d
}

func (d *fakeDaemon) serve() {
	for {
		c, err := d.ln.Accept()
		if err != nil {
			return
		}
		go d.handle(c)
	}
}

func (d *fakeDaemon) handle(c net.Conn) {
	defer func() {
		// Silently swallow panics so a test bug doesn't take the test
		// binary down before the assertion fires.
		_ = recover()
		c.Close()
	}()
	for {
		args, err := readFrame(c)
		if err != nil {
			return
		}
		// Real pm2-axon req/rep wire format: [json_body, id]
		if len(args) != 2 {
			return
		}
		body, err := unpackJSON(args[0])
		if err != nil {
			return
		}
		var req callRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return
		}
		d.mu.Lock()
		d.received = append(d.received, req)
		d.mu.Unlock()

		result, err := d.handler(req.Method, req.Args)
		var resp any
		if err != nil {
			if fe, ok := err.(*fakeError); ok {
				resp = map[string]string{"error": fe.msg}
			} else {
				resp = map[string]string{"error": err.Error()}
			}
		} else {
			resp = map[string]any{"args": []any{result}}
		}
		respBody, _ := json.Marshal(resp)
		// Reply mirrors pm2-axon's rep socket: [body, echoed_id].
		reply, _ := encodeAMP([][]byte{packJSON(respBody), args[1]})
		c.Write(reply)
	}
}

func (d *fakeDaemon) stop() {
	d.ln.Close()
}

func (d *fakeDaemon) path() string {
	return d.ln.Addr().String()
}

func readFrame(c net.Conn) ([][]byte, error) {
	var meta [1]byte
	if _, err := io.ReadFull(c, meta[:]); err != nil {
		return nil, err
	}
	argc := int(meta[0] & 0x0f)
	args := make([][]byte, argc)
	var lenBuf [4]byte
	for i := 0; i < argc; i++ {
		if _, err := io.ReadFull(c, lenBuf[:]); err != nil {
			return nil, err
		}
		l := uint32(lenBuf[0])<<24 | uint32(lenBuf[1])<<16 | uint32(lenBuf[2])<<8 | uint32(lenBuf[3])
		args[i] = make([]byte, l)
		if l == 0 {
			continue
		}
		if _, err := io.ReadFull(c, args[i]); err != nil {
			return nil, err
		}
	}
	return args, nil
}

func TestClientRoundTripsAgainstFakeDaemon(t *testing.T) {
	d := startFakeDaemon(t)
	defer d.stop()

	// getMonitorData replies with a JSON array of process objects.
	d.handler = func(method string, args []any) (any, error) {
		return []map[string]any{
			{"name": "fake-app", "pm_id": 0, "monit": map[string]any{"cpu": 1.5, "memory": 12345}},
			{"name": "another", "pm_id": 1, "monit": map[string]any{"cpu": 0.0, "memory": 0}},
		}, nil
	}

	c, err := Dial(d.path())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.Close()

	got, err := c.GetMonitorData()
	if err != nil {
		t.Fatalf("getMonitorData: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d processes, want 2", len(got))
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.received) != 1 {
		t.Fatalf("daemon saw %d calls, want 1", len(d.received))
	}
	if d.received[0].Type != "call" {
		t.Fatalf("req.type = %q, want call", d.received[0].Type)
	}
	if d.received[0].Method != "getMonitorData" {
		t.Fatalf("req.method = %q", d.received[0].Method)
	}
}

func TestClientSurfacesDaemonError(t *testing.T) {
	d := startFakeDaemon(t)
	defer d.stop()
	d.handler = func(method string, args []any) (any, error) {
		return nil, &fakeError{msg: "method \"" + method + "\" does not exist"}
	}

	c, err := Dial(d.path())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	_, err = c.GetMonitorData()
	if err == nil {
		t.Fatal("expected error from daemon")
	}
	if got := err.Error(); got != `pm2 rpc getMonitorData: method "getMonitorData" does not exist` {
		t.Fatalf("unexpected error: %s", got)
	}
}

type fakeError struct{ msg string }

func (e *fakeError) Error() string { return e.msg }