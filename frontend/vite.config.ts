/// <reference types="vitest" />
import { defineConfig, loadEnv, type Plugin } from "vite";
import react from "@vitejs/plugin-react";
import path from "path";

function appVersionPlugin(version: string, buildId: string): Plugin {
  return {
    name: "lottery-app-version",
    generateBundle() {
      this.emitFile({
        type: "asset",
        fileName: "app-version.json",
        source: `${JSON.stringify({ version, buildId }, null, 2)}\n`,
      });
    },
  };
}

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd());
  const localHost = "127.0.0.1";
  const proxyTarget = env.VITE_PROXY_HOST?.replace("localhost", localHost);
  const appVersion = (env.VITE_APP_VERSION || process.env.VITE_APP_VERSION || "dev-local").trim();
  const appBuildId = (
    env.VITE_APP_BUILD_ID ||
    process.env.VITE_APP_BUILD_ID ||
    `${appVersion}-${new Date().toISOString()}`
  ).trim();

  return {
    plugins: [react(), appVersionPlugin(appVersion, appBuildId)],
    define: {
      "import.meta.env.VITE_APP_VERSION": JSON.stringify(appVersion),
      "import.meta.env.VITE_APP_BUILD_ID": JSON.stringify(appBuildId),
    },
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
