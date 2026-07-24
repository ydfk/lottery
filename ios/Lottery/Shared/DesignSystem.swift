import SwiftUI

enum LotteryPalette {
    static let ink = Color(red: 0.04, green: 0.07, blue: 0.12)
    static let porcelain = Color(red: 0.96, green: 0.97, blue: 0.99)
    static let red = Color(red: 0.90, green: 0.22, blue: 0.28)
    static let blue = Color(red: 0.18, green: 0.42, blue: 0.82)
    static let jade = Color(red: 0.08, green: 0.56, blue: 0.40)
    static let amber = Color(red: 0.84, green: 0.54, blue: 0.09)
}

struct AppBackground: View {
    var body: some View {
        LinearGradient(
            colors: [Color(.systemBackground), Color(.secondarySystemBackground).opacity(0.7)],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
        )
        .ignoresSafeArea()
    }
}

struct LotteryCard<Content: View>: View {
    @ViewBuilder let content: Content

    var body: some View {
        content
            .padding(16)
            .background(Color(.secondarySystemGroupedBackground), in: .rect(cornerRadius: 22))
    }
}

struct NumberBalls: View {
    let redNumbers: String
    let blueNumbers: String
    var compact = false

    var body: some View {
        LotteryBalls(
            redNumbers: redNumbers,
            blueNumbers: blueNumbers,
            compact: compact,
            match: nil
        )
        .accessibilityElement(children: .ignore)
        .accessibilityLabel("红球 \(redNumbers)，蓝球 \(blueNumbers)")
    }
}

struct HitNumberBalls: View {
    let redNumbers: String
    let blueNumbers: String
    let drawRedNumbers: String
    let drawBlueNumbers: String
    var compact = false

    private var match: LotteryNumberMatch {
        LotteryNumberMatch(
            redNumbers: redNumbers,
            blueNumbers: blueNumbers,
            drawRedNumbers: drawRedNumbers,
            drawBlueNumbers: drawBlueNumbers
        )
    }

    var body: some View {
        LotteryBalls(
            redNumbers: redNumbers,
            blueNumbers: blueNumbers,
            compact: compact,
            match: match.hasDrawNumbers ? match : nil
        )
        .accessibilityElement(children: .ignore)
        .accessibilityLabel(accessibilityText)
    }

    private var accessibilityText: String {
        guard match.hasDrawNumbers else { return "红球 \(redNumbers)，蓝球 \(blueNumbers)，待开奖" }
        return "红球 \(redNumbers)，命中 \(match.redHitCount) 个；蓝球 \(blueNumbers)，命中 \(match.blueHitCount) 个"
    }
}

struct LotteryEntryDisplay: Identifiable, Equatable, Sendable {
    let id: String
    let sequence: Int
    let redNumbers: String
    let blueNumbers: String
    var detail: String?

    static func recommendations(_ entries: [RecommendationEntry]) -> [LotteryEntryDisplay] {
        entries.map {
            LotteryEntryDisplay(
                id: $0.id,
                sequence: $0.sequence,
                redNumbers: $0.redNumbers,
                blueNumbers: $0.blueNumbers
            )
        }
    }

    static func tickets(_ entries: [TicketEntry]) -> [LotteryEntryDisplay] {
        entries.map {
            LotteryEntryDisplay(
                id: $0.id,
                sequence: $0.sequence,
                redNumbers: $0.redNumbers,
                blueNumbers: $0.blueNumbers,
                detail: "\($0.multiple) 倍\($0.isAdditional ? " · 追加" : "")"
            )
        }
    }
}

struct LotteryEntryNumbersList: View {
    let entries: [LotteryEntryDisplay]
    let drawRedNumbers: String
    let drawBlueNumbers: String

    var body: some View {
        if entries.isEmpty {
            Text("暂无号码")
                .font(.caption)
                .foregroundStyle(.secondary)
                .frame(maxWidth: .infinity, alignment: .leading)
        } else {
            VStack(alignment: .leading, spacing: 10) {
                ForEach(entries) { entry in
                    VStack(alignment: .leading, spacing: 5) {
                        HStack(spacing: 6) {
                            Text("第 \(entry.sequence) 注")
                                .font(.caption.weight(.semibold))
                            if let detail = entry.detail, !detail.isEmpty {
                                Text(detail)
                                    .font(.caption2)
                                    .foregroundStyle(.secondary)
                            }
                        }
                        HitNumberBalls(
                            redNumbers: entry.redNumbers,
                            blueNumbers: entry.blueNumbers,
                            drawRedNumbers: drawRedNumbers,
                            drawBlueNumbers: drawBlueNumbers,
                            compact: true
                        )
                    }
                    .accessibilityElement(children: .contain)
                }
            }
            .fixedSize(horizontal: false, vertical: true)
        }
    }
}

struct LotteryNumberMatch: Equatable, Sendable {
    let red: Set<Int>
    let blue: Set<Int>
    let drawRed: Set<Int>
    let drawBlue: Set<Int>

    init(redNumbers: String, blueNumbers: String, drawRedNumbers: String, drawBlueNumbers: String) {
        red = Set(redNumbers.lotteryNumbers)
        blue = Set(blueNumbers.lotteryNumbers)
        drawRed = Set(drawRedNumbers.lotteryNumbers)
        drawBlue = Set(drawBlueNumbers.lotteryNumbers)
    }

    var hasDrawNumbers: Bool { !drawRed.isEmpty || !drawBlue.isEmpty }
    var redHitCount: Int { red.intersection(drawRed).count }
    var blueHitCount: Int { blue.intersection(drawBlue).count }

    func isRedHit(_ number: Int) -> Bool { drawRed.contains(number) }
    func isBlueHit(_ number: Int) -> Bool { drawBlue.contains(number) }
}

private struct LotteryBalls: View {
    let redNumbers: String
    let blueNumbers: String
    let compact: Bool
    let match: LotteryNumberMatch?

    var body: some View {
        HStack(spacing: compact ? 5 : 7) {
            ForEach(Array(redNumbers.lotteryNumbers.enumerated()), id: \.offset) { _, number in
                LotteryBall(
                    number: number,
                    color: LotteryPalette.red,
                    compact: compact,
                    isActive: match?.isRedHit(number) ?? true
                )
            }
            ForEach(Array(blueNumbers.lotteryNumbers.enumerated()), id: \.offset) { _, number in
                LotteryBall(
                    number: number,
                    color: LotteryPalette.blue,
                    compact: compact,
                    isActive: match?.isBlueHit(number) ?? true
                )
            }
        }
    }
}

private struct LotteryBall: View {
    let number: Int
    let color: Color
    let compact: Bool
    let isActive: Bool

    var body: some View {
        Text(String(format: "%02d", number))
            .font(.system(size: compact ? 11 : 14, weight: .bold, design: .rounded))
            .foregroundStyle(isActive ? .white : color)
            .frame(width: compact ? 25 : 32, height: compact ? 25 : 32)
            .background(isActive ? color : Color(.secondarySystemGroupedBackground), in: .circle)
            .overlay {
                if !isActive {
                    Circle().stroke(color.opacity(0.38), lineWidth: 1.25)
                }
            }
    }
}

struct StatusPill: View {
    let status: String

    var body: some View {
        Text(label)
            .font(.caption.weight(.semibold))
            .foregroundStyle(color)
            .padding(.horizontal, 10)
            .padding(.vertical, 5)
            .background(color.opacity(0.12), in: .capsule)
    }

    private var label: String {
        switch status {
        case "won": return "已中奖"
        case "not_won": return "未中奖"
        case "checked": return "已开奖"
        default: return "待开奖"
        }
    }

    private var color: Color {
        switch status {
        case "won": return LotteryPalette.jade
        case "not_won", "checked": return .secondary
        default: return LotteryPalette.amber
        }
    }
}

struct LotteryDateStrip: View {
    let createdAt: String?
    let drawDate: String?
    var createdTitle = "生成"

    var body: some View {
        ViewThatFits(in: .horizontal) {
            HStack(spacing: 6) {
                dateLabel(createdTitle, value: LotteryFormatters.displayDateTime(createdAt))
                Text("·").foregroundStyle(.tertiary)
                dateLabel("开奖", value: LotteryFormatters.displayDate(drawDate))
            }
            VStack(alignment: .leading, spacing: 2) {
                dateLabel(createdTitle, value: LotteryFormatters.displayDateTime(createdAt))
                dateLabel("开奖", value: LotteryFormatters.displayDate(drawDate))
            }
        }
        .font(.caption2)
        .foregroundStyle(.secondary)
        .accessibilityElement(children: .combine)
    }

    private func dateLabel(_ title: String, value: String) -> some View {
        Text("\(title) \(value)").lineLimit(1)
    }
}

struct EmptyState: View {
    let icon: String
    let title: String
    let message: String

    var body: some View {
        ContentUnavailableView(title, systemImage: icon, description: Text(message))
    }
}

struct ErrorBanner: View {
    let message: String

    var body: some View {
        Label(message, systemImage: "exclamationmark.triangle.fill")
            .font(.footnote)
            .foregroundStyle(.red)
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(12)
            .background(Color.red.opacity(0.1), in: .rect(cornerRadius: 14))
            .accessibilityLabel("错误：\(message)")
    }
}

struct LoadingOverlay: View {
    let title: String

    var body: some View {
        HStack(spacing: 10) {
            ProgressView()
            Text(title).font(.subheadline)
        }
        .padding(.horizontal, 18)
        .padding(.vertical, 12)
        .glassEffect(.regular, in: .capsule)
    }
}
