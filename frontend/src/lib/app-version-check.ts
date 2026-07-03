import { APP_BUILD_ID } from "@/lib/app-meta";

interface AppVersionPayload {
  buildId?: string;
  version?: string;
}

const VERSION_URL = "/app-version.json";
const RELOAD_MARKER_KEY = "lottery:app-version-reload";
const MIN_CHECK_INTERVAL = 30_000;

let checking = false;
let lastCheckedAt = 0;

function shouldSkipCheck(force: boolean) {
  if (force) {
    return false;
  }
  return Date.now() - lastCheckedAt < MIN_CHECK_INTERVAL;
}

function reloadWithBuildId(buildId: string) {
  const marker = getReloadMarker();
  if (marker === buildId) {
    return;
  }

  setReloadMarker(buildId);
  const url = new URL(window.location.href);
  url.searchParams.set("_app_v", buildId);
  window.location.replace(url.toString());
}

function getReloadMarker() {
  try {
    return window.sessionStorage.getItem(RELOAD_MARKER_KEY);
  } catch {
    return null;
  }
}

function setReloadMarker(buildId: string) {
  try {
    window.sessionStorage.setItem(RELOAD_MARKER_KEY, buildId);
  } catch {
    // 无法写入 sessionStorage 时仍尝试刷新，下一次启动会重新检测版本。
  }
}

async function checkAppVersion(force = false) {
  if (checking || shouldSkipCheck(force)) {
    return;
  }

  checking = true;
  lastCheckedAt = Date.now();
  try {
    const response = await fetch(`${VERSION_URL}?t=${Date.now()}`, {
      cache: "no-store",
      headers: {
        "Cache-Control": "no-cache",
      },
    });
    if (!response.ok) {
      return;
    }

    const payload = (await response.json()) as AppVersionPayload;
    const nextBuildId = payload.buildId?.trim();
    if (nextBuildId && nextBuildId !== APP_BUILD_ID) {
      reloadWithBuildId(nextBuildId);
    }
  } catch {
    // 版本探测失败不影响正常使用，下次回到前台时会再次检查。
  } finally {
    checking = false;
  }
}

export function initAppVersionCheck() {
  if (typeof window === "undefined") {
    return;
  }

  void checkAppVersion(true);
  window.addEventListener("focus", () => void checkAppVersion());
  window.addEventListener("pageshow", () => void checkAppVersion(true));
  document.addEventListener("visibilitychange", () => {
    if (document.visibilityState === "visible") {
      void checkAppVersion();
    }
  });
}
