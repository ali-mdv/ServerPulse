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
│   ├── Dockerfile           # multi-stage Go build → distroless
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
│   ├── nginx.conf           # SPA fallback + /api reverse proxy to backend
│   ├── netlify.toml         # Netlify deploy config
│   ├── package.json
│   ├── pnpm-lock.yaml
│   ├── tsconfig.json
│   ├── vite.config.ts
│   └── vite.config.server.ts
│
├── docker-compose.yml       # mongo + backend + frontend orchestration
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
cp .env.example .env
docker compose up --build
```

| URL                              | Service                |
|----------------------------------|------------------------|
| http://localhost:8080            | Frontend (nginx → SPA) |
| http://localhost:12000           | Backend (Gin)          |
| mongodb://localhost:27017        | MongoDB                |

Frontend `/api/*` is reverse-proxied by nginx to the backend container. The backend talks to Mongo by container name (`mongo`), not localhost.

---

## Local development

### Backend

```bash
cd backend
cp ../.env.example .env       # backend reads GIN_MODE, PORT, DB_HOST, DB_PORT, DB_USERNAME, DB_PASSWORD
# point DB_HOST at a local mongod (e.g. 127.0.0.1:27017) — the docker stack uses the "mongo" service name instead

air                            # hot reload (uses .air.toml)
# or
go run ./cmd                   # one-shot run
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

- `VITE_SERVER_ADDRESS` — base URL the SPA expects itself to be served at
- `VITE_API_BASE_URL` — usually `/api` so nginx can proxy
- `VITE_API_TIME_OUT` — request timeout in ms
- `VITE_ENVIRONMENT` — `development` | `production`
- `VITE_APP_NAME`, `VITE_APP_VERSION`

When building outside Docker, set `VITE_API_BASE_URL=http://localhost:12000` so the dev server can talk to a locally-running backend.

---

## API surface (v1)

All routes are mounted under `/api/v1`.

| Route prefix       | Purpose                                  |
|--------------------|------------------------------------------|
| `/api/v1/auth`     | login, register, refresh                 |
| `/api/v1/users`    | user CRUD                                |
| `/api/v1/servers`  | monitored server list, fetch by id       |
| `/api/v1/docker`   | list containers, start/stop/restart/logs |
| `/api/v1/pm2`      | list processes, start/stop/restart/logs  |
| `/api/v1/system`   | per-host CPU / memory / disk / network   |
| `/api/v1/report`   | aggregated reports                       |

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
| `VITE_API_BASE_URL` | frontend build   | `/api`                   |
| `VITE_SERVER_ADDRESS` | frontend build | `http://localhost:8080`  |

---

## Notes

- `nginx.conf` in the frontend folder is **only** used by the Docker image — local `pnpm dev` does not use it.
- `backend/.air.toml` is **only** used by local `air` runs.
- MongoDB credentials live in `.env`; the bundled example uses `ali` / `123` for local convenience — change before exposing publicly.
