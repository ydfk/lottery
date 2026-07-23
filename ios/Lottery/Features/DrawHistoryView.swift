import Observation
import SwiftUI

@MainActor
@Observable
private final class DrawHistoryModel {
    var items: [DrawResult] = []
    var lottery = ""
    var issue = ""
    var isLoading = false
    var isSyncing = false
    var hasMore = false
    var page = 1
    var errorMessage: String?

    func load(using api: APIClient?, reset: Bool) async {
        guard let api, !isLoading else { return }
        isLoading = true
        defer { isLoading = false }
        do {
            let targetPage = reset ? 1 : page + 1
            let result = try await api.draws(page: targetPage, lottery: lottery, issue: issue)
            items = reset ? result.items : items + result.items
            page = result.page
            hasMore = result.hasMore
            errorMessage = nil
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func sync(using api: APIClient?) async {
        guard let api, let kind = LotteryKind(rawValue: lottery), !issue.isEmpty else {
            errorMessage = "同步前请选择彩种并填写期号。"
            return
        }
        isSyncing = true
        defer { isSyncing = false }
        do {
            _ = try await api.syncDraw(code: kind.rawValue, issue: issue)
            await load(using: api, reset: true)
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}

struct DrawHistoryView: View {
    @Environment(AppSession.self) private var session
    @State private var model = DrawHistoryModel()
    @State private var selectedDraw: DrawResult?
    @State private var showsSync = false

    var body: some View {
        List {
            Section {
                Picker("彩种", selection: $model.lottery) {
                    Text("全部彩种").tag("")
                    ForEach(LotteryKind.allCases) { kind in Text(kind.name).tag(kind.rawValue) }
                }
                if let errorMessage = model.errorMessage { ErrorBanner(message: errorMessage) }
            }
            Section("开奖记录") {
                if model.items.isEmpty, !model.isLoading {
                    EmptyState(icon: "list.number", title: "暂无开奖记录", message: "可同步指定期号后再查看。")
                        .listRowBackground(Color.clear)
                }
                ForEach(model.items) { draw in
                    Button { selectedDraw = draw } label: { drawRow(draw) }
                        .buttonStyle(.plain)
                        .onAppear {
                            if draw.id == model.items.last?.id, model.hasMore {
                                Task { await model.load(using: session.api, reset: false) }
                            }
                        }
                }
                if model.isLoading { HStack { Spacer(); ProgressView(); Spacer() } }
            }
        }
        .navigationTitle("开奖历史")
        .toolbar {
            Button("同步", systemImage: "arrow.triangle.2.circlepath") { showsSync = true }
        }
        .onChange(of: model.lottery) { _, _ in Task { await model.load(using: session.api, reset: true) } }
        .refreshable { await model.load(using: session.api, reset: true) }
        .task { await model.load(using: session.api, reset: true) }
        .sheet(item: $selectedDraw) { DrawDetailView(draw: $0) }
        .sheet(isPresented: $showsSync) { syncSheet }
    }

    private func drawRow(_ draw: DrawResult) -> some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack {
                Text(LotteryKind(rawValue: draw.lotteryCode)?.name ?? draw.lotteryCode).font(.headline)
                Text("第 \(draw.issue) 期").foregroundStyle(.secondary)
                Spacer()
                Image(systemName: "chevron.right").foregroundStyle(.tertiary)
            }
            NumberBalls(redNumbers: draw.redNumbers, blueNumbers: draw.blueNumbers, compact: true)
        }
        .padding(.vertical, 6)
    }

    private var syncSheet: some View {
        NavigationStack {
            Form {
                Picker("彩票类型", selection: $model.lottery) {
                    Text("请选择").tag("")
                    ForEach(LotteryKind.allCases) { kind in Text(kind.name).tag(kind.rawValue) }
                }
                TextField("期号，例如 2026048", text: $model.issue)
                    .keyboardType(.numberPad)
                if let errorMessage = model.errorMessage { ErrorBanner(message: errorMessage) }
            }
            .navigationTitle("同步开奖")
            .toolbar {
                ToolbarItem(placement: .cancellationAction) { Button("取消") { showsSync = false } }
                ToolbarItem(placement: .confirmationAction) {
                    Button(model.isSyncing ? "同步中" : "同步") {
                        Task {
                            await model.sync(using: session.api)
                            if model.errorMessage == nil { showsSync = false }
                        }
                    }
                    .disabled(model.isSyncing)
                }
            }
        }
        .presentationDetents([.medium])
    }
}

private struct DrawDetailView: View {
    @Environment(\.dismiss) private var dismiss
    let draw: DrawResult

    var body: some View {
        NavigationStack {
            List {
                Section("开奖号码") {
                    NumberBalls(redNumbers: draw.redNumbers, blueNumbers: draw.blueNumbers)
                        .padding(.vertical, 8)
                }
                Section("开奖信息") {
                    LabeledContent("彩种", value: LotteryKind(rawValue: draw.lotteryCode)?.name ?? draw.lotteryCode)
                    LabeledContent("期号", value: draw.issue)
                    LabeledContent("开奖日期", value: String(draw.drawDate.prefix(10)))
                    LabeledContent("奖池", value: LotteryFormatters.currency(draw.prizePoolAmount))
                }
                Section("奖级") {
                    if draw.prizeDetails.isEmpty { Text("暂无奖级明细").foregroundStyle(.secondary) }
                    ForEach(draw.prizeDetails) { prize in
                        LabeledContent(prize.prizeName, value: "\(prize.winnerCount) 注 · \(LotteryFormatters.currency(prize.singleBonus))")
                    }
                }
            }
            .navigationTitle("第 \(draw.issue) 期")
            .toolbar { Button("完成") { dismiss() } }
        }
    }
}

