# ServerPulse — Agent guidance

ServerPulse is a server-monitoring monorepo. Backend (Go + Gin + MongoDB) and frontend (Vue 3 + Vite) live side-by-side. Read this before making structural changes.
Frontend use pnpm as package manager.

## Top-level layout

```
ServerPulse/
├── backend/        Go service. Entry: backend/cmd/main.go. Config: backend/pkg/config (Viper).
├── frontend/       Vue SPA. Entry: frontend/src/main.ts. Routes: frontend/src/router/index.ts.
├── docker-compose.yml
├── docker-compose.dev.yml
├── .env / .env.example
├── .gitignore  .dockerignore  .prettierrc
├── AGENTS.md (this file)
└── README.md
```

Subproject files that must stay where they are (referenced by tools/builds):

- `backend/.air.toml`, `backend/.air.docker.toml` — Go hot reload (local / Docker dev)
- `backend/Dockerfile`, `backend/Dockerfile.dev` — prod build and dev-overlay image (`Dockerfile.dev` bakes the toolchain + `go mod download` only; source is mounted at runtime)
- `frontend/Dockerfile`, `frontend/Dockerfile.dev`, `frontend/nginx.conf` — `docker compose` and the image builds
- `frontend/vite.config.ts`, `frontend/vite.config.docker.ts` — the docker variant is used only by the dev container (native TS config loader; the bundler can't run on a read-only mount)
- `frontend/netlify.toml` — Netlify deploy
- `frontend/.npmrc` — pnpm config

## Build & run

| Task                    | Command                                                                                                                                                                               |
| ----------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Run everything          | `docker compose up --build` (from repo root)                                                                                                                                          |
| Run in dev (hot reload) | `docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build` (code mounted read-only; rebuild only when `backend/go.mod`/`go.sum` or `frontend/pnpm-lock.yaml` change) |
| Backend tests           | `cd backend && go test ./...`                                                                                                                                                         |
| Backend only (local)    | `cd backend && air` (requires local mongod)                                                                                                                                           |
| Frontend only (local)   | `cd frontend && pnpm dev`                                                                                                                                                             |
| Typecheck frontend      | `cd frontend && pnpm typecheck`                                                                                                                                                       |
| Test frontend           | `cd frontend && pnpm test`                                                                                                                                                            |
| Build frontend          | `cd frontend && pnpm build`                                                                                                                                                           |
| Lint                    | no project-level linter; prettier via root `.prettierrc`                                                                                                                              |

## Environment

- `.env` at the root is the single source for compose and the build.
- `VITE_*` vars are baked at build time — changing `.env` requires `docker compose up --build` (or `pnpm build`) to take effect.
- Backend env (`GIN_MODE`, `PORT`, `DB_*`, `PM2_SOCKET_PATH`) is read at container start; changes require `docker compose up <service>`.

### Host-monitoring env (required for `/docker/*` and `/pm2/*`)

The backend container must reach two host-side daemons. Both are configured via `.env`:

- `PM2_HOME` — directory on the Docker host holding PM2's `rpc.sock`. Bind-mounted to `/run/pm2` in the container. **Do not** mount under `/root/.pm2` — distroless's `/root` is mode 700 and unreachable by the non-root backend user.
- `PM2_SOCKET_PATH` — path inside the container, almost always `/run/pm2/rpc.sock`. Read by `pkg/config` and passed to `services.NewPM2Service`.
- `HOST_UID`, `HOST_GID` — uid/gid of the user that owns PM2 on the host (`id -u` / `id -g`). The backend container runs as this user so it can read the PM2 socket.
- `DOCKER_GID` — gid of the `docker` group on the host (`stat -c %g /var/run/docker.sock`). Required for the backend to reach the docker socket.

If these are wrong or unset, `/pm2/services` and `/docker/containers` return `200 {"available": false}` with empty lists (the dashboard hides the sections), but `pkg/config` only complains about _missing_ `DB_*` / `PORT`, not about `PM2_SOCKET_PATH`, so check the container's startup log for `pm2 service: dial …` to see the real cause. User-initiated actions return the unmasked 503.

### Agent control channel (`/servers/:id/...`)

The Go agent opens a Socket.IO connection to `/api/v1/agents/socket.io` (auth: `AgentToken` in the handshake `auth` payload) and executes start/stop/restart/logs for Docker and PM2 on the remote host. The backend's `AgentHub` emits `agent:command` and waits for the agent's ack (synchronous request/response). Timeouts: 10s for actions, 20s for logs; errors map to 503 (not connected), 504 (timeout), 502 (transport), or the agent's own message (mapped to 404/400/503/500 by `resultToError`). The Python agent implements the same channel when `python-socketio` is installed.

## What NOT to do

- Don't introduce a top-level `client/` or `server/` directory — the repos were merged into `frontend/` and `backend/` deliberately.
- Don't mount PM2's directory at `/root/.pm2` — `/root` is mode 700 in distroless and the bind mount will be unreachable.

## Docker

- Compose file: `docker-compose.yml` at root; dev overlay `docker-compose.dev.yml` is merged on top for hot reload.
- Builds use `context: .` with relative dockerfile paths (`./backend/Dockerfile`, `./frontend/Dockerfile`). The root `.dockerignore` is honored for both.
- Backend image: multi-stage Go → distroless static, nonroot user **overridden** at runtime by `user:` in compose so it can read the host's PM2 socket.
- Frontend image: multi-stage pnpm build → nginx, with `/api` reverse-proxied to the `backend` service.
- Dev overlay (`-f docker-compose.dev.yml`): code bind-mounted read-only, `air`/vite hot reload, `perms-init` one-shot chowns the shared volumes for `HOST_UID`, and frontend `ports:` uses `!override` because compose merges port lists by append (double-binding the same host port otherwise).
- Mongo data persists in the named volume `mongo_data`.
- Bind mounts that cross the host/container boundary:
  - `/var/run/docker.sock:/var/run/docker.sock` (host docker daemon)
  - `${PM2_HOME}:/run/pm2` (host PM2 daemon sockets)
