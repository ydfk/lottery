import type {
  PasskeyCredentialResponse,
  PasskeyLoginOptions,
  PasskeyRegistrationOptions,
} from "@/types/auth";

function decodeBase64URL(value: string) {
  const normalized = value.replace(/-/g, "+").replace(/_/g, "/");
  const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, "=");
  const binary = window.atob(padded);
  return Uint8Array.from(binary, (character) => character.charCodeAt(0));
}

function encodeBase64URL(value: ArrayBuffer) {
  const bytes = new Uint8Array(value);
  let binary = "";
  for (const byte of bytes) {
    binary += String.fromCharCode(byte);
  }
  return window.btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/g, "");
}

function ensurePasskeySupport() {
  if (!window.isSecureContext || !("PublicKeyCredential" in window)) {
    throw new Error("当前浏览器或连接环境不支持 Passkey，请使用 HTTPS 和较新的浏览器");
  }
}

export async function createPasskeyCredential(
  options: PasskeyRegistrationOptions
): Promise<PasskeyCredentialResponse> {
  ensurePasskeySupport();
  const source = options.publicKey;
  const publicKey = {
    ...source,
    challenge: decodeBase64URL(source.challenge),
    user: { ...source.user, id: decodeBase64URL(source.user.id) },
    excludeCredentials: source.excludeCredentials?.map((credential) => ({
      ...credential,
      id: decodeBase64URL(credential.id),
    })),
  } as PublicKeyCredentialCreationOptions;
  const credential = await navigator.credentials.create({ publicKey });
  if (!(credential instanceof PublicKeyCredential)) {
    throw new Error("没有取得有效的 Passkey 凭据");
  }
  const response = credential.response as AuthenticatorAttestationResponse;
  return {
    id: credential.id,
    rawId: encodeBase64URL(credential.rawId),
    type: "public-key",
    response: {
      clientDataJSON: encodeBase64URL(response.clientDataJSON),
      attestationObject: encodeBase64URL(response.attestationObject),
    },
  };
}

export async function getPasskeyCredential(
  options: PasskeyLoginOptions
): Promise<PasskeyCredentialResponse> {
  ensurePasskeySupport();
  const source = options.publicKey;
  const publicKey = {
    ...source,
    challenge: decodeBase64URL(source.challenge),
    allowCredentials: source.allowCredentials?.map((credential) => ({
      ...credential,
      id: decodeBase64URL(credential.id),
    })),
  } as PublicKeyCredentialRequestOptions;
  const credential = await navigator.credentials.get({ publicKey });
  if (!(credential instanceof PublicKeyCredential)) {
    throw new Error("没有取得有效的 Passkey 凭据");
  }
  const response = credential.response as AuthenticatorAssertionResponse;
  return {
    id: credential.id,
    rawId: encodeBase64URL(credential.rawId),
    type: "public-key",
    response: {
      clientDataJSON: encodeBase64URL(response.clientDataJSON),
      authenticatorData: encodeBase64URL(response.authenticatorData),
      signature: encodeBase64URL(response.signature),
      userHandle: response.userHandle ? encodeBase64URL(response.userHandle) : null,
    },
  };
}
