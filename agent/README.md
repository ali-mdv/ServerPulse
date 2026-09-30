# ServerPulse Agent

Lightweight remote agents that collect system metrics and push them to a ServerPulse backend.

## Workflow

1. Create a server/agent record in the backend:
   ```bash
   curl -X POST http://localhost:8080/api/v1/servers \
     -H "Authorization: Bearer <user-jwt>" \
     -H "Content-Type: application/json" \
     -d '{"name": "web-01", "description": "frontend host"}'
   ```

2. Generate an API key for that server:
   ```bash
   curl -X POST http://localhost:8080/api/v1/servers/<server-id>/api-key \
     -H "Authorization: Bearer <user-jwt>"
   ```

3. Copy the revealed `apiKey` and run one of the agents below.

## Environment variables

Copy `.env.example` to `.env` and fill in the values.

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `SERVER_PULSE_URL` | yes | — | Backend base URL, e.g. `http://localhost:8080` |
| `SERVER_PULSE_API_KEY` | yes | — | API key revealed in step 2 |
| `SERVER_PULSE_INTERVAL` | no | `30` | Push interval in seconds |
| `SERVER_PULSE_NAME` | no | hostname | Agent/server name reported to the backend |
| `SERVER_PULSE_HOST` | no | hostname | Host address reported to the backend |
| `SERVER_PULSE_PORT` | no | `0` | Port reported to the backend (`0` means "unknown") |
| `SERVER_PULSE_DESCRIPTION` | no | — | Optional description reported to the backend |
| `PM2_SOCKET_PATH` | no | `~/.pm2/rpc.sock` | Path to PM2's RPC unix socket |

## Python agent

Install dependencies:

```bash
pip install psutil requests
# Optional, for Docker container reporting:
pip install docker
# Optional, for the Socket.IO control channel (remote start/stop/restart/logs):
pip install 'python-socketio[client]'
```

Run:

```bash
source .env
python agent.py
```

When `python-socketio` is installed, the agent also opens a control channel
to `/api/v1/agents/socket.io` (auth: `SERVER_PULSE_API_KEY`) and executes
start/stop/restart/logs for Docker containers and PM2 processes on demand.

## Go agent

Build:

```bash
cd agent
go mod tidy
go build -o serverpulse-agent .
```

Run:

```bash
source .env
./serverpulse-agent
```

The Go agent also opens a Socket.IO control channel to `/api/v1/agents/socket.io`
(auth: `SERVER_PULSE_API_KEY` in the handshake `auth` payload) and executes
start/stop/restart/logs for Docker containers and PM2 processes on demand.

## Collected metrics

Both agents report:

- CPU usage (%)
- Memory usage (used, total, percent)
- Disk usage (used, total, percent)
- Network I/O (bytes sent/received per second)
- PM2 processes (when `PM2_SOCKET_PATH` is reachable)

The Python agent also reports Docker containers when the `docker` package is installed and the daemon is reachable.
