import { apiDelete, apiGet, apiPost, clearStoredToken, setStoredToken } from "../client";
import { createPasskeyCredential, getPasskeyCredential } from "@/lib/passkey";
import type {
  AuthToken,
  AuthUser,
  PasskeyFinishPayload,
  PasskeyItem,
  PasskeyLoginOptions,
  PasskeyRegistrationOptions,
} from "@/types/auth";

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

export function beginPasskeyLogin() {
  return apiPost<PasskeyLoginOptions>("/api/auth/passkeys/login/options");
}

export function finishPasskeyLogin(payload: PasskeyFinishPayload) {
  return apiPost<AuthToken, PasskeyFinishPayload>("/api/auth/passkeys/login/verify", payload);
}

export async function loginWithPasskeyAndStoreToken() {
  const options = await beginPasskeyLogin();
  const credential = await getPasskeyCredential(options);
  const result = await finishPasskeyLogin({ ceremonyId: options.ceremonyId, credential });
  setStoredToken(result.token);
  return result;
}

export function beginPasskeyRegistration() {
  return apiPost<PasskeyRegistrationOptions>("/api/auth/passkeys/registration/options");
}

export function finishPasskeyRegistration(payload: PasskeyFinishPayload) {
  return apiPost<{ registered: boolean }, PasskeyFinishPayload>(
    "/api/auth/passkeys/registration/verify",
    payload
  );
}

export async function registerPasskey() {
  const options = await beginPasskeyRegistration();
  const credential = await createPasskeyCredential(options);
  return finishPasskeyRegistration({ ceremonyId: options.ceremonyId, credential });
}

export function getPasskeys() {
  return apiGet<PasskeyItem[]>("/api/auth/passkeys");
}

export function deletePasskey(id: string) {
  return apiDelete<{ deleted: boolean }>(`/api/auth/passkeys/${encodeURIComponent(id)}`);
}

export function logout() {
  clearStoredToken();
}
