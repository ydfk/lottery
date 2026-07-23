import Foundation
import Testing
@testable import Lottery

private final class MockURLProtocol: URLProtocol, @unchecked Sendable {
    nonisolated(unsafe) static var handler: ((URLRequest) throws -> (HTTPURLResponse, Data))?

    override class func canInit(with request: URLRequest) -> Bool { true }
    override class func canonicalRequest(for request: URLRequest) -> URLRequest { request }

    override func startLoading() {
        guard let handler = Self.handler else {
            client?.urlProtocol(self, didFailWithError: APIError.invalidResponse)
            return
        }
        do {
            let (response, data) = try handler(request)
            client?.urlProtocol(self, didReceive: response, cacheStoragePolicy: .notAllowed)
            client?.urlProtocol(self, didLoad: data)
            client?.urlProtocolDidFinishLoading(self)
        } catch {
            client?.urlProtocol(self, didFailWithError: error)
        }
    }

    override func stopLoading() {}
}

@Suite(.serialized)
struct APIClientTests {
    @Test("登录请求使用正确路径和 JSON")
    func loginContract() async throws {
        let api = makeClient()
        MockURLProtocol.handler = { request in
            #expect(request.url?.path == "/api/auth/login")
            #expect(request.httpMethod == "POST")
            #expect(request.value(forHTTPHeaderField: "Content-Type") == "application/json")
            let body = try #require(Self.bodyData(from: request))
            let object = try JSONSerialization.jsonObject(with: body) as? [String: String]
            #expect(object?["username"] == "alice")
            return Self.response(for: request, json: #"{"flag":true,"code":200,"data":{"token":"jwt"}}"#)
        }
        let result = try await api.login(username: "alice", password: "secret")
        #expect(result.token == "jwt")
    }

    @Test("历史分页查询传递筛选条件")
    func historyContract() async throws {
        let api = makeClient()
        MockURLProtocol.handler = { request in
            guard let url = request.url,
                  let components = URLComponents(url: url, resolvingAgainstBaseURL: false)
            else {
                throw APIError.invalidResponse
            }
            #expect(request.url?.path == "/api/lotteries/tickets/history")
            #expect(components.queryItems?.contains(URLQueryItem(name: "lotteryCode", value: "ssq")) == true)
            #expect(components.queryItems?.contains(URLQueryItem(name: "status", value: "won")) == true)
            let json = #"{"flag":true,"code":200,"data":{"items":[],"page":1,"pageSize":20,"total":0,"hasMore":false}}"#
            return Self.response(for: request, json: json)
        }
        let result = try await api.tickets(page: 1, lottery: "ssq", status: "won", sort: "latest")
        #expect(result.items.isEmpty)
    }

    @Test("401 转换为未授权错误")
    func unauthorized() async {
        let api = makeClient()
        MockURLProtocol.handler = { request in
            Self.response(for: request, status: 401, json: #"{"flag":false,"code":401,"msg":"未登录"}"#)
        }
        await #expect(throws: APIError.unauthorized) {
            _ = try await api.profile()
        }
    }

    private func makeClient() -> APIClient {
        let configuration = URLSessionConfiguration.ephemeral
        configuration.protocolClasses = [MockURLProtocol.self]
        return APIClient(
            configuration: APIConfiguration(baseURL: URL(string: "https://example.test/api")!),
            session: URLSession(configuration: configuration)
        )
    }

    private static func response(
        for request: URLRequest,
        status: Int = 200,
        json: String
    ) -> (HTTPURLResponse, Data) {
        let response = HTTPURLResponse(
            url: request.url!,
            statusCode: status,
            httpVersion: nil,
            headerFields: ["Content-Type": "application/json"]
        )!
        return (response, Data(json.utf8))
    }

    private static func bodyData(from request: URLRequest) -> Data? {
        if let body = request.httpBody { return body }
        guard let stream = request.httpBodyStream else { return nil }
        stream.open()
        defer { stream.close() }
        var result = Data()
        let buffer = UnsafeMutablePointer<UInt8>.allocate(capacity: 1024)
        defer { buffer.deallocate() }
        while stream.hasBytesAvailable {
            let count = stream.read(buffer, maxLength: 1024)
            guard count > 0 else { break }
            result.append(buffer, count: count)
        }
        return result
    }
}
