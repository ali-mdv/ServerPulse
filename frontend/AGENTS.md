# ServerPulse UI(Frontend) — Agent guidance

- **Stack:** Vue 3 (`<script setup>`), vue-router 4, Pinia, VueUse, Axios, Naive UI, ECharts (via `vue-echarts`), TailwindCSS 3, TypeScript, Vite 7.
- **Pages** live in `frontend/src/pages/`. Components in `frontend/src/components/`. Pinia stores in `frontend/src/stores/`. Typed API clients in `frontend/src/api/`.
- **Routing:** add a new page by dropping a `.vue` in `pages/` and registering it in `frontend/src/router/index.ts`.
- **API calls:** import from `@/api/*`, not raw axios. The axios instance in `plugins/axios.ts` handles auth header injection.
- **State:** Pinia only. Don't introduce another state library.
- **Forms:** `vee-validate` + `yup` (see `pages/Profile.vue`, `pages/AddUser.vue` for the pattern).
- **Charts:** ECharts via `vue-echarts`. Don't pull in chart.js.
- **Styling:** Tailwind utility classes + Naive UI components. The `cn()` helper from `lib/utils.ts` is the way to compose class strings.
- **Tests:** Vitest. Co-locate as `*.spec.ts`.
- **TypeScript:** strict mode (`tsconfig.json`). Don't loosen `strict` or `noImplicitAny`.

## Conventions

- **Vue components:** `<script setup lang="ts">`. Props via `defineProps`, emits via `defineEmits`, no Options API.
- **Naming:** Uses PascalCase components, camelCase functions.
- **Imports:** prefer `@/` alias in frontend over deep relative paths.

## What NOT to do

- Don't replace Naive UI or ECharts without checking the consuming pages first.
