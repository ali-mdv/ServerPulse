# # ServerPulseServer(Backend) — Agent guidance

- **Module path:** `server-monitoring` (see `backend/go.mod`).
- **Entry:** `backend/cmd/main.go`. Loads config, initialises Mongo, then `routes.Setup()`.
- **HTTP routing:** gin. Each domain gets its own `internal/routes/v1/<name>.go` and is wired in `internal/routes/v1/routes.go` (or `router.go`). Keep new domains in this layered shape: `handlers` → `services` → `repository`.
- **DTOs vs models:** `internal/dtos` is for request/response shapes; `internal/models` is the persisted entity. Don't return models directly from handlers — wrap in DTOs.
- **Errors:** `pkg/errors` exposes `AppError` with `New()` + `Wrap()` (+ `IsUnavailable()`). Use it; don't `errors.New` in handlers.
- **Poll-endpoint degradation:** list endpoints polled by the dashboard (`pm2/services`, `docker/containers`) must respond `200 {"…": [], "available": false}` when their host daemon is unreachable (`errors.IsUnavailable` / `isDockerUnavailable` in the handler) — never a 503 there, it would spam the browser console every interval. Only user-initiated actions (start/stop/restart/logs) surface real error codes.
- **Config:** `pkg/config.Load()` reads env via Viper. Don't read `os.Getenv` directly elsewhere.
- **Auth:** JWT (`pkg/...`). Protected routes go through the auth middleware in `internal/middlewares/auth.go`.
- **System metrics:** `gopsutil` is the source. Don't shell out to `top`/`free`/etc.
- **Docker control:** use the official `github.com/docker/docker/client` SDK (`client.FromEnv`). The backend container needs `/var/run/docker.sock` bind-mounted and the host's `docker` group added via `group_add` in `docker-compose.yml`.
- **PM2 control:** never shell out to the `pm2` binary — the distroless image has no Node. `internal/services/pm2/` is a from-scratch Go client that speaks PM2's AMP + axon-rpc protocol directly over the daemon's unix socket (see "PM2 wire protocol" below).
- **Agent control channel:** remote start/stop/restart/logs go through Socket.IO (`internal/services/agent_hub.go` + `agent_socket.go`) mounted at `/api/v1/agents/socket.io`. `CommandService` (`command_service.go`) branches local vs remote by `serverID`; routes live under `/servers/:serverId/...` (`routes/v1/server_control.go`). The hub blocks on the agent's ack (10s actions / 20s logs) and maps failures to 503/504/502 via `resultToError`.
- **Mongo:** `pkg/mongo` owns the client. Get collections via `database.GetCollection(name)`.

### PM2 wire protocol (internal/services/pm2/)

The package owns three concerns, in three files:

- `transport.go` — AMP wire format (meta byte = `(version<<4)|argc`, then per-arg `uint32 BE length` + body) and a thread-safe single-connection reader/writer over the unix socket.
- `rpc.go` — axon-rpc envelope. Every call sends `[j:<json call object>, s:"<pid>:<n>"]`; the daemon's rep socket pops the id off to wire up the reply. Replies come back as `[j:<json body>, s:<echoed id>]` and we decode the first arg.
- `client.go` — high-level methods (`GetMonitorData`, `StartProcessID`, `StopProcessID`, `RestartProcessID`, `TailLogs`). Logs are read directly from `pm_out_log_path` / `pm_err_log_path`, no RPC involved.

If you change anything here, mirror it in the fake server in `client_test.go` (`fakeDaemon`) so the round-trip tests still cover it. The `transient` cases to watch for: missing id → daemon logs `reply false` and never responds; wrong prefix (`j:` vs `s:`) → unpack fails; argc > 15 → AMP refuses to encode.

If PM2's daemon is down, `Dial` fails and the service keeps the error to surface later as a 503 (not 500). Don't cache a successful client at construction time without also storing the dial error — otherwise the error becomes invisible after restart.

## Conventions

- **One thing per file at the top level of a domain:** one handler per file, one service per file, one route registration per file.
- **No business logic in handlers:** handlers parse, call service, return. Services do the work and call repositories.
- **No direct Mongo in handlers/services:** always go through `repository/`.
- **Naming:** backend uses snake_case filenames (`user_repository.go`), exported symbols PascalCase.
- **Provider availability:** stores expose an `available` flag next to the payload (from `pm2/services` / `docker/containers`); components hide a provider's section when it's `false` instead of rendering placeholders or toasting every interval.

## What NOT to do

- Don't bypass the layered structure (handler → service → repository) in the backend.
- Don't reinstall Node or the `pm2` binary inside the backend image — the whole point of `internal/services/pm2/` is to talk to the host's PM2 without installing anything in the container. If you need a new PM2 operation, extend `internal/services/pm2/client.go` using the existing transport.
- Don't shell out from the agent control channel — `agent/control.go` must use the Docker SDK and the PM2 axon-rpc client, same as the backend.
