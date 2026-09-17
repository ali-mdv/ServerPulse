package pm2

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync/atomic"
)

// callRequest is the JSON envelope sent over the wire to the daemon.
// Mirrors axon-rpc Client.call (visionmedia/axon-rpc).
type callRequest struct {
	Type   string `json:"type"`
	Method string `json:"method"`
	Args   []any  `json:"args"`
}

// reply is the JSON envelope returned by the daemon for a single call.
// On error it has Error (and optionally Stack); on success it has Args.
type reply struct {
	Args  []json.RawMessage `json:"args"`
	Error string            `json:"error"`
	Stack string            `json:"stack"`
}

// rpcClient is a thin Go counterpart of @pm2/axon-rpc's Client.
// Each request is synchronous over one connection.
type rpcClient struct {
	tr    *transport
	idCtr atomic.Uint32
}

func newRPCClient(tr *transport) *rpcClient {
	return &rpcClient{tr: tr}
}

// nextID generates the request id that pm2-axon's req socket would
// normally append to every send. The daemon's rep socket pops the last
// arg off and treats it as the id; without it the server's reply fn is
// never wired up and the daemon logs "reply false".
//
// pm2-axon formats ids as "<process_pid>:<counter>". We mirror that
// shape so future log inspection matches what a Node.js client would do.
func (c *rpcClient) nextID() string {
	n := c.idCtr.Add(1)
	return strconv.Itoa(os.Getpid()) + ":" + strconv.FormatUint(uint64(n), 10)
}

// call invokes `method` on the daemon. On success, returns the first
// element of the daemon's reply Args (the actual return value), which is
// what most PM2 methods do (`cb(null, result)` -> reply.args = [result]).
//
// Returns the raw JSON payload so callers can unmarshal into the shape
// they expect (e.g. []PM2Process for getMonitorData).
func (c *rpcClient) call(method string, args []any) (json.RawMessage, error) {
	if args == nil {
		args = []any{}
	}
	req := callRequest{Type: "call", Method: method, Args: args}
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("pm2 rpc: marshal %s: %w", method, err)
	}

	// Wire format mirrors axon-rpc + pm2-axon's req socket:
	//   AMP frame = [j:<json_call_object>, s:<request_id>]
	// The rep socket pops the id off and emits ('message', obj, reply)
	// to axon-rpc's Server.onmessage.
	frame, err := encodeAMP([][]byte{packJSON(payload), packString(c.nextID())})
	if err != nil {
		return nil, fmt.Errorf("pm2 rpc: encode %s: %w", method, err)
	}
	if err := c.tr.send(frame); err != nil {
		return nil, fmt.Errorf("pm2 rpc: send %s: %w", method, err)
	}

	rawArgs, err := c.tr.recv()
	if err != nil {
		return nil, fmt.Errorf("pm2 rpc: recv %s: %w", method, err)
	}
	if len(rawArgs) != 2 {
		return nil, fmt.Errorf("pm2 rpc: %s: expected 2 reply args (body, id), got %d", method, len(rawArgs))
	}

	body, err := unpackJSON(rawArgs[0])
	if err != nil {
		return nil, fmt.Errorf("pm2 rpc: %s: %w", method, err)
	}

	var r reply
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("pm2 rpc: %s: bad reply json: %w", method, err)
	}

	if r.Error != "" {
		return nil, fmt.Errorf("pm2 rpc %s: %s", method, r.Error)
	}
	if len(r.Args) == 0 {
		return nil, nil
	}
	return r.Args[0], nil
}