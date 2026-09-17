# ServerPulse — Agent guidance

ServerPulse is a server-monitoring monorepo. Backend (Go + Gin + MongoDB) and frontend (Vue 3 + Vite) live side-by-side. Read this before making structural changes.

## Top-level layout

```
ServerPulse/
├── backend/        Go service. Entry: backend/cmd/main.go. Config: backend/pkg/config (Viper).
├── frontend/       Vue SPA. Entry: frontend/src/main.ts. Routes: frontend/src/router/index.ts.
├── docker-compose.yml
├── .env / .env.example
├── .gitignore  .dockerignore  .prettierrc
├── AGENTS.md (this file)
└── README.md
```

Base config files (`.gitignore`, `.dockerignore`, `.env.example`, `.prettierrc`, `README.md`, `AGENTS.md`) live at the **root**. Do not duplicate them inside `backend/` or `frontend/`.

Subproject files that must stay where they are (referenced by tools/builds):
- `backend/.air.toml` — Go hot reload
- `backend/Dockerfile` — `docker compose` builds from `./backend/Dockerfile`
- `frontend/Dockerfile`, `frontend/nginx.conf` — `docker compose` and the image build
- `frontend/netlify.toml` — Netlify deploy
- `frontend/.npmrc` — pnpm config

## Backend

- **Module path:** `server-monitoring` (see `backend/go.mod`).
- **Entry:** `backend/cmd/main.go`. Loads config, initialises Mongo, then `routes.Setup()`.
- **HTTP routing:** gin. Each domain gets its own `internal/routes/v1/<name>.go` and is wired in `internal/routes/v1/routes.go` (or `router.go`). Keep new domains in this layered shape: `handlers` → `services` → `repository`.
- **DTOs vs models:** `internal/dtos` is for request/response shapes; `internal/models` is the persisted entity. Don't return models directly from handlers — wrap in DTOs.
- **Errors:** `pkg/errors` exposes `AppError` with `New()` + `Wrap()`. Use it; don't `errors.New` in handlers.
- **Config:** `pkg/config.Load()` reads env via Viper. Don't read `os.Getenv` directly elsewhere.
- **Auth:** JWT (`pkg/...`). Protected routes go through the auth middleware in `internal/middlewares/auth.go`.
- **System metrics:** `gopsutil` is the source. Don't shell out to `top`/`free`/etc.
- **Docker / PM2 control:** use the official SDKs, not CLI shelling.
- **Mongo:** `pkg/mongo` owns the client. Get collections via `database.GetCollection(name)`.

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
| Backend only (local)          | `cd backend && air` (requires local mongod)      |
| Frontend only (local)         | `cd frontend && pnpm dev`                        |
| Typecheck frontend            | `cd frontend && pnpm typecheck`                  |
| Test frontend                 | `cd frontend && pnpm test`                       |
| Build frontend                | `cd frontend && pnpm build`                      |
| Lint                          | no project-level linter; prettier via root `.prettierrc` |

## Environment

- `.env` at the root is the single source for compose and the build.
- `VITE_*` vars are baked at build time — changing `.env` requires `docker compose up --build` (or `pnpm build`) to take effect.
- Backend env (`GIN_MODE`, `PORT`, `DB_*`) is read at container start; changes require `docker compose up <service>`.

## Conventions

- **One thing per file at the top level of a domain:** one handler per file, one service per file, one route registration per file.
- **No business logic in handlers:** handlers parse, call service, return. Services do the work and call repositories.
- **No direct Mongo in handlers/services:** always go through `repository/`.
- **Vue components:** `<script setup lang="ts">`. Props via `defineProps`, emits via `defineEmits`, no Options API.
- **Naming:** backend uses snake_case filenames (`user_repository.go`), exported symbols PascalCase. Frontend uses PascalCase components, camelCase functions.
- **Imports:** prefer `@/` alias in frontend over deep relative paths.

## What NOT to do

- Don't split `backend/.gitignore` or any other base config back into the subprojects — keep them at the root.
- Don't add another Node dependency manager. The project uses pnpm.
- Don't replace Naive UI or ECharts without checking the consuming pages first.
- Don't bypass the layered structure (handler → service → repository) in the backend.
- Don't introduce a top-level `client/` or `server/` directory — the repos were merged into `frontend/` and `backend/` deliberately.

## Docker

- Compose file: `docker-compose.yml` at root.
- Builds use `context: .` with relative dockerfile paths (`./backend/Dockerfile`, `./frontend/Dockerfile`). The root `.dockerignore` is honored for both.
- Backend image: multi-stage Go → distroless static, nonroot user.
- Frontend image: multi-stage pnpm build → nginx, with `/api` reverse-proxied to the `backend` service.
- Mongo data persists in the named volume `mongo_data`.
