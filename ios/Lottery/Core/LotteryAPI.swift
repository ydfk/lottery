import Foundation

struct LoginPayload: Encodable { let username: String; let password: String }
struct RecognizePayload: Encodable { let uploadId: String; let ocrText: String? }
struct SyncPayload: Encodable { let issue: String }

struct TicketEntryPayload: Codable, Sendable {
    let redNumbers: String
    let blueNumbers: String
    let multiple: Int
    let isAdditional: Bool
}

struct TicketPayload: Encodable, Sendable {
    let lotteryCode: String
    let uploadId: String?
    let recommendationId: String?
    let issue: String
    let drawDate: String
    let purchasedAt: String
    let costAmount: Double
    let notes: String
    let entries: [TicketEntryPayload]
}

extension APIClient {
    func login(username: String, password: String) async throws -> AuthToken {
        let body = try jsonBody(LoginPayload(username: username, password: password))
        return try await request("auth/login", method: "POST", body: body, contentType: "application/json")
    }

    func profile() async throws -> UserProfile {
        try await request("auth/profile")
    }

    func dashboard() async throws -> DashboardData {
        try await request("lotteries/dashboard")
    }

    func lotteryTypes() async throws -> [LotteryType] {
        try await request("lotteries/")
    }

    func recommendations(page: Int, lottery: String, status: String, sort: String) async throws -> RecommendationPage {
        var query = [
            URLQueryItem(name: "page", value: String(page)),
            URLQueryItem(name: "pageSize", value: "20"),
            URLQueryItem(name: "sort", value: sort),
        ]
        if !lottery.isEmpty { query.append(URLQueryItem(name: "lotteryCode", value: lottery)) }
        if !status.isEmpty { query.append(URLQueryItem(name: "status", value: status)) }
        return try await request("lotteries/recommendations", query: query)
    }

    func recommendation(code: String, id: String) async throws -> Recommendation {
        try await request("lotteries/\(code)/recommendations/\(id)")
    }

    func generateRecommendation(code: String) async throws -> Recommendation {
        try await request("lotteries/\(code)/recommendations/generate", method: "POST")
    }

    func recheckRecommendation(code: String, id: String) async throws -> Recommendation {
        try await request("lotteries/\(code)/recommendations/\(id)/recheck", method: "POST")
    }

    func deleteRecommendation(code: String, id: String) async throws -> DeleteResult {
        try await request("lotteries/\(code)/recommendations/\(id)", method: "DELETE")
    }

    func tickets(page: Int, lottery: String, status: String, sort: String) async throws -> TicketPage {
        var query = [
            URLQueryItem(name: "page", value: String(page)),
            URLQueryItem(name: "pageSize", value: "20"),
            URLQueryItem(name: "sort", value: sort),
        ]
        if !lottery.isEmpty { query.append(URLQueryItem(name: "lotteryCode", value: lottery)) }
        if !status.isEmpty { query.append(URLQueryItem(name: "status", value: status)) }
        return try await request("lotteries/tickets/history", query: query)
    }

    func createTicket(_ payload: TicketPayload) async throws -> Ticket {
        let body = try jsonBody(payload)
        return try await request("lotteries/tickets", method: "POST", body: body, contentType: "application/json")
    }

    func updateTicket(id: String, payload: TicketPayload) async throws -> Ticket {
        let body = try jsonBody(payload)
        return try await request("lotteries/tickets/\(id)", method: "PUT", body: body, contentType: "application/json")
    }

    func recheckTicket(id: String) async throws -> Ticket {
        try await request("lotteries/tickets/\(id)/recheck", method: "POST")
    }

    func deleteTicket(id: String) async throws -> DeleteResult {
        try await request("lotteries/tickets/\(id)", method: "DELETE")
    }

    func uploadTicketImage(data: Data, filename: String) async throws -> TicketUpload {
        let multipart = multipartBody(parts: [
            MultipartPart(name: "image", filename: filename, mimeType: "image/jpeg", data: data),
        ])
        return try await request(
            "lotteries/tickets/upload-image",
            method: "POST",
            body: multipart.data,
            contentType: multipart.contentType
        )
    }

    func recognizeTicket(uploadId: String) async throws -> RecognitionDraft {
        let body = try jsonBody(RecognizePayload(uploadId: uploadId, ocrText: nil))
        return try await request(
            "lotteries/tickets/recognize",
            method: "POST",
            body: body,
            contentType: "application/json"
        )
    }

    func draws(page: Int, lottery: String, issue: String = "") async throws -> DrawPage {
        var query = [
            URLQueryItem(name: "page", value: String(page)),
            URLQueryItem(name: "pageSize", value: "20"),
            URLQueryItem(name: "sort", value: "latest"),
        ]
        if !lottery.isEmpty { query.append(URLQueryItem(name: "lotteryCode", value: lottery)) }
        if !issue.isEmpty { query.append(URLQueryItem(name: "issue", value: issue)) }
        return try await request("lotteries/draws/history", query: query)
    }

    func syncDraw(code: String, issue: String) async throws -> SyncResult {
        let body = try jsonBody(SyncPayload(issue: issue))
        return try await request(
            "lotteries/\(code)/draws/sync",
            method: "POST",
            body: body,
            contentType: "application/json"
        )
    }

    func importTickets(workbook: MultipartPart, imagesZip: MultipartPart?) async throws -> ImportResult {
        let multipart = multipartBody(parts: [workbook, imagesZip].compactMap { $0 })
        return try await request(
            "lotteries/tickets/import",
            method: "POST",
            body: multipart.data,
            contentType: multipart.contentType
        )
    }
}

