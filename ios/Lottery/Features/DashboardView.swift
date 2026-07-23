import Observation
import SwiftUI

@MainActor
@Observable
private final class DashboardModel {
    var dashboard: DashboardData?
    var isLoading = false
    var generatingCode: String?
    var errorMessage: String?

    func load(using api: APIClient?) async {
        guard let api, !isLoading else { return }
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }
        do {
            dashboard = try await api.dashboard()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func generate(code: String, using api: APIClient?) async {
        guard let api, generatingCode == nil else { return }
        generatingCode = code
        errorMessage = nil
        defer { generatingCode = nil }
        do {
            _ = try await api.generateRecommendation(code: code)
            dashboard = try await api.dashboard()
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}

struct DashboardView: View {
    @Environment(AppSession.self) private var session
    @State private var model = DashboardModel()

    var body: some View {
        ScrollView {
            LazyVStack(spacing: 16) {
                hero
                if let errorMessage = model.errorMessage {
                    ErrorBanner(message: errorMessage)
                }
                if let dashboard = model.dashboard {
                    stats(dashboard.stats)
                    latestDraw(dashboard.latestDraw)
                    recommendationActions
                } else if model.isLoading {
                    LoadingOverlay(title: "正在载入看板")
                        .padding(.top, 48)
                }
            }
            .padding()
        }
        .navigationTitle("看板")
        .toolbar {
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
        .refreshable { await model.load(using: session.api) }
        .task { await model.load(using: session.api) }
    }

    private var hero: some View {
        ZStack(alignment: .bottomLeading) {
            LotteryTrail()
            VStack(alignment: .leading, spacing: 7) {
                Text("彩迹")
                    .font(.system(size: 34, weight: .bold, design: .rounded))
                Text("每一张票，都有迹可循")
                    .font(.subheadline)
                    .foregroundStyle(.white.opacity(0.72))
            }
            .foregroundStyle(.white)
            .padding(22)
        }
        .frame(height: 168)
        .background(LotteryPalette.ink, in: .rect(cornerRadius: 28))
        .accessibilityElement(children: .combine)
    }

    private func stats(_ stats: DashboardStats) -> some View {
        VStack(alignment: .leading, spacing: 14) {
            HStack(alignment: .firstTextBaseline) {
                VStack(alignment: .leading, spacing: 4) {
                    Text("累计结余").font(.caption).foregroundStyle(.secondary)
                    Text(LotteryFormatters.currency(stats.balance))
                        .font(.system(size: 30, weight: .bold, design: .rounded))
                        .foregroundStyle(stats.balance >= 0 ? LotteryPalette.jade : .primary)
                }
                Spacer()
                Text("共 \(stats.totalTickets) 张票")
                    .font(.footnote.weight(.medium))
                    .foregroundStyle(.secondary)
            }
            Divider()
            HStack {
                stat("累计花费", LotteryFormatters.currency(stats.totalCost))
                Spacer()
                stat("累计中奖", LotteryFormatters.currency(stats.totalPrize))
                Spacer()
                stat("中奖票数", "\(stats.wonTickets)")
            }
            .accessibilityElement(children: .contain)
        }
        .padding(18)
        .background(Color(.secondarySystemGroupedBackground), in: .rect(cornerRadius: 24))
    }

    private func stat(_ title: String, _ value: String) -> some View {
        VStack(alignment: .leading, spacing: 4) {
            Text(title).font(.caption).foregroundStyle(.secondary)
            Text(value).font(.headline.monospacedDigit())
        }
    }

    @ViewBuilder
    private func latestDraw(_ draw: DrawResult?) -> some View {
        Button {
            session.selectedTab = .draws
        } label: {
            LotteryCard {
                VStack(alignment: .leading, spacing: 14) {
                    HStack {
                        Label("最新开奖", systemImage: "clock.badge.checkmark")
                            .font(.headline)
                        Spacer()
                        Image(systemName: "chevron.right").foregroundStyle(.tertiary)
                    }
                    if let draw {
                        Text("\(LotteryKind(rawValue: draw.lotteryCode)?.name ?? draw.lotteryCode) · 第 \(draw.issue) 期")
                            .font(.subheadline)
                            .foregroundStyle(.secondary)
                        NumberBalls(redNumbers: draw.redNumbers, blueNumbers: draw.blueNumbers)
                    } else {
                        Text("暂无已同步的开奖信息").foregroundStyle(.secondary)
                    }
                }
            }
        }
        .buttonStyle(.plain)
    }

    private var recommendationActions: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("生成推荐").font(.headline)
            HStack(spacing: 12) {
                generationButton(kind: .ssq, tint: LotteryPalette.red)
                generationButton(kind: .dlt, tint: LotteryPalette.blue)
            }
        }
    }

    private func generationButton(kind: LotteryKind, tint: Color) -> some View {
        Button {
            Task { await model.generate(code: kind.rawValue, using: session.api) }
        } label: {
            HStack {
                if model.generatingCode == kind.rawValue { ProgressView() }
                Text(model.generatingCode == kind.rawValue ? "生成中" : kind.name)
                Spacer()
                Image(systemName: "sparkles")
            }
            .padding(16)
            .foregroundStyle(tint)
            .background(tint.opacity(0.1), in: .rect(cornerRadius: 18))
        }
        .buttonStyle(.plain)
        .disabled(model.generatingCode != nil)
    }
}

private struct LotteryTrail: View {
    @Environment(\.accessibilityReduceMotion) private var reduceMotion

    var body: some View {
        Canvas { context, size in
            var path = Path()
            path.move(to: CGPoint(x: size.width * 0.12, y: size.height * 0.35))
            path.addCurve(
                to: CGPoint(x: size.width * 0.92, y: size.height * 0.58),
                control1: CGPoint(x: size.width * 0.42, y: size.height * 0.02),
                control2: CGPoint(x: size.width * 0.66, y: size.height * 0.82)
            )
            context.stroke(path, with: .linearGradient(
                Gradient(colors: [LotteryPalette.red, .white.opacity(0.8), LotteryPalette.blue]),
                startPoint: .zero,
                endPoint: CGPoint(x: size.width, y: size.height)
            ), lineWidth: reduceMotion ? 5 : 7)
        }
        .opacity(0.6)
        .accessibilityHidden(true)
    }
}
