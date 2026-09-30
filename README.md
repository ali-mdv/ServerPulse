# ServerPulse

A monorepo for **ServerPulse** — a server-monitoring dashboard built from two merged repos:

- `backend/` — Go + Gin REST API that probes system metrics (CPU, memory, disk, network) and manages Docker containers and PM2 processes.
- `frontend/` — Vue 3 + Vite SPA that visualises servers, Docker containers, PM2 services, system usage, and per-server logs.

MongoDB persists users, authentication, and configuration.

---

## Stack

| Layer    | Tech                                                                    |
|----------|-------------------------------------------------------------------------|
| Backend  | Go 1.24+, Gin, MongoDB driver, Viper, JWT, gopsutil, Docker SDK, Air     |
| Frontend | Vue 3, vue-router, Pinia, VueUse, Vite, TypeScript, TailwindCSS, Naive UI, ECharts, Axios |
| Database | MongoDB 7                                                               |
| Infra    | Docker, docker-compose, nginx (reverse proxy + SPA host)                |

---

## Project layout

```
ServerPulse/
├── backend/                 # Go service
│   ├── cmd/main.go          # entrypoint
│   ├── internal/
│   │   ├── handlers/        # gin HTTP handlers
│   │   ├── services/        # business logic (system, docker, pm2, auth, user)
│   │   │   └── pm2/         # axon-rpc client over PM2's unix socket
│   │   │       ├── transport.go      # AMP wire format + unix-socket I/O
│   │   │       ├── rpc.go            # axon-rpc call envelope
│   │   │       ├── client.go         # high-level PM2 methods (list/start/stop/logs)
│   │   │       └── *_test.go         # wire-format + fake-daemon integration tests
│   │   ├── repository/      # mongo persistence
│   │   ├── routes/          # gin route registration
│   │   ├── models/          # domain types
│   │   ├── dtos/            # request/response shapes
│   │   └── middlewares/     # auth, etc.
│   ├── pkg/
│   │   ├── config/          # env-driven config (Viper)
│   │   ├── mongo/           # mongo client init
│   │   └── errors/          # app error wrapper
│   ├── .air.toml            # hot-reload config (local dev)
│   │   └── .air.docker.toml # hot-reload config (Docker dev, writes to /out)
│   ├── Dockerfile           # multi-stage Go build → distroless
│   ├── Dockerfile.dev       # dev image: toolchain + deps, code mounted at runtime
│   ├── go.mod
│   └── go.sum
│
├── frontend/                # Vue SPA
│   ├── src/
│   │   ├── api/             # typed API clients (servers, docker, pm2, system, auth, user, report)
│   │   ├── components/      # reusable UI (cards, modals, items, panels)
│   │   ├── pages/           # routed views (Dashboard, Servers, ServerDetails, Alerts, Settings, Profile, Login, AddUser, NotFound, ServerError)
│   │   ├── stores/          # Pinia stores (auth, root, system)
│   │   ├── plugins/         # axios setup
│   │   ├── hooks/           # use-mobile, use-toast
│   │   ├── types/           # shared TS types
│   │   └── global.css       # Tailwind entry
│   ├── Dockerfile           # multi-stage pnpm build → nginx
│   ├── Dockerfile.dev       # dev image: pnpm + node_modules baked, code mounted at runtime
│   ├── nginx.conf           # SPA fallback + /api reverse proxy to backend
│   ├── netlify.toml         # Netlify deploy config
│   ├── package.json
│   ├── pnpm-lock.yaml
│   ├── tsconfig.json
│   ├── vite.config.ts
│   ├── vite.config.server.ts
│   └── vite.config.docker.ts # vite config used only by the Docker dev container
│
├── docker-compose.yml        # mongo + backend + frontend orchestration
├── docker-compose.dev.yml    # dev overlay: read-only code mounts + hot reload
├── .env / .env.example      # build-time and runtime env
├── .dockerignore            # root build-context exclusions
├── .gitignore
├── .prettierrc
├── AGENTS.md                # project guidance for AI agents
└── README.md
```

---

## Quick start (Docker)

```bash
cp .env.example .env            # set PM2_HOME, DOCKER_GID, HOST_UID/GID first
docker compose up --build       # production: code changes require a rebuild
```

| URL                              | Service                |
|----------------------------------|------------------------|
| `http://localhost:${FRONTEND_PORT}` | Frontend (nginx → SPA) |
| http://localhost:12000           | Backend (Gin)          |
| mongodb://localhost:27017        | MongoDB                |

Frontend `/api/*` is reverse-proxied by nginx to the backend container. The backend talks to Mongo by container name (`mongo`), not localhost.

### Development (hot reload, read-only code mounts)

```bash
docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build
```

Same URLs as production, but with live reload:

- **Backend** — source bind-mounted read-only; `air` recompiles in-container and writes binaries to the `/out` volume. Code edits never rebuild the image.
- **Frontend** — source bind-mounted read-only; `vite` runs with `vite.config.docker.ts` (native TS config loader — the normal config bundler writes temp files next to the config, which is impossible on a read-only mount) and `--configLoader native`.
- The image is **only** rebuilt when dependencies change: `backend/go.mod` / `go.sum`, or `frontend/pnpm-lock.yaml`. A one-shot `perms-init` service chowns the backend volumes so `air` can write as `HOST_UID`.
- The dev overlay pins frontend ports with `!override` (compose otherwise appends and double-binds the host port).

Frontend `/api/*` is reverse-proxied by nginx to the backend container. The backend talks to Mongo by container name (`mongo`), not localhost.

---

## Host monitoring

The backend monitors **the host it's running on** (the Docker host), not just itself. Two host-side daemons need to be reachable from the container without anything being installed inside it:

- **Docker daemon** — reached via `/var/run/docker.sock`, bind-mounted into the backend container. The container also needs the host's `docker` group GID (see `DOCKER_GID` below) so the backend user can read the socket.
- **PM2 daemon** — reached via `$PM2_HOME/rpc.sock` on the host, bind-mounted into the container at `/run/pm2`. PM2's directory is bind-mounted under `/run/pm2` rather than `/root/.pm2` because distroless's `/root` is mode 700 and unreachable by a non-root uid.

The backend runs as `HOST_UID:HOST_GID` (uid 1000 / gid 1000 by default) so it can read the PM2 socket as its owner. Bind mounts inherit host permissions, so this matches whatever user runs PM2 on the host.

If you want the backend to monitor a *different* host later, the same `/run/pm2` mount becomes an HTTP bridge on that host instead — the in-container code stays the same.

---

## PM2 client (no shell-out, no Node in the container)

The backend never shells out to the `pm2` binary. The distroless image has no shell, no Node, no `pm2` — and adding them would defeat the purpose of a minimal monitoring image. Instead, `internal/services/pm2/` is a from-scratch Go implementation of PM2's wire protocol:

- **AMP** (`visionmedia/node-amp`) — version-byte + per-arg length-prefixed framing.
- **axon-rpc** (`@pm2/axon-rpc`) — JSON `{type,method,args}` envelopes wrapped as `j:`/`s:`-prefixed AMP args. Every call carries a request id (`<pid>:<n>`); the daemon's rep socket pops it off to wire up the reply callback.
- **Logs** — read directly from `pm_out_log_path` / `pm_err_log_path` instead of going through `pm2 logs`, so the daemon doesn't have to be involved for tail.

Methods exposed to the rest of the backend match the previous interface (`List`, `Start`, `Stop`, `Restart`, `TailLogs`), so handlers/routes don't need to know the protocol changed.

If PM2's daemon isn't running, the dial fails fast at startup and the cause is logged. Polling endpoints (`pm2/services`, `docker/containers`) respond with `HTTP 200` and `{"available": false, ...empty list}` instead of failing — the frontend hides the section when a provider isn't reachable. User-triggered actions (start/stop/restart/logs) still return `503 pm2 daemon unreachable: <reason>`.

---

## Remote agent control channel

The Go agent opens a Socket.IO connection to `/api/v1/agents/socket.io` (auth: `AgentToken` in the handshake `auth` payload) and executes `start`/`stop`/`restart`/`logs` for Docker and PM2 on the remote host. The backend's `AgentHub` emits `agent:command` and waits for the agent's ack (synchronous request/response). Timeouts: 10s for actions, 20s for logs; errors map to 503 (not connected), 504 (timeout), 502 (transport), or the agent's own message (mapped to 404/400/503/500 by `resultToError`). The Python agent implements the same channel when `python-socketio` is installed.

---

## Local development

### Backend

```bash
cd backend
cp ../.env.example .env       # backend reads GIN_MODE, PORT, DB_HOST, DB_PORT, DB_USERNAME, DB_PASSWORD, PM2_SOCKET_PATH
# point DB_HOST at a local mongod (e.g. 127.0.0.1:27017) — the docker stack uses the "mongo" service name instead
# set PM2_SOCKET_PATH to the host's $HOME/.pm2/rpc.sock if you want to talk to a local PM2 daemon

air                            # hot reload (uses .air.toml)
# or
go run ./cmd                   # one-shot run
go test ./...                  # unit tests (incl. axon-rpc round-trip + fake daemon)
```

Requires Go 1.24+. MongoDB must be reachable.

### Frontend

```bash
cd frontend
pnpm install
pnpm dev                       # vite dev server (HMR)
pnpm typecheck
pnpm test
pnpm build                     # production bundle into dist/
```

The frontend reads Vite build-time vars (see `.env.example` at root):

- `VITE_SERVER_ADDRESS` — leave empty for same-origin requests through `/api/*` (nginx in prod, the vite dev proxy in dev)
- `VITE_API_BASE_URL` — usually `/api/v1` so the proxy can forward
- `VITE_API_TIME_OUT` — request timeout in ms
- `VITE_ENVIRONMENT` — `development` | `production`
- `VITE_APP_NAME`, `VITE_APP_VERSION`

Both the nginx image and the vite dev server proxy `/api/*` to the backend, so the app works from any device on the network without a hardcoded `localhost`. For `pnpm dev` on the host, the proxy targets `http://localhost:12000` (override with `VITE_PROXY_TARGET`); in the Docker dev overlay it targets `http://backend:12000`.

---

## API surface (v1)

All routes are mounted under `/api/v1`.

| Route prefix       | Purpose                                  |
|--------------------|------------------------------------------|
| `/api/v1/auth`     | login                                    |
| `/api/v1/users`    | user CRUD                                |
| `/api/v1/servers`  | server/agent CRUD + API-key generation   |
| `/api/v1/servers/:id/docker/...` | per-server Docker start/stop/restart/logs |
| `/api/v1/servers/:id/pm2/...`    | per-server PM2 start/stop/restart/logs    |
| `/api/v1/docker`   | list containers, start/stop/restart/logs |
| `/api/v1/pm2`      | list processes, start/stop/restart/logs  |
| `/api/v1/system`   | per-host CPU / memory / disk / network   |
| `/api/v1/report`   | aggregated reports                       |
| `/api/v1/agents/socket.io` | agent control channel (Socket.IO) |

> The frontend's server list/details pages use the real `/api/v1/servers` endpoints.
> `/servers/:id/docker/*` and `/servers/:id/pm2/*` accept the local server's hex id
> (or `local`) and short-circuit to the in-process services; any other id is
> dispatched to that server's agent over Socket.IO (`AgentHub.Execute`).

Authentication: JWT bearer in `Authorization: Bearer <token>` header on protected routes.

---

## Ports & env

| Variable            | Where used       | Default                  |
|---------------------|------------------|--------------------------|
| `BACKEND_PORT`      | docker-compose   | `12000`                  |
| `FRONTEND_PORT`     | docker-compose   | `8080`                   |
| `PORT`              | backend runtime  | `12000`                  |
| `GIN_MODE`          | backend runtime  | `release` (docker), `debug` (local) |
| `DB_HOST`           | backend runtime  | `mongo` in docker, `127.0.0.1` locally |
| `DB_PORT`           | backend runtime  | `27017`                  |
| `PM2_HOME`          | host bind mount  | `/home/<user>/.pm2` (find yours with `ls -ld ~/.pm2`) |
| `PM2_SOCKET_PATH`   | backend runtime  | `/run/pm2/rpc.sock` (matches the bind mount target) |
| `HOST_UID`          | docker-compose   | `1000` — uid running the host's PM2 (`id -u`) |
| `HOST_GID`          | docker-compose   | `1000` — primary gid of that user (`id -g`) |
| `DOCKER_GID`        | docker-compose   | find with `stat -c %g /var/run/docker.sock` (commonly `999` or `125`) |
| `VITE_API_BASE_URL` | frontend build   | `/api`                   |
| `VITE_SERVER_ADDRESS` | frontend build | `http://localhost:8080`  |

---

## Notes

- `nginx.conf` in the frontend folder is **only** used by the Docker image — local `pnpm dev` does not use it.
- `backend/.air.toml` is used by local `air` runs; `backend/.air.docker.toml` is used by the Docker dev overlay (same watcher, but the binary goes to `/out` because the source tree is mounted read-only).
- `frontend/vite.config.docker.ts` is used only by the Docker dev container (read-only source); `pnpm dev` uses `vite.config.ts`.
- MongoDB credentials live in `.env`; the bundled example uses `ali` / `123` for local convenience — change before exposing publicly.
- Polling endpoints that depend on host daemons return `200` with `available: false` instead of `503` so the browser console stays quiet; action endpoints still return real error codes.