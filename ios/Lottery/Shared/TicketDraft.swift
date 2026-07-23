import Foundation

struct TicketEntryDraft: Codable, Identifiable, Equatable, Sendable {
    var id = UUID()
    var red: Set<Int> = []
    var blue: Set<Int> = []
    var multiple = 1
    var isAdditional = false

    func isValid(for kind: LotteryKind) -> Bool {
        red.count == kind.rules.redCount && blue.count == kind.rules.blueCount && multiple > 0
    }

    func payload() -> TicketEntryPayload {
        TicketEntryPayload(
            redNumbers: red.sorted().formattedNumbers,
            blueNumbers: blue.sorted().formattedNumbers,
            multiple: multiple,
            isAdditional: isAdditional
        )
    }
}

struct TicketDraft: Codable, Equatable, Sendable {
    var lottery = LotteryKind.ssq
    var issue = ""
    var drawDate = Date()
    var purchasedAt = Date()
    var costAmount = 0.0
    var notes = ""
    var entries = [TicketEntryDraft()]
    var recommendationId: String?
    var uploadId: String?

    var calculatedCost: Double {
        entries.reduce(0) { sum, entry in
            sum + Double(entry.multiple * (entry.isAdditional ? 3 : 2))
        }
    }

    var isValid: Bool {
        !issue.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
            && !entries.isEmpty
            && entries.allSatisfy { $0.isValid(for: lottery) }
            && costAmount > 0
    }
}

struct TicketDraftStore: Sendable {
    private let key = "lottery.ticket-draft.v1"

    func load() -> TicketDraft? {
        guard let data = UserDefaults.standard.data(forKey: key) else { return nil }
        return try? JSONDecoder().decode(TicketDraft.self, from: data)
    }

    func save(_ draft: TicketDraft) {
        guard let data = try? JSONEncoder().encode(draft) else { return }
        UserDefaults.standard.set(data, forKey: key)
    }

    func clear() {
        UserDefaults.standard.removeObject(forKey: key)
    }
}

extension Array where Element == Int {
    var formattedNumbers: String {
        map { String(format: "%02d", $0) }.joined(separator: ",")
    }
}

extension String {
    var lotteryNumbers: [Int] {
        split(separator: ",").compactMap { Int($0.trimmingCharacters(in: .whitespaces)) }
    }
}

enum LotteryFormatters {
    static let currency: NumberFormatter = {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        formatter.currencyCode = "CNY"
        formatter.maximumFractionDigits = 2
        return formatter
    }()

    static let dateOnly: DateFormatter = {
        let formatter = DateFormatter()
        formatter.locale = Locale(identifier: "zh_CN")
        formatter.dateFormat = "yyyy-MM-dd"
        return formatter
    }()

    static func currency(_ value: Double) -> String {
        currency.string(from: NSNumber(value: value)) ?? "¥0.00"
    }
}
