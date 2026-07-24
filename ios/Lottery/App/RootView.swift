import SwiftUI

struct RootView: View {
    @Environment(AppSession.self) private var session

    var body: some View {
        ZStack {
            AppBackground()
            switch session.state {
            case .checking:
                LoadingOverlay(title: "正在恢复登录")
            case .signedOut:
                LoginView()
            case .signedIn:
                MainTabView()
            case let .configurationError(message):
                ContentUnavailableView(
                    "服务配置不可用",
                    systemImage: "network.slash",
                    description: Text(message)
                )
            }
        }
        .alert(
            "提示",
            isPresented: Binding(
                get: { session.message != nil },
                set: { if !$0 { session.message = nil } }
            )
        ) {
            Button("知道了") { session.message = nil }
        } message: {
            Text(session.message ?? "")
        }
    }
}

private struct MainTabView: View {
    @Environment(AppSession.self) private var session

    var body: some View {
        @Bindable var session = session
        mainTabs(selection: $session.selectedTab)
            .fullScreenCover(isPresented: $session.isEditorPresented) {
                NavigationStack { TicketEditorView() }
                    .environment(session)
            }
    }

    private func mainTabs(selection: Binding<AppTab>) -> some View {
        TabView(selection: selection) {
            Tab("看板", systemImage: "square.grid.2x2.fill", value: AppTab.dashboard) {
                NavigationStack { DashboardView() }
            }
            Tab("推荐", systemImage: "sparkles", value: AppTab.recommendations) {
                NavigationStack { RecommendationsView() }
            }
            Tab("历史", systemImage: "clock.arrow.circlepath", value: AppTab.history) {
                NavigationStack { HistoryView() }
            }
            Tab("开奖", systemImage: "chart.xyaxis.line", value: AppTab.draws) {
                NavigationStack { DrawResultsView() }
            }
        }
        .tabBarMinimizeBehavior(.onScrollDown)
    }
}
