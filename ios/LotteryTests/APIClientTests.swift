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

    @Test("Passkey 登录挑战不携带用户名")
    func passkeyLoginOptionsContract() async throws {
        let api = makeClient()
        MockURLProtocol.handler = { request in
            #expect(request.url?.path == "/api/auth/passkeys/login/options")
            #expect(request.httpMethod == "POST")
            #expect(Self.bodyData(from: request) == nil)
            let json = #"{"flag":true,"code":200,"data":{"ceremonyId":"ceremony","publicKey":{"challenge":"Y2hhbGxlbmdl","rpId":"example.test"}}}"#
            return Self.response(for: request, json: json)
        }
        let result = try await api.passkeyLoginOptions()
        #expect(result.ceremonyId == "ceremony")
    }

    @Test("Passkey 管理使用列表和删除接口")
    func passkeyManagementContract() async throws {
        let api = makeClient()
        var requestIndex = 0
        MockURLProtocol.handler = { request in
            requestIndex += 1
            if requestIndex == 1 {
                #expect(request.url?.path == "/api/auth/passkeys")
                #expect(request.httpMethod == "GET")
                let json = #"{"flag":true,"code":200,"data":[{"id":"passkey-id","createdAt":"2026-07-31T12:00:00Z"}]}"#
                return Self.response(for: request, json: json)
            }
            #expect(request.url?.path == "/api/auth/passkeys/passkey-id")
            #expect(request.httpMethod == "DELETE")
            return Self.response(
                for: request,
                json: #"{"flag":true,"code":200,"data":{"deleted":true}}"#
            )
        }
        let items = try await api.passkeys()
        #expect(items.first?.id == "passkey-id")
        try await api.deletePasskey(id: "passkey-id")
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

    @Test("看板趋势查询传递时间范围")
    func dashboardTrendContract() async throws {
        let api = makeClient()
        MockURLProtocol.handler = { request in
            guard let url = request.url,
                  let components = URLComponents(url: url, resolvingAgainstBaseURL: false)
            else {
                throw APIError.invalidResponse
            }
            #expect(url.path == "/api/lotteries/dashboard/trend")
            #expect(components.queryItems?.contains(URLQueryItem(name: "range", value: "90d")) == true)
            let json = #"{"flag":true,"code":200,"data":{"range":"90d","bucket":"week","items":[]}}"#
            return Self.response(for: request, json: json)
        }
        let result = try await api.dashboardTrend(range: .ninetyDays)
        #expect(result.range == "90d")
        #expect(result.bucket == "week")
    }

    @Test("推荐查询可以启用待开奖优先")
    func recommendationPendingFirstContract() async throws {
        let api = makeClient()
        MockURLProtocol.handler = { request in
            guard let url = request.url,
                  let components = URLComponents(url: url, resolvingAgainstBaseURL: false)
            else {
                throw APIError.invalidResponse
            }
            let query = components.queryItems ?? []
            #expect(url.path == "/api/lotteries/recommendations")
            #expect(query.contains(URLQueryItem(name: "pendingFirst", value: "true")))
            #expect(query.contains(URLQueryItem(name: "sort", value: "latest")))
            let json = #"{"flag":true,"code":200,"data":{"items":[],"page":1,"pageSize":20,"total":0,"hasMore":false}}"#
            return Self.response(for: request, json: json)
        }
        let result = try await api.recommendations(
            page: 1,
            lottery: "",
            status: "",
            sort: "latest",
            pendingFirst: true
        )
        #expect(result.items.isEmpty)
    }

    @Test("开奖查询传递完整筛选条件")
    func drawQueryContract() async throws {
        let api = makeClient()
        MockURLProtocol.handler = { request in
            guard let url = request.url,
                  let components = URLComponents(url: url, resolvingAgainstBaseURL: false)
            else {
                throw APIError.invalidResponse
            }
            let query = components.queryItems ?? []
            #expect(request.url?.path == "/api/lotteries/draws/history")
            #expect(query.contains(URLQueryItem(name: "page", value: "2")))
            #expect(query.contains(URLQueryItem(name: "pageSize", value: "50")))
            #expect(query.contains(URLQueryItem(name: "lotteryCode", value: "dlt")))
            #expect(query.contains(URLQueryItem(name: "issue", value: "26088")))
            #expect(query.contains(URLQueryItem(name: "drawDate", value: "2026-07-22")))
            #expect(query.contains(URLQueryItem(name: "sort", value: "oldest")))
            let json = #"{"flag":true,"code":200,"data":{"items":[],"page":2,"pageSize":50,"total":0,"hasMore":false}}"#
            return Self.response(for: request, json: json)
        }
        let result = try await api.draws(DrawQuery(
            page: 2,
            pageSize: 50,
            lotteryCode: "dlt",
            issue: "26088",
            drawDate: "2026-07-22",
            sort: "oldest"
        ))
        #expect(result.page == 2)
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

    @Test("成功响应无法解码时返回明确错误")
    func invalidSuccessResponse() async {
        let api = makeClient()
        MockURLProtocol.handler = { request in
            Self.response(for: request, json: #"{"flag":true,"code":200,"data":{"unexpected":true}}"#)
        }
        await #expect(throws: APIError.invalidResponse) {
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
