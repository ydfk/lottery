export interface AuthToken {
  token: string;
}

export interface AuthUser {
  id: string;
  username: string;
}

interface PasskeyCredentialDescriptor {
  id: string;
  type: PublicKeyCredentialType;
  transports?: AuthenticatorTransport[];
}

export interface PasskeyRegistrationOptions {
  ceremonyId: string;
  publicKey: {
    challenge: string;
    rp: PublicKeyCredentialRpEntity;
    user: Omit<PublicKeyCredentialUserEntity, "id"> & { id: string };
    pubKeyCredParams: PublicKeyCredentialParameters[];
    timeout?: number;
    excludeCredentials?: PasskeyCredentialDescriptor[];
    authenticatorSelection?: AuthenticatorSelectionCriteria;
    attestation?: AttestationConveyancePreference;
  };
}

export interface PasskeyLoginOptions {
  ceremonyId: string;
  publicKey: {
    challenge: string;
    rpId: string;
    timeout?: number;
    allowCredentials?: PasskeyCredentialDescriptor[];
    userVerification?: UserVerificationRequirement;
  };
}

export interface PasskeyCredentialResponse {
  id: string;
  rawId: string;
  type: "public-key";
  response: {
    clientDataJSON: string;
    attestationObject?: string;
    authenticatorData?: string;
    signature?: string;
    userHandle?: string | null;
  };
}

export interface PasskeyFinishPayload {
  ceremonyId: string;
  credential: PasskeyCredentialResponse;
}

export interface PasskeyItem {
  id: string;
  createdAt: string;
}
