/// <reference types="vitest" />
import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";
import path from "path";

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd());
  const localHost = "127.0.0.1";
  const proxyTarget = env.VITE_PROXY_HOST?.replace("localhost", localHost);

  return {
    plugins: [react()],
    resolve: {
      alias: {
        "@": path.resolve(__dirname, "./src"),
      },
    },
    server: {
      host: localHost,
      port: Number(env.VITE_PORT),
      proxy: proxyTarget
        ? {
            "/api": {
              target: proxyTarget,
              changeOrigin: true,
            },
            "/uploads": {
              target: proxyTarget,
              changeOrigin: true,
            },
          }
        : undefined,
    },
    preview: {
      host: localHost,
      port: Number(env.VITE_PORT),
    },
    test: {
      environment: "jsdom",
      environmentOptions: {
        jsdom: {
          url: `http://${localHost}/`,
        },
      },
      globals: true,
      exclude: ["**/node_modules/**", "**/.git/**", "**/.worktrees/**"],
      passWithNoTests: true,
      setupFiles: ["src/test/setup.ts"],
    },
  };
});
