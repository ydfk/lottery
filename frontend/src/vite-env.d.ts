/// <reference types="vite/client" />
/// <reference types="vitest" />
/// <reference types="@testing-library/jest-dom" />
interface ImportMetaEnv {
  VITE_APP_BUILD_ID: string;
  VITE_APP_VERSION: string;
  VITE_PORT: number;
  VITE_USE_MOCK: string;
  VITE_PROXY_HOST: string;
}
