import Charts
import SwiftUI

struct DashboardSection<Content: View>: View {
    let title: String
    let subtitle: String
    @ViewBuilder let content: Content

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack(alignment: .firstTextBaseline) {
                Text(title).font(.headline)
                Spacer()
                Text(subtitle).font(.caption).foregroundStyle(.secondary)
            }
            content
        }
    }
}

struct DashboardHero: View {
    var body: some View {
        ZStack(alignment: .bottomLeading) {
            LotteryTrail()
            VStack(alignment: .leading, spacing: 6) {
                Text("彩迹")
                    .font(.system(size: 32, weight: .bold, design: .rounded))
                Text("每一张票，都有迹可循")
                    .font(.subheadline)
                    .foregroundStyle(.white.opacity(0.72))
            }
            .foregroundStyle(.white)
            .padding(21)
        }
        .frame(height: 142)
        .background(LotteryPalette.ink, in: .rect(cornerRadius: 28))
        .accessibilityElement(children: .combine)
    }
}

struct DashboardBalanceCard: View {
    let stats: DashboardStats

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack(alignment: .firstTextBaseline) {
                VStack(alignment: .leading, spacing: 3) {
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
                dashboardStat("累计花费", LotteryFormatters.currency(stats.totalCost))
                Spacer()
                dashboardStat("累计中奖", LotteryFormatters.currency(stats.totalPrize))
                Spacer()
                dashboardStat("中奖票数", "\(stats.wonTickets)")
            }
        }
        .padding(18)
        .background(Color(.secondarySystemGroupedBackground), in: .rect(cornerRadius: 24))
    }
}

struct DashboardEfficiencyCard: View {
    let stats: DashboardStats

    var body: some View {
        HStack(spacing: 0) {
            efficiency("中奖率", stats.winningRate, icon: "target")
            Divider().frame(height: 42)
            efficiency("奖金回报", stats.returnRate, icon: "chart.line.uptrend.xyaxis")
            Divider().frame(height: 42)
            efficiency("推荐购买", stats.recommendationPurchaseRate, icon: "link")
        }
        .padding(.vertical, 14)
        .background(Color(.secondarySystemGroupedBackground), in: .rect(cornerRadius: 22))
    }

    private func efficiency(_ title: String, _ value: Double, icon: String) -> some View {
        VStack(spacing: 5) {
            Label(title, systemImage: icon).font(.caption2).foregroundStyle(.secondary)
            Text(LotteryFormatters.percent(value)).font(.headline.monospacedDigit())
        }
        .frame(maxWidth: .infinity)
        .accessibilityElement(children: .combine)
    }
}

struct DashboardTrendCard: View {
    @Binding var range: DashboardTrendRange
    let trend: DashboardTrend?
    let isLoading: Bool
    let errorMessage: String?
    let retry: () -> Void

    var body: some View {
        DashboardSection(title: "收支走势", subtitle: "中国标准时间") {
            VStack(alignment: .leading, spacing: 12) {
                Picker("趋势范围", selection: $range) {
                    ForEach(DashboardTrendRange.allCases) { value in
                        Text(value.title).tag(value)
                    }
                }
                .pickerStyle(.segmented)

                if let trend {
                    trendSummary(trend)
                    trendChart(trend)
                } else if isLoading {
                    HStack { Spacer(); ProgressView(); Spacer() }.frame(height: 188)
                } else if let errorMessage {
                    VStack(alignment: .leading, spacing: 8) {
                        ErrorBanner(message: errorMessage)
                        Button("重新载入趋势", systemImage: "arrow.clockwise", action: retry)
                    }
                } else {
                    Text("暂无趋势数据").foregroundStyle(.secondary).frame(height: 188)
                }
            }
            .padding(16)
            .background(Color(.secondarySystemGroupedBackground), in: .rect(cornerRadius: 24))
        }
    }

    private func trendSummary(_ trend: DashboardTrend) -> some View {
        HStack {
            dashboardStat("区间净收益", LotteryFormatters.currency(trend.totalPrize - trend.totalCost))
            Spacer()
            dashboardStat("票据", "\(trend.ticketCount)")
            Spacer()
            dashboardStat("推荐", "\(trend.recommendationCount)")
        }
    }

    private func trendChart(_ trend: DashboardTrend) -> some View {
        Chart {
            ForEach(trend.items) { item in
                if let date = LotteryFormatters.dateOnly.date(from: item.periodStart) {
                    LineMark(
                        x: .value("日期", date),
                        y: .value("金额", item.totalCost),
                        series: .value("项目", "花费")
                    )
                    .foregroundStyle(LotteryPalette.red)
                    .symbol(.circle)
                    LineMark(
                        x: .value("日期", date),
                        y: .value("金额", item.totalPrize),
                        series: .value("项目", "中奖")
                    )
                    .foregroundStyle(LotteryPalette.jade)
                    .symbol(.diamond)
                }
            }
        }
        .chartLegend(.hidden)
        .chartYAxis { AxisMarks(position: .leading) }
        .frame(height: 178)
        .overlay(alignment: .topTrailing) {
            HStack(spacing: 10) {
                chartLegend("花费", color: LotteryPalette.red)
                chartLegend("中奖", color: LotteryPalette.jade)
            }
            .font(.caption2)
        }
        .accessibilityLabel("收支走势图，花费 \(LotteryFormatters.currency(trend.totalCost))，中奖 \(LotteryFormatters.currency(trend.totalPrize))")
    }

    private func chartLegend(_ title: String, color: Color) -> some View {
        Label { Text(title) } icon: { Circle().fill(color).frame(width: 7, height: 7) }
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
            context.stroke(
                path,
                with: .linearGradient(
                    Gradient(colors: [LotteryPalette.red, .white.opacity(0.8), LotteryPalette.blue]),
                    startPoint: .zero,
                    endPoint: CGPoint(x: size.width, y: size.height)
                ),
                lineWidth: reduceMotion ? 5 : 7
            )
        }
        .opacity(0.6)
        .accessibilityHidden(true)
    }
}

private func dashboardStat(_ title: String, _ value: String) -> some View {
    VStack(alignment: .leading, spacing: 3) {
        Text(title).font(.caption2).foregroundStyle(.secondary)
        Text(value).font(.subheadline.monospacedDigit().weight(.semibold))
    }
}
