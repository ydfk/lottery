const rawVersion = import.meta.env.VITE_APP_VERSION?.trim();
const rawBuildId = import.meta.env.VITE_APP_BUILD_ID?.trim();

export const APP_VERSION = rawVersion || "dev-local";
export const APP_BUILD_ID = rawBuildId || APP_VERSION;
export const APP_VERSION_LABEL = `版本 ${APP_VERSION}`;
