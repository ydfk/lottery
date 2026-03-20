const rawVersion = import.meta.env.VITE_APP_VERSION?.trim();

export const APP_VERSION = rawVersion || "dev-local";
export const APP_VERSION_LABEL = `版本 ${APP_VERSION}`;
