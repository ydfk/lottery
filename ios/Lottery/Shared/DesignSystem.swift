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
        HStack(spacing: compact ? 5 : 7) {
            ForEach(Array(redNumbers.lotteryNumbers.enumerated()), id: \.offset) { _, number in
                ball(number, color: LotteryPalette.red)
            }
            ForEach(Array(blueNumbers.lotteryNumbers.enumerated()), id: \.offset) { _, number in
                ball(number, color: LotteryPalette.blue)
            }
        }
        .accessibilityElement(children: .ignore)
        .accessibilityLabel("红球 \(redNumbers)，蓝球 \(blueNumbers)")
    }

    private func ball(_ number: Int, color: Color) -> some View {
        Text(String(format: "%02d", number))
            .font(.system(size: compact ? 11 : 14, weight: .bold, design: .rounded))
            .foregroundStyle(.white)
            .frame(width: compact ? 25 : 32, height: compact ? 25 : 32)
            .background(color, in: .circle)
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

