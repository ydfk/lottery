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

    @ViewBuilder
    var body: some View {
        @Bindable var session = session
        if #available(iOS 26.1, *) {
            mainTabs(selection: $session.selectedTab)
                .tabViewBottomAccessory(isEnabled: session.selectedTab == .record) {
                    saveTicketButton
                }
        } else if session.selectedTab == .record {
            mainTabs(selection: $session.selectedTab)
                .tabViewBottomAccessory {
                    saveTicketButton
                }
        } else {
            mainTabs(selection: $session.selectedTab)
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
            Tab("记录", systemImage: "ticket.fill", value: AppTab.record) {
                NavigationStack { TicketEditorView() }
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

    private var saveTicketButton: some View {
        Button {
            session.saveTicketTrigger += 1
        } label: {
            Label(
                session.editorSaving ? "保存中" : session.editingTicket == nil ? "保存票据" : "保存修改",
                systemImage: "checkmark.circle.fill"
            )
            .frame(maxWidth: .infinity)
        }
        .buttonStyle(.glassProminent)
        .disabled(!session.editorCanSave || session.editorSaving)
        .padding(.horizontal)
        .accessibilityIdentifier("save-ticket")
    }
}
