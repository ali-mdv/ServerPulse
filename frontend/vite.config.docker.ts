import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import path from "path";

// Loading variant for Docker dev (frontend/Dockerfile.dev):
// used with `vite --configLoader native`, which loads TS configs without
// bundling -- important because the bundler writes a `_tmp_*` file next to
// the config and code is mounted read-only.
// Uses import.meta.dirname instead of __dirname (not defined in ESM).
export default defineConfig(({ mode }) => ({
  // Dep-optimizer cache must live outside the read-only /app bind mount.
  cacheDir: "/tmp/vite-cache",
  server: {
    host: "::",
    port: 8080,
    // Same-origin dev proxy: the browser calls /api/v1/... on this dev
    // server and vite forwards to the backend container. Avoids baking a
    // `localhost` URL into the app, so other LAN devices work too.
    proxy: {
      "/api": {
        target: process.env.VITE_PROXY_TARGET || "http://backend:12000",
        changeOrigin: true,
        ws: true,
      },
    },
    fs: {
      allow: ["./src", "./shared", "./node_modules"],
      deny: [".env", ".env.*", "*.{crt,pem}", "**/.git/**", "server/**"],
    },
  },
  build: {
    outDir: "dist/spa",
  },
  plugins: [vue()],
  resolve: {
    alias: {
      "@": path.resolve(import.meta.dirname, "./src"),
      "@shared": path.resolve(import.meta.dirname, "./shared"),
    },
  },
}));
