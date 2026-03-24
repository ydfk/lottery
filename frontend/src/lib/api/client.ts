import type { ApiResponse } from "@/types/lottery";

const TOKEN_KEY = "lottery_token";

function getRequestHeaders(body?: unknown) {
  const headers = new Headers();
  const token = localStorage.getItem(TOKEN_KEY);
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }
  if (!(body instanceof FormData)) {
    headers.set("Content-Type", "application/json");
  }
  return headers;
}

async function parseResponse<T>(response: Response): Promise<T> {
  const payload = (await response.json()) as ApiResponse<T>;
  if (!response.ok || !payload.flag) {
    if (response.status === 401) {
      localStorage.removeItem(TOKEN_KEY);
    }
    throw new Error(payload.msg || "请求失败");
  }
  return payload.data;
}

export async function apiGet<T>(url: string): Promise<T> {
  const response = await fetch(url, {
    headers: getRequestHeaders(),
  });
  return parseResponse<T>(response);
}

export async function apiPost<T, B = unknown>(url: string, body?: B): Promise<T> {
  const response = await fetch(url, {
    method: "POST",
    headers: getRequestHeaders(body),
    body: body instanceof FormData ? body : body ? JSON.stringify(body) : undefined,
  });
  return parseResponse<T>(response);
}

export async function apiDelete<T>(url: string): Promise<T> {
  const response = await fetch(url, {
    method: "DELETE",
    headers: getRequestHeaders(),
  });
  return parseResponse<T>(response);
}

export function getStoredToken() {
  return localStorage.getItem(TOKEN_KEY) || "";
}

export function setStoredToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token);
}

export function clearStoredToken() {
  localStorage.removeItem(TOKEN_KEY);
}
