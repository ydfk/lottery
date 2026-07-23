import Observation
import SwiftUI

struct DrawTrendFrequency: Equatable, Sendable {
    let red: [Int: Int]
    let blue: [Int: Int]
}

enum DrawTrendCalculator {
    static func frequencies(for draws: [DrawResult]) -> DrawTrendFrequency {
        var red: [Int: Int] = [:]
        var blue: [Int: Int] = [:]
        for draw in draws {
            for number in draw.redNumbers.lotteryNumbers { red[number, default: 0] += 1 }
            for number in draw.blueNumbers.lotteryNumbers { blue[number, default: 0] += 1 }
        }
        return DrawTrendFrequency(red: red, blue: blue)
    }
}

@MainActor
@Observable
private final class DrawTrendModel {
    var draws: [DrawResult] = []
    var recommendation: Recommendation?
    var isLoading = false
    var errorMessage: String?
    private var loadID = UUID()

    func load(kind: LotteryKind, using api: APIClient?) async {
        guard let api else { return }
        let requestID = UUID()
        loadID = requestID
        isLoading = true
        defer {
            if loadID == requestID { isLoading = false }
        }
        do {
            var query = DrawQuery()
            query.pageSize = 50
            query.lotteryCode = kind.rawValue
            let result = try await api.draws(query)
            let latestRecommendation = try? await api.recommendations(
                page: 1,
                lottery: kind.rawValue,
                status: "",
                sort: "latest",
                pageSize: 1
            ).items.first
            guard loadID == requestID else { return }
            draws = result.items
            recommendation = latestRecommendation
            errorMessage = nil
        } catch is CancellationError {
            return
        } catch {
            if loadID == requestID { errorMessage = error.localizedDescription }
        }
    }
}

struct DrawTrendView: View {
    @Environment(AppSession.self) private var session
    @State private var model = DrawTrendModel()
    @State private var lottery = LotteryKind.ssq
    let refreshID: UUID

    private var rules: LotteryType { lottery.rules }
    private var frequency: DrawTrendFrequency { DrawTrendCalculator.frequencies(for: model.draws) }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                Picker("彩票类型", selection: $lottery) {
                    ForEach(LotteryKind.allCases) { kind in Text(kind.name).tag(kind) }
                }
                .pickerStyle(.segmented)

                if let errorMessage = model.errorMessage { ErrorBanner(message: errorMessage) }

                if model.draws.isEmpty, !model.isLoading {
                    EmptyState(icon: "chart.xyaxis.line", title: "暂无走势数据", message: "同步开奖后即可计算最近 50 期走势。")
                } else {
                    trendSummary
                    matrix
                }

                if model.isLoading { HStack { Spacer(); ProgressView("计算走势"); Spacer() } }
            }
            .padding(.horizontal)
            .padding(.bottom, 24)
        }
        .refreshable { await model.load(kind: lottery, using: session.api) }
        .task(id: "\(lottery.rawValue)-\(refreshID.uuidString)") {
            await model.load(kind: lottery, using: session.api)
        }
    }

    private var trendSummary: some View {
        LotteryCard {
            VStack(alignment: .leading, spacing: 12) {
                HStack {
                    Label("最近 \(model.draws.count) 期", systemImage: "chart.bar.xaxis")
                        .font(.headline)
                    Spacer()
                    Text("客户端实时统计").font(.caption).foregroundStyle(.secondary)
                }
                HStack(alignment: .top, spacing: 20) {
                    hotNumbers(title: lottery == .dlt ? "前区高频" : "红球高频", values: frequency.red)
                    hotNumbers(title: lottery == .dlt ? "后区高频" : "蓝球高频", values: frequency.blue)
                }
            }
        }
    }

    private func hotNumbers(title: String, values: [Int: Int]) -> some View {
        VStack(alignment: .leading, spacing: 5) {
            Text(title).font(.caption).foregroundStyle(.secondary)
            Text(
                values.sorted { lhs, rhs in
                    lhs.value == rhs.value ? lhs.key < rhs.key : lhs.value > rhs.value
                }
                .prefix(4)
                .map { String(format: "%02d", $0.key) }
                .joined(separator: " · ")
            )
            .font(.subheadline.monospacedDigit().weight(.semibold))
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }

    private var matrix: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack {
                Text("号码矩阵").font(.headline)
                Spacer()
                Text("左右滑动查看全部号码").font(.caption).foregroundStyle(.secondary)
            }

            ScrollView(.horizontal) {
                VStack(alignment: .leading, spacing: 5) {
                    headerRow
                    countRow
                    if let recommendation = model.recommendation {
                        ForEach(recommendation.entries) { entry in
                            numberRow(
                                label: "推荐 \(entry.sequence)",
                                red: Set(entry.redNumbers.lotteryNumbers),
                                blue: Set(entry.blueNumbers.lotteryNumbers),
                                style: .recommendation
                            )
                        }
                    }
                    ForEach(model.draws) { draw in
                        numberRow(
                            label: draw.issue,
                            red: Set(draw.redNumbers.lotteryNumbers),
                            blue: Set(draw.blueNumbers.lotteryNumbers),
                            style: .draw
                        )
                    }
                }
                .padding(12)
            }
            .background(Color(.secondarySystemGroupedBackground), in: .rect(cornerRadius: 20))
            .scrollIndicators(.visible)
            .accessibilityLabel("近期开奖结果和推荐号码矩阵")
        }
    }

    private var headerRow: some View {
        HStack(spacing: 4) {
            matrixLabel("期号")
            ForEach(rules.redMin...rules.redMax, id: \.self) { number in
                matrixCell(String(format: "%02d", number), foreground: LotteryPalette.red)
            }
            separator
            ForEach(rules.blueMin...rules.blueMax, id: \.self) { number in
                matrixCell(String(format: "%02d", number), foreground: LotteryPalette.blue)
            }
        }
    }

    private var countRow: some View {
        HStack(spacing: 4) {
            matrixLabel("出现次数")
            ForEach(rules.redMin...rules.redMax, id: \.self) { number in
                matrixCell("\(frequency.red[number, default: 0])", foreground: .secondary)
            }
            separator
            ForEach(rules.blueMin...rules.blueMax, id: \.self) { number in
                matrixCell("\(frequency.blue[number, default: 0])", foreground: .secondary)
            }
        }
    }

    private func numberRow(label: String, red: Set<Int>, blue: Set<Int>, style: TrendRowStyle) -> some View {
        HStack(spacing: 4) {
            matrixLabel(label)
            ForEach(rules.redMin...rules.redMax, id: \.self) { number in
                selectionCell(number: number, selected: red.contains(number), color: LotteryPalette.red, style: style)
            }
            separator
            ForEach(rules.blueMin...rules.blueMax, id: \.self) { number in
                selectionCell(number: number, selected: blue.contains(number), color: LotteryPalette.blue, style: style)
            }
        }
    }

    private func selectionCell(number: Int, selected: Bool, color: Color, style: TrendRowStyle) -> some View {
        Text(selected ? String(format: "%02d", number) : "·")
            .font(.caption2.monospacedDigit().weight(selected ? .bold : .regular))
            .foregroundStyle(selected ? (style == .draw ? Color.white : color) : Color.secondary.opacity(0.42))
            .frame(width: 27, height: 27)
            .background {
                if selected {
                    if style == .draw {
                        color.clipShape(.circle)
                    } else {
                        color.opacity(0.12).clipShape(.circle)
                    }
                }
            }
            .overlay {
                if selected, style == .recommendation {
                    Circle().stroke(color, lineWidth: 1.5)
                }
            }
    }

    private func matrixCell(_ text: String, foreground: Color) -> some View {
        Text(text)
            .font(.caption2.monospacedDigit())
            .foregroundStyle(foreground)
            .frame(width: 27, height: 27)
    }

    private func matrixLabel(_ text: String) -> some View {
        Text(text)
            .font(.caption2.monospacedDigit().weight(.semibold))
            .lineLimit(1)
            .frame(width: 72, alignment: .leading)
    }

    private var separator: some View {
        Divider().frame(width: 8, height: 24)
    }
}

private enum TrendRowStyle {
    case draw
    case recommendation
}
