import Foundation

extension Notification.Name {
    static let sessionUnauthorized = Notification.Name("LotterySessionUnauthorized")
}

enum APIError: LocalizedError, Equatable {
    case invalidConfiguration
    case invalidResponse
    case server(String, Int)
    case unauthorized
    case security(String)

    var errorDescription: String? {
        switch self {
        case .invalidConfiguration:
            return "服务地址未配置，请检查当前构建环境。"
        case .invalidResponse:
            return "服务器返回了无法识别的数据。"
        case let .server(message, _):
            return message
        case .unauthorized:
            return "登录已失效，请重新登录。"
        case let .security(message):
            return message
        }
    }
}

struct APIEnvelope<Value: Decodable>: Decodable {
    let flag: Bool
    let code: Int
    let data: Value?
    let msg: String?
}

struct APIConfiguration: Sendable {
    let baseURL: URL

    static func current(bundle: Bundle = .main) throws -> APIConfiguration {
        guard let value = bundle.object(forInfoDictionaryKey: "APIBaseURL") as? String,
              let url = URL(string: value),
              let host = url.host,
              !host.isEmpty
        else {
            throw APIError.invalidConfiguration
        }
        return APIConfiguration(baseURL: url)
    }
}

struct MultipartPart: Sendable {
    let name: String
    let filename: String
    let mimeType: String
    let data: Data
}

actor APIClient {
    private let configuration: APIConfiguration
    private let session: URLSession
    private let keychain: KeychainStore
    private var token: String?

    init(
        configuration: APIConfiguration,
        session: URLSession = .shared,
        keychain: KeychainStore = KeychainStore()
    ) {
        self.configuration = configuration
        self.session = session
        self.keychain = keychain
        token = keychain.readToken()
    }

    func setToken(_ value: String?) throws {
        token = value
        if let value {
            try keychain.saveToken(value)
        } else {
            keychain.deleteToken()
        }
    }

    func hasStoredToken() -> Bool {
        token != nil
    }

    func request<Value: Decodable>(
        _ path: String,
        method: String = "GET",
        query: [URLQueryItem] = [],
        body: Data? = nil,
        contentType: String? = nil
    ) async throws -> Value {
        var components = URLComponents(
            url: configuration.baseURL.appending(path: path),
            resolvingAgainstBaseURL: false
        )
        components?.queryItems = query.isEmpty ? nil : query
        guard let url = components?.url else {
            throw APIError.invalidConfiguration
        }

        var request = URLRequest(url: url)
        request.httpMethod = method
        request.timeoutInterval = 45
        request.httpBody = body
        request.setValue("application/json", forHTTPHeaderField: "Accept")
        if let token {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }
        if let contentType {
            request.setValue(contentType, forHTTPHeaderField: "Content-Type")
        }

        let (responseData, response) = try await session.data(for: request)
        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError.invalidResponse
        }
        let envelope = try? JSONDecoder().decode(APIEnvelope<Value>.self, from: responseData)
        if httpResponse.statusCode == 401 {
            try? setToken(nil)
            await MainActor.run {
                NotificationCenter.default.post(name: .sessionUnauthorized, object: nil)
            }
            throw APIError.unauthorized
        }
        guard httpResponse.statusCode < 400, envelope?.flag == true, let value = envelope?.data else {
            let message = envelope?.msg ?? HTTPURLResponse.localizedString(forStatusCode: httpResponse.statusCode)
            throw APIError.server(message, httpResponse.statusCode)
        }
        return value
    }

    func jsonBody<Value: Encodable>(_ value: Value) throws -> Data {
        try JSONEncoder().encode(value)
    }

    func multipartBody(parts: [MultipartPart]) -> (data: Data, contentType: String) {
        let boundary = "Lottery-\(UUID().uuidString)"
        var body = Data()
        for part in parts {
            body.append("--\(boundary)\r\n")
            body.append("Content-Disposition: form-data; name=\"\(part.name)\"; filename=\"\(part.filename)\"\r\n")
            body.append("Content-Type: \(part.mimeType)\r\n\r\n")
            body.append(part.data)
            body.append("\r\n")
        }
        body.append("--\(boundary)--\r\n")
        return (body, "multipart/form-data; boundary=\(boundary)")
    }

    func absoluteImageURL(_ value: String) -> URL? {
        if let url = URL(string: value), url.scheme != nil {
            return url
        }
        guard let origin = URL(string: "/", relativeTo: configuration.baseURL)?.absoluteURL else {
            return nil
        }
        return URL(string: value, relativeTo: origin)?.absoluteURL
    }
}

private extension Data {
    mutating func append(_ string: String) {
        append(Data(string.utf8))
    }
}

