import Observation
import SwiftUI

@MainActor
@Observable
private final class DashboardModel {
    var globalDashboard: DashboardData?
    var trend: DashboardTrend?
    var trendRange = DashboardTrendRange.thirtyDays
    var isLoading = false
    var isTrendLoading = false
    var errorMessage: String?
    var trendErrorMessage: String?

    func load(using api: APIClient?) async {
        guard let api, !isLoading else { return }
        isLoading = true
        defer { isLoading = false }

        async let globalResult = Self.fetchDashboard(using: api)
        async let trendResult = Self.fetchTrend(using: api, range: trendRange)
        let (global, trendResultValue) = await (globalResult, trendResult)

        globalDashboard = global.value
        errorMessage = global.error
        trend = trendResultValue.value
        trendErrorMessage = trendResultValue.error
    }

    func loadTrend(using api: APIClient?) async {
        guard let api, !isTrendLoading else { return }
        isTrendLoading = true
        defer { isTrendLoading = false }
        let result = await Self.fetchTrend(using: api, range: trendRange)
        trend = result.value
        trendErrorMessage = result.error
    }

    private static func fetchDashboard(using api: APIClient) async -> (value: DashboardData?, error: String?) {
        do {
            return (try await api.dashboard(), nil)
        } catch {
            return (nil, error.localizedDescription)
        }
    }

    private static func fetchTrend(
        using api: APIClient,
        range: DashboardTrendRange
    ) async -> (value: DashboardTrend?, error: String?) {
        do {
            return (try await api.dashboardTrend(range: range), nil)
        } catch {
            return (nil, error.localizedDescription)
        }
    }
}

struct DashboardView: View {
    @Environment(AppSession.self) private var session
    @State private var model = DashboardModel()

    var body: some View {
        ScrollView {
            LazyVStack(spacing: 14) {
                DashboardHero()
                if let errorMessage = model.errorMessage {
                    ErrorBanner(message: errorMessage)
                }
                if let dashboard = model.globalDashboard {
                    DashboardBalanceCard(stats: dashboard.stats)
                    DashboardEfficiencyCard(stats: dashboard.stats)
                } else if model.isLoading {
                    LoadingOverlay(title: "正在载入看板").padding(.top, 36)
                }

                DashboardTrendCard(
                    range: trendRangeBinding,
                    trend: model.trend,
                    isLoading: model.isLoading || model.isTrendLoading,
                    errorMessage: model.trendErrorMessage,
                    retry: { Task { await model.loadTrend(using: session.api) } }
                )
            }
            .padding()
        }
        .navigationTitle("看板")
        .toolbar { accountMenu }
        .refreshable { await model.load(using: session.api) }
        .task { await model.load(using: session.api) }
    }

    private var trendRangeBinding: Binding<DashboardTrendRange> {
        Binding(
            get: { model.trendRange },
            set: { value in
                model.trendRange = value
                Task { await model.loadTrend(using: session.api) }
            }
        )
    }

    @ToolbarContentBuilder
    private var accountMenu: some ToolbarContent {
        ToolbarItem(placement: .topBarTrailing) {
            Menu {
                Text(session.user?.username ?? "")
                Button("退出登录", systemImage: "rectangle.portrait.and.arrow.right", role: .destructive) {
                    session.logout()
                }
            } label: {
                Image(systemName: "person.crop.circle")
            }
            .accessibilityLabel("账户")
        }
    }

}
