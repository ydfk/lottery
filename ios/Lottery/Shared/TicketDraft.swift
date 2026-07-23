import Foundation

struct TicketEntryDraft: Codable, Identifiable, Equatable, Sendable {
    var id = UUID()
    var red: Set<Int> = []
    var blue: Set<Int> = []
    var multiple = 1
    var isAdditional = false

    static func clampedMultiple(_ value: Int) -> Int {
        min(99, max(1, value))
    }

    func isValid(for kind: LotteryKind) -> Bool {
        red.count == kind.rules.redCount
            && blue.count == kind.rules.blueCount
            && 1 ... 99 ~= multiple
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
    var notes = ""
    var entries = [TicketEntryDraft()]
    var recommendationId: String?
    var uploadId: String?

    var calculatedCost: Double {
        entries.reduce(0) { sum, entry in
            guard entry.isValid(for: lottery) else { return sum }
            return sum + Double(TicketEntryDraft.clampedMultiple(entry.multiple) * (entry.isAdditional ? 3 : 2))
        }
    }

    var totalMultiple: Int {
        entries.reduce(0) { sum, entry in
            guard entry.isValid(for: lottery) else { return sum }
            return sum + TicketEntryDraft.clampedMultiple(entry.multiple)
        }
    }

    var validEntryCount: Int {
        entries.count { $0.isValid(for: lottery) }
    }

    func recognizedCostDiffers(from amount: Double) -> Bool {
        amount > 0 && abs(amount - calculatedCost) > 0.005
    }

    var isValid: Bool {
        !issue.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
            && !entries.isEmpty
            && entries.allSatisfy { $0.isValid(for: lottery) }
            && calculatedCost > 0
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

    static let displayDateTimeFormatter: DateFormatter = {
        let formatter = DateFormatter()
        formatter.locale = Locale(identifier: "zh_CN")
        formatter.dateFormat = "yyyy-MM-dd HH:mm"
        return formatter
    }()

    static func currency(_ value: Double) -> String {
        currency.string(from: NSNumber(value: value)) ?? "¥0.00"
    }

    static func displayDate(_ value: String?) -> String {
        guard let value, !value.isEmpty else { return "待确认" }
        guard let date = DateParser.date(value) else { return String(value.prefix(10)) }
        return dateOnly.string(from: date)
    }

    static func displayDateTime(_ value: String?) -> String {
        guard let value, !value.isEmpty else { return "未知" }
        guard let date = DateParser.date(value) else {
            return String(value.prefix(16)).replacingOccurrences(of: "T", with: " ")
        }
        return displayDateTimeFormatter.string(from: date)
    }
}
