import SwiftUI

struct QuickPurchaseSheet: View {
    @Environment(AppSession.self) private var session
    @Environment(\.dismiss) private var dismiss
    let recommendation: Recommendation
    var onFullEntry: ((TicketDraft) -> Void)? = nil
    let onSaved: () -> Void
    @State private var multiple = 1
    @State private var additional = false
    @State private var isSaving = false
    @State private var errorMessage: String?

    private var draft: TicketDraft {
        .purchase(recommendation, multiple: multiple, additional: additional)
    }

    var body: some View {
        NavigationStack {
            Form {
                Section {
                    LabeledContent(draft.lottery.name, value: "第 \(recommendation.issue) 期")
                    Stepper("\(multiple) 倍", value: $multiple, in: 1...99)
                        .accessibilityLabel("投注倍数")
                        .accessibilityValue("\(multiple) 倍")
                    if draft.lottery == .dlt {
                        Toggle("追加投注", isOn: $additional)
                            .accessibilityHint("开启后每注每倍增加 1 元")
                    }
                    LabeledContent("合计 \(draft.entries.count) 注", value: LotteryFormatters.currency(draft.calculatedCost))
                } footer: {
                    Text("使用当前推荐的全部号码，仅记录购买信息。")
                }
                if let onFullEntry {
                    Section {
                        Button("完整录入", systemImage: "square.and.pencil") {
                            onFullEntry(draft)
                            dismiss()
                        }
                    } footer: {
                        Text("可修改号码、日期、备注及照片。")
                    }
                }
                if recommendation.isPurchased == true {
                    Section { Text("这条推荐已有购买记录，本次确认将新增一条。").foregroundStyle(.secondary) }
                }
                if let errorMessage {
                    Section { ErrorBanner(message: errorMessage) }
                }
                Section {
                    Button(isSaving ? "保存中…" : "确认记录购买") { save() }
                        .frame(maxWidth: .infinity, minHeight: 44)
                        .disabled(isSaving || !draft.isValid || session.api == nil)
                        .accessibilityIdentifier("stealth-confirm-purchase")
                }
            }
            .disabled(isSaving)
            .navigationTitle("快捷购买")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("取消") { dismiss() }.disabled(isSaving)
                }
            }
        }
        .presentationDetents([.medium, .large])
        .interactiveDismissDisabled(isSaving)
    }

    private func save() {
        guard !isSaving, draft.isValid, let api = session.api else { return }
        isSaving = true
        errorMessage = nil
        let purchase = draft
        Task {
            defer { isSaving = false }
            do {
                _ = try await api.createTicket(TicketPayload(
                    lotteryCode: purchase.lottery.rawValue,
                    uploadId: nil,
                    recommendationId: purchase.recommendationId,
                    issue: purchase.issue,
                    drawDate: LotteryFormatters.dateOnly.string(from: purchase.drawDate),
                    purchasedAt: ISO8601DateFormatter().string(from: purchase.purchasedAt),
                    costAmount: purchase.calculatedCost,
                    notes: "",
                    entries: purchase.entries.map { $0.payload() }
                ))
                session.purchaseDidSave()
                onSaved()
                dismiss()
            } catch {
                errorMessage = error.localizedDescription
            }
        }
    }
}
