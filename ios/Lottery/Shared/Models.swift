import Foundation

struct UserProfile: Codable, Identifiable, Sendable {
    let id: String
    let username: String
}

struct AuthToken: Codable, Sendable {
    let token: String
}

struct LotteryType: Codable, Identifiable, Sendable {
    let id: String
    let code: String
    let name: String
    let status: String
    let redCount: Int
    let blueCount: Int
    let redMin: Int
    let redMax: Int
    let blueMin: Int
    let blueMax: Int
    let recommendationCount: Int

    static let ssq = rules(for: .ssq)
    static let dlt = rules(for: .dlt)

    static func rules(for kind: LotteryKind) -> LotteryType {
        switch kind {
        case .ssq:
            return LotteryType(
                id: "ssq", code: "ssq", name: "双色球", status: "enabled",
                redCount: 6, blueCount: 1, redMin: 1, redMax: 33,
                blueMin: 1, blueMax: 16, recommendationCount: 5
            )
        case .dlt:
            return LotteryType(
                id: "dlt", code: "dlt", name: "大乐透", status: "enabled",
                redCount: 5, blueCount: 2, redMin: 1, redMax: 35,
                blueMin: 1, blueMax: 12, recommendationCount: 5
            )
        }
    }
}

enum LotteryKind: String, Codable, CaseIterable, Identifiable, Sendable {
    case ssq
    case dlt

    var id: String { rawValue }
    var name: String { self == .ssq ? "双色球" : "大乐透" }
    var rules: LotteryType { LotteryType.rules(for: self) }
}

struct DashboardData: Codable, Sendable {
    let lottery: LotteryType?
    let latestDraw: DrawResult?
    let latestRecommendation: Recommendation?
    let recentTickets: [Ticket]
    let stats: DashboardStats
}

struct DashboardStats: Codable, Sendable {
    let totalTickets: Int
    let wonTickets: Int
    let totalCost: Double
    let totalPrize: Double
    let totalRecommendations: Int
    let purchasedRecommendations: Int

    var balance: Double { totalPrize - totalCost }
}

struct DrawResult: Codable, Identifiable, Sendable {
    let id: String
    let createdAt: String
    let updatedAt: String
    let lotteryCode: String
    let issue: String
    let drawDate: String
    let redNumbers: String
    let blueNumbers: String
    let saleAmount: Double
    let prizePoolAmount: Double
    let firstPrizeAmount: Double?
    let secondPrizeAmount: Double?
    let source: String
    let rawPayload: String
    let prizeDetails: [DrawPrize]
}

struct DrawQuery: Equatable, Sendable {
    var page = 1
    var pageSize = 20
    var lotteryCode = ""
    var issue = ""
    var drawDate = ""
    var sort = "latest"
}

struct DrawPrize: Codable, Identifiable, Sendable {
    let id: String
    let prizeName: String
    let prizeRule: String
    let winnerCount: Int
    let singleBonus: Double
}

struct DrawPage: Codable, Sendable {
    let items: [DrawResult]
    let page: Int
    let pageSize: Int
    let total: Int
    let hasMore: Bool
}

struct Recommendation: Codable, Identifiable, Sendable {
    let id: String
    let lotteryCode: String
    let issue: String
    let drawDate: String?
    let provider: String
    let model: String
    let strategy: String
    let promptVersion: String
    let summary: String
    let basis: String
    let checkedAt: String?
    let prizeAmount: Double
    let createdAt: String
    let entries: [RecommendationEntry]
    let drawRedNumbers: String?
    let drawBlueNumbers: String?
    let entryCount: Int?
    let winningCount: Int?
    let isPurchased: Bool?
    let purchasedCount: Int?
    let purchasedTickets: [Ticket]?
}

struct RecommendationEntry: Codable, Identifiable, Sendable {
    let id: String
    let sequence: Int
    let redNumbers: String
    let blueNumbers: String
    let confidence: Double
    let reason: String
    let isWinning: Bool
    let prizeName: String
    let prizeAmount: Double
    let matchSummary: String
}

struct RecommendationPage: Codable, Sendable {
    let items: [Recommendation]
    let page: Int
    let pageSize: Int
    let total: Int
    let hasMore: Bool
}

struct Ticket: Codable, Identifiable, Sendable {
    let id: String
    let createdAt: String?
    let updatedAt: String?
    let lotteryCode: String
    let recommendationId: String?
    let issue: String
    let manualDrawDate: String?
    let source: String
    let recognizedText: String
    let status: String
    let costAmount: Double
    let prizeAmount: Double
    let purchasedAt: String
    let checkedAt: String?
    let notes: String
    let entries: [TicketEntry]
    let imageUrl: String
    let drawDate: String?
    let drawRedNumbers: String
    let drawBlueNumbers: String
    let recommendation: TicketRecommendation?
}

struct TicketEntry: Codable, Identifiable, Sendable {
    let id: String
    let sequence: Int
    let redNumbers: String
    let blueNumbers: String
    let multiple: Int
    let isAdditional: Bool
    let isWinning: Bool
    let prizeName: String
    let prizeAmount: Double
    let matchSummary: String
}

struct TicketRecommendation: Codable, Identifiable, Sendable {
    let id: String
    let issue: String
    let drawDate: String?
    let summary: String
    let createdAt: String
    let entries: [TicketRecommendationEntry]
}

struct TicketRecommendationEntry: Codable, Identifiable, Sendable {
    let id: String
    let sequence: Int
    let redNumbers: String
    let blueNumbers: String
    let prizeAmount: Double
    let prizeName: String
}

struct TicketPage: Codable, Sendable {
    let items: [Ticket]
    let page: Int
    let pageSize: Int
    let total: Int
    let hasMore: Bool
}

struct TicketUpload: Codable, Identifiable, Sendable {
    let id: String
    let lotteryCode: String
    let status: String
    let originalFilename: String
    let imagePath: String
    let imageUrl: String
    let recognizedText: String
    let recognitionIssue: String
    let recognitionConfidence: Double
    let errorMessage: String
}

struct RecognitionDraft: Codable, Sendable {
    let upload: TicketUpload
    let lotteryCode: String
    let issue: String
    let drawDate: String
    let costAmount: Double
    let rawText: String
    let confidence: Double
    let entries: [ParsedEntry]
}

struct ParsedEntry: Codable, Sendable {
    let red: [Int]
    let blue: [Int]
    let multiple: Int
    let isAdditional: Bool
}

struct SyncResult: Codable, Sendable {
    let lotteryCode: String
    let issue: String?
    let requestedCount: Int
    let syncedCount: Int
}

struct DeleteResult: Codable, Sendable {
    let deleted: Bool
}
