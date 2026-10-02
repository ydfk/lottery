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
    case history
    case draws
}

@MainActor
@Observable
final class AppSession {
    private(set) var state: SessionState = .checking
    private(set) var user: UserProfile?
    let api: APIClient?
    let isUITestSession: Bool

    var selectedTab = AppTab.dashboard
    var ticketDraft = TicketDraft()
    var editingTicket: Ticket?
    var editorSaving = false
    var isEditorPresented = false
    var ticketHistoryRevision = 0
    var message: String?

    init(bundle: Bundle = .main) {
        let launchArguments = ProcessInfo.processInfo.arguments
        let launchEnvironment = ProcessInfo.processInfo.environment
        let shouldResetSession = launchArguments.contains("UITEST_RESET_SESSION")
            || launchEnvironment["UITEST_RESET_SESSION"] == "1"
        let shouldUseTestSession = launchArguments.contains("UITEST_SIGNED_IN")
            || launchEnvironment["UITEST_SIGNED_IN"] == "1"
        isUITestSession = shouldUseTestSession

        if shouldResetSession {
            KeychainStore().deleteToken()
        }
        do {
            api = APIClient(configuration: try APIConfiguration.current(bundle: bundle))
            if shouldUseTestSession {
                user = UserProfile(id: "ui-test", username: "测试用户")
                state = .signedIn
            } else {
                Task { await restoreSession() }
            }
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

    func loginWithPasskey() async throws {
        guard let api else { throw APIError.invalidConfiguration }
        let options = try await api.passkeyLoginOptions()
        let credential = try await PasskeyAuthorizationCoordinator().authenticate(with: options)
        let result = try await api.finishPasskeyLogin(
            PasskeyFinishPayload(ceremonyId: options.ceremonyId, credential: credential)
        )
        try await api.setToken(result.token)
        user = try await api.profile()
        state = .signedIn
    }

    func registerPasskey() async throws {
        guard let api else { throw APIError.invalidConfiguration }
        let options = try await api.passkeyRegistrationOptions()
        let credential = try await PasskeyAuthorizationCoordinator().register(with: options)
        try await api.finishPasskeyRegistration(
            PasskeyFinishPayload(ceremonyId: options.ceremonyId, credential: credential)
        )
    }

    func passkeys() async throws -> [PasskeyItem] {
        guard let api else { throw APIError.invalidConfiguration }
        return try await api.passkeys()
    }

    func deletePasskey(id: String) async throws {
        guard let api else { throw APIError.invalidConfiguration }
        try await api.deletePasskey(id: id)
    }

    func logout(message: String? = nil) {
        if let api {
            Task { try? await api.setToken(nil) }
        }
        user = nil
        state = .signedOut
        selectedTab = .dashboard
        isEditorPresented = false
        resetEditor()
        self.message = message
    }

    func recordPurchase(_ recommendation: Recommendation) {
        ticketDraft = .purchase(recommendation)
        editingTicket = nil
        isEditorPresented = true
    }

    func edit(_ ticket: Ticket) {
        let kind = LotteryKind(rawValue: ticket.lotteryCode) ?? .ssq
        ticketDraft = TicketDraft(
            lottery: kind,
            issue: ticket.issue,
            drawDate: (ticket.drawDate ?? ticket.manualDrawDate).flatMap(DateParser.date) ?? Date(),
            purchasedAt: DateParser.date(ticket.purchasedAt) ?? Date(),
            notes: ticket.notes,
            entries: ticket.entries.map { entry in
                TicketEntryDraft(
                    red: Set(entry.redNumbers.lotteryNumbers),
                    blue: Set(entry.blueNumbers.lotteryNumbers),
                    multiple: TicketEntryDraft.clampedMultiple(entry.multiple),
                    isAdditional: entry.isAdditional
                )
            },
            recommendationId: ticket.recommendationId
        )
        editingTicket = ticket
        isEditorPresented = true
    }

    func startNewTicket() {
        resetEditor()
        isEditorPresented = true
    }

    func closeEditor(saved: Bool) {
        isEditorPresented = false
        resetEditor()
        guard saved else { return }
        ticketHistoryRevision += 1
        selectedTab = .history
    }

    func resetEditor() {
        ticketDraft = TicketDraft()
        editingTicket = nil
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
