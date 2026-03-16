import { apiGet, apiPost, clearStoredToken, setStoredToken } from "../client";
import type { AuthToken, AuthUser } from "@/types/auth";

export function login(username: string, password: string) {
  return apiPost<AuthToken, { username: string; password: string }>("/api/auth/login", {
    username,
    password,
  });
}

export function register(username: string, password: string) {
  return apiPost<AuthUser, { username: string; password: string }>("/api/auth/register", {
    username,
    password,
  });
}

export function getProfile() {
  return apiGet<AuthUser>("/api/auth/profile");
}

export async function loginAndStoreToken(username: string, password: string) {
  const result = await login(username, password);
  setStoredToken(result.token);
  return result;
}

export function logout() {
  clearStoredToken();
}
