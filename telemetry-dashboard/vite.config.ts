import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import path from "node:path";

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  server: {
    port: 5173,
    proxy: {
      "/telemetry": "http://localhost:8081",
      "/health": "http://localhost:8081",
      "/ready": "http://localhost:8081",
    },
  },
});

