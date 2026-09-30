import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import path from "path";

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => ({
  server: {
    host: "::",
    port: 8080,
    // Dev proxy so the browser talks same-origin (/api/v1/...) and the
    // dev server forwards to the backend. This makes the app reachable
    // from other devices on the LAN without baking in `localhost`.
    proxy: {
      "/api": {
        target: process.env.VITE_PROXY_TARGET || "http://localhost:12000",
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
      "@": path.resolve(__dirname, "./src"),
      "@shared": path.resolve(__dirname, "./shared"),
    },
  },
}));
