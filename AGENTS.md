# ServerPulse — Agent guidance

ServerPulse is a server-monitoring monorepo. Backend (Go + Gin + MongoDB) and frontend (Vue 3 + Vite) live side-by-side. Read this before making structural changes.

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

Base config files (`.gitignore`, `.dockerignore`, `.env.example`, `.prettierrc`, `README.md`, `AGENTS.md`) live at the **root**. Do not duplicate them inside `backend/` or `frontend/`.

Subproject files that must stay where they are (referenced by tools/builds):
- `backend/.air.toml`, `backend/.air.docker.toml` — Go hot reload (local / Docker dev)
- `backend/Dockerfile`, `backend/Dockerfile.dev` — prod build and dev-overlay image (`Dockerfile.dev` bakes the toolchain + `go mod download` only; source is mounted at runtime)
- `frontend/Dockerfile`, `frontend/Dockerfile.dev`, `frontend/nginx.conf` — `docker compose` and the image builds
- `frontend/vite.config.ts`, `frontend/vite.config.docker.ts` — the docker variant is used only by the dev container (native TS config loader; the bundler can't run on a read-only mount)
- `frontend/netlify.toml` — Netlify deploy
- `frontend/.npmrc` — pnpm config

## Backend

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
- **Mongo:** `pkg/mongo` owns the client. Get collections via `database.GetCollection(name)`.

### PM2 wire protocol (internal/services/pm2/)

The package owns three concerns, in three files:

- `transport.go` — AMP wire format (meta byte = `(version<<4)|argc`, then per-arg `uint32 BE length` + body) and a thread-safe single-connection reader/writer over the unix socket.
- `rpc.go` — axon-rpc envelope. Every call sends `[j:<json call object>, s:"<pid>:<n>"]`; the daemon's rep socket pops the id off to wire up the reply. Replies come back as `[j:<json body>, s:<echoed id>]` and we decode the first arg.
- `client.go` — high-level methods (`GetMonitorData`, `StartProcessID`, `StopProcessID`, `RestartProcessID`, `TailLogs`). Logs are read directly from `pm_out_log_path` / `pm_err_log_path`, no RPC involved.

If you change anything here, mirror it in the fake server in `client_test.go` (`fakeDaemon`) so the round-trip tests still cover it. The `transient` cases to watch for: missing id → daemon logs `reply false` and never responds; wrong prefix (`j:` vs `s:`) → unpack fails; argc > 15 → AMP refuses to encode.

If PM2's daemon is down, `Dial` fails and the service keeps the error to surface later as a 503 (not 500). Don't cache a successful client at construction time without also storing the dial error — otherwise the error becomes invisible after restart.

## Frontend

- **Stack:** Vue 3 (`<script setup>`), vue-router 4, Pinia, VueUse, Axios, Naive UI, ECharts (via `vue-echarts`), TailwindCSS 3, TypeScript, Vite 7. **Not React** — ignore any old "client/pages" / Radix / Lucide-React guidance that may be lingering in comments.
- **Pages** live in `frontend/src/pages/`. Components in `frontend/src/components/`. Pinia stores in `frontend/src/stores/`. Typed API clients in `frontend/src/api/`.
- **Routing:** add a new page by dropping a `.vue` in `pages/` and registering it in `frontend/src/router/index.ts`.
- **API calls:** import from `@/api/*`, not raw axios. The axios instance in `plugins/axios.ts` handles auth header injection.
- **State:** Pinia only. Don't introduce another state library.
- **Forms:** `vee-validate` + `yup` (see `pages/Profile.vue`, `pages/AddUser.vue` for the pattern).
- **Charts:** ECharts via `vue-echarts`. Don't pull in chart.js.
- **Styling:** Tailwind utility classes + Naive UI components. The `cn()` helper from `lib/utils.ts` is the way to compose class strings.
- **Tests:** Vitest. Co-locate as `*.spec.ts`.
- **TypeScript:** strict mode (`tsconfig.json`). Don't loosen `strict` or `noImplicitAny`.

## Build & run

| Task                          | Command                                          |
|-------------------------------|--------------------------------------------------|
| Run everything                | `docker compose up --build` (from repo root)     |
| Run in dev (hot reload)       | `docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build` (code mounted read-only; rebuild only when `backend/go.mod`/`go.sum` or `frontend/pnpm-lock.yaml` change) |
| Backend tests                 | `cd backend && go test ./...`                    |
| Backend only (local)          | `cd backend && air` (requires local mongod)      |
| Frontend only (local)         | `cd frontend && pnpm dev`                        |
| Typecheck frontend            | `cd frontend && pnpm typecheck`                  |
| Test frontend                 | `cd frontend && pnpm test`                       |
| Build frontend                | `cd frontend && pnpm build`                      |
| Lint                          | no project-level linter; prettier via root `.prettierrc` |

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

If these are wrong or unset, `/pm2/services` and `/docker/containers` return `200 {"available": false}` with empty lists (the dashboard hides the sections), but `pkg/config` only complains about *missing* `DB_*` / `PORT`, not about `PM2_SOCKET_PATH`, so check the container's startup log for `pm2 service: dial …` to see the real cause. User-initiated actions return the unmasked 503.

## Conventions

- **One thing per file at the top level of a domain:** one handler per file, one service per file, one route registration per file.
- **No business logic in handlers:** handlers parse, call service, return. Services do the work and call repositories.
- **No direct Mongo in handlers/services:** always go through `repository/`.
- **Vue components:** `<script setup lang="ts">`. Props via `defineProps`, emits via `defineEmits`, no Options API.
- **Naming:** backend uses snake_case filenames (`user_repository.go`), exported symbols PascalCase. Frontend uses PascalCase components, camelCase functions.
- **Imports:** prefer `@/` alias in frontend over deep relative paths.
- **Sample data:** `api/servers.ts` serves bundled sample data without network calls — there is no `/servers` backend endpoint; adding one means wiring it through `@/api/*` and removing the sample fallback.
- **Provider availability:** stores expose an `available` flag next to the payload (from `pm2/services` / `docker/containers`); components hide a provider's section when it's `false` instead of rendering placeholders or toasting every interval.

## What NOT to do

- Don't split `backend/.gitignore` or any other base config back into the subprojects — keep them at the root.
- Don't add another Node dependency manager. The project uses pnpm.
- Don't replace Naive UI or ECharts without checking the consuming pages first.
- Don't bypass the layered structure (handler → service → repository) in the backend.
- Don't introduce a top-level `client/` or `server/` directory — the repos were merged into `frontend/` and `backend/` deliberately.
- Don't reinstall Node or the `pm2` binary inside the backend image — the whole point of `internal/services/pm2/` is to talk to the host's PM2 without installing anything in the container. If you need a new PM2 operation, extend `internal/services/pm2/client.go` using the existing transport.
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