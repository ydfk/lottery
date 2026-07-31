import AuthenticationServices
import UIKit

@MainActor
final class PasskeyAuthorizationCoordinator: NSObject {
    private var controller: ASAuthorizationController?
    private var continuation: CheckedContinuation<ASAuthorization, Error>?

    func register(with options: PasskeyRegistrationOptions) async throws -> PasskeyCredentialResponse {
        guard let challenge = Data(base64URL: options.publicKey.challenge),
              let userID = Data(base64URL: options.publicKey.user.id)
        else {
            throw APIError.invalidResponse
        }
        let provider = ASAuthorizationPlatformPublicKeyCredentialProvider(
            relyingPartyIdentifier: options.publicKey.rp.id
        )
        let request = provider.createCredentialRegistrationRequest(
            challenge: challenge,
            name: options.publicKey.user.name,
            userID: userID
        )
        let authorization = try await perform(request)
        guard let credential = authorization.credential as? ASAuthorizationPlatformPublicKeyCredentialRegistration,
              let attestationObject = credential.rawAttestationObject
        else {
            throw APIError.invalidResponse
        }
        return PasskeyCredentialResponse(
            id: credential.credentialID.base64URLString,
            rawId: credential.credentialID.base64URLString,
            response: .init(
                clientDataJSON: credential.rawClientDataJSON.base64URLString,
                attestationObject: attestationObject.base64URLString,
                authenticatorData: nil,
                signature: nil,
                userHandle: nil
            )
        )
    }

    func authenticate(with options: PasskeyLoginOptions) async throws -> PasskeyCredentialResponse {
        guard let challenge = Data(base64URL: options.publicKey.challenge) else {
            throw APIError.invalidResponse
        }
        let provider = ASAuthorizationPlatformPublicKeyCredentialProvider(
            relyingPartyIdentifier: options.publicKey.rpId
        )
        let request = provider.createCredentialAssertionRequest(challenge: challenge)
        let authorization = try await perform(request)
        guard let credential = authorization.credential as? ASAuthorizationPlatformPublicKeyCredentialAssertion else {
            throw APIError.invalidResponse
        }
        return PasskeyCredentialResponse(
            id: credential.credentialID.base64URLString,
            rawId: credential.credentialID.base64URLString,
            response: .init(
                clientDataJSON: credential.rawClientDataJSON.base64URLString,
                attestationObject: nil,
                authenticatorData: credential.rawAuthenticatorData.base64URLString,
                signature: credential.signature.base64URLString,
                userHandle: credential.userID.base64URLString
            )
        )
    }

    private func perform(_ request: ASAuthorizationRequest) async throws -> ASAuthorization {
        try await withCheckedThrowingContinuation { continuation in
            self.continuation = continuation
            let controller = ASAuthorizationController(authorizationRequests: [request])
            controller.delegate = self
            controller.presentationContextProvider = self
            self.controller = controller
            controller.performRequests()
        }
    }

    private func finish(_ result: Result<ASAuthorization, Error>) {
        continuation?.resume(with: result)
        continuation = nil
        controller = nil
    }
}

extension PasskeyAuthorizationCoordinator: ASAuthorizationControllerDelegate {
    func authorizationController(controller: ASAuthorizationController, didCompleteWithAuthorization authorization: ASAuthorization) {
        finish(.success(authorization))
    }

    func authorizationController(controller: ASAuthorizationController, didCompleteWithError error: Error) {
        finish(.failure(error))
    }
}

extension PasskeyAuthorizationCoordinator: ASAuthorizationControllerPresentationContextProviding {
    func presentationAnchor(for controller: ASAuthorizationController) -> ASPresentationAnchor {
        guard let windowScene = UIApplication.shared.connectedScenes
            .compactMap({ $0 as? UIWindowScene })
            .first
        else {
            fatalError("无法获取 Passkey 授权窗口")
        }
        return windowScene.keyWindow ?? ASPresentationAnchor(windowScene: windowScene)
    }
}

private extension Data {
    init?(base64URL value: String) {
        let padded = value
            .replacingOccurrences(of: "-", with: "+")
            .replacingOccurrences(of: "_", with: "/")
        let padding = String(repeating: "=", count: (4 - padded.count % 4) % 4)
        self.init(base64Encoded: padded + padding)
    }

    var base64URLString: String {
        base64EncodedString()
            .replacingOccurrences(of: "+", with: "-")
            .replacingOccurrences(of: "/", with: "_")
            .replacingOccurrences(of: "=", with: "")
    }
}
