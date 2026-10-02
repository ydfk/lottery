import SwiftUI
import UIKit

struct StealthRecommendationView: View {
    @Environment(\.dismiss) private var dismiss
    @Environment(\.scenePhase) private var scenePhase
    let recommendation: Recommendation
    @State private var exitCount = 0
    @State private var showsPurchase = false
    @State private var purchaseSaved = false
    @State private var purchaseDeadline: TimeInterval?

    var body: some View {
        VStack(spacing: 18) {
            Button(action: registerExitTap) {
                HStack(spacing: 7) {
                    ForEach(0..<5, id: \.self) { index in
                        Circle().fill(index < exitCount ? Color.secondary : Color(.systemGray5)).frame(width: 7, height: 7)
                    }
                }
                .frame(height: 44)
            }
            .accessibilityLabel("退出隐览")
            .accessibilityHint("连续点按五次退出")
            .accessibilityAction(.escape) { dismiss() }

            ScrollView {
                VStack(spacing: 0) {
                    ForEach(recommendation.entries) { entry in
                        NumberBalls(redNumbers: entry.redNumbers, blueNumbers: entry.blueNumbers)
                            .frame(maxWidth: .infinity)
                            .padding(.vertical, 20)
                            .background(Color(.systemBackground))
                            .contentShape(.rect)
                            .onTapGesture(count: 2, perform: openPurchase)
                        Divider()
                    }
                }
                .overlay(RoundedRectangle(cornerRadius: 2).stroke(Color(.separator)))
                .padding(.horizontal, 12)
            }
        }
        .background(Color(.systemBackground).ignoresSafeArea())
        .statusBarHidden()
        .onChange(of: scenePhase) { _, phase in
            if phase != .active { purchaseDeadline = nil }
        }
        .sheet(isPresented: $showsPurchase) {
            StealthPurchaseSheet(recommendation: recommendation) {
                purchaseSaved = true
                UINotificationFeedbackGenerator().notificationOccurred(.success)
            }
        }
    }

    private func openPurchase() {
        let deadline = purchaseDeadline
        purchaseDeadline = nil
        guard !purchaseSaved, !showsPurchase,
              let deadline, ProcessInfo.processInfo.systemUptime <= deadline else { return }
        exitCount = 0
        UIImpactFeedbackGenerator(style: .soft).impactOccurred()
        showsPurchase = true
    }

    private func registerExitTap() {
        // 只有单次顶部点按才启用短暂窗口，连续点按继续用于退出。
        purchaseDeadline = exitCount == 0 && !purchaseSaved
            ? ProcessInfo.processInfo.systemUptime + 2 : nil
        exitCount += 1
        UIImpactFeedbackGenerator(style: .soft).impactOccurred()
        if exitCount >= 5 { dismiss() }
        else {
            Task {
                try? await Task.sleep(for: .seconds(1.8))
                if exitCount < 5 { exitCount = 0 }
            }
        }
    }
}
