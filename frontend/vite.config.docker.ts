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
