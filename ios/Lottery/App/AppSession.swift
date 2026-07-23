import Foundation
import Observation

enum SessionState: Equatable {
    case checking
    case signedOut
    case signedIn
    case configurationError(String)
}

enum AppTab: Hashable {
    case dashboard
    case recommendations
    case record
    case history
}

@MainActor
@Observable
final class AppSession {
    private(set) var state: SessionState = .checking
    private(set) var user: UserProfile?
    let api: APIClient?

    var selectedTab = AppTab.dashboard
    var ticketDraft = TicketDraft()
    var editingTicket: Ticket?
    var editorCanSave = false
    var editorSaving = false
    var saveTicketTrigger = 0
    var message: String?

    init(bundle: Bundle = .main) {
        if ProcessInfo.processInfo.arguments.contains("UITEST_RESET_SESSION") {
            KeychainStore().deleteToken()
        }
        do {
            api = APIClient(configuration: try APIConfiguration.current(bundle: bundle))
            Task { await restoreSession() }
        } catch {
            api = nil
            state = .configurationError(error.localizedDescription)
        }
    }

    func login(username: String, password: String) async throws {
        guard let api else { throw APIError.invalidConfiguration }
        let result = try await api.login(username: username, password: password)
        try await api.setToken(result.token)
        user = try await api.profile()
        state = .signedIn
    }

    func logout(message: String? = nil) {
        if let api {
            Task { try? await api.setToken(nil) }
        }
        user = nil
        state = .signedOut
        selectedTab = .dashboard
        resetEditor()
        self.message = message
    }

    func recordPurchase(_ recommendation: Recommendation) {
        let kind = LotteryKind(rawValue: recommendation.lotteryCode) ?? .ssq
        ticketDraft = TicketDraft(
            lottery: kind,
            issue: recommendation.issue,
            drawDate: recommendation.drawDate.flatMap(DateParser.date) ?? Date(),
            purchasedAt: Date(),
            costAmount: 0,
            notes: "",
            entries: recommendation.entries.map { entry in
                TicketEntryDraft(
                    red: Set(entry.redNumbers.lotteryNumbers),
                    blue: Set(entry.blueNumbers.lotteryNumbers)
                )
            },
            recommendationId: recommendation.id
        )
        ticketDraft.costAmount = ticketDraft.calculatedCost
        editingTicket = nil
        selectedTab = .record
    }

    func edit(_ ticket: Ticket) {
        let kind = LotteryKind(rawValue: ticket.lotteryCode) ?? .ssq
        ticketDraft = TicketDraft(
            lottery: kind,
            issue: ticket.issue,
            drawDate: (ticket.drawDate ?? ticket.manualDrawDate).flatMap(DateParser.date) ?? Date(),
            purchasedAt: DateParser.date(ticket.purchasedAt) ?? Date(),
            costAmount: ticket.costAmount,
            notes: ticket.notes,
            entries: ticket.entries.map { entry in
                TicketEntryDraft(
                    red: Set(entry.redNumbers.lotteryNumbers),
                    blue: Set(entry.blueNumbers.lotteryNumbers),
                    multiple: max(1, entry.multiple),
                    isAdditional: entry.isAdditional
                )
            },
            recommendationId: ticket.recommendationId
        )
        editingTicket = ticket
        selectedTab = .record
    }

    func resetEditor() {
        ticketDraft = TicketDraft()
        editingTicket = nil
        editorCanSave = false
        editorSaving = false
    }

    private func restoreSession() async {
        guard let api else { return }
        guard await api.hasStoredToken() else {
            state = .signedOut
            return
        }
        do {
            user = try await api.profile()
            state = .signedIn
        } catch {
            try? await api.setToken(nil)
            state = .signedOut
        }
    }
}

enum DateParser {
    static func date(_ value: String) -> Date? {
        let fractional = ISO8601DateFormatter()
        fractional.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        let standard = ISO8601DateFormatter()
        standard.formatOptions = [.withInternetDateTime]
        for formatter in [fractional, standard] {
            if let date = formatter.date(from: value) { return date }
        }
        return LotteryFormatters.dateOnly.date(from: String(value.prefix(10)))
    }
}
