import Observation
import SwiftUI

private enum DrawDisplayMode: String, CaseIterable, Identifiable {
    case trend = "号码走势"
    case list = "开奖列表"

    var id: String { rawValue }
}

@MainActor
@Observable
private final class DrawResultsModel {
    var items: [DrawResult] = []
    var query = DrawQuery()
    var total = 0
    var hasMore = false
    var isLoading = false
    var syncingIDs: Set<String> = []
    var syncLottery = LotteryKind.ssq
    var syncIssue = ""
    var errorMessage: String?

    func load(using api: APIClient?, reset: Bool) async {
        guard let api, !isLoading else { return }
        isLoading = true
        defer { isLoading = false }
        do {
            var nextQuery = query
            nextQuery.page = reset ? 1 : query.page + 1
            let result = try await api.draws(nextQuery)
            items = reset ? result.items : merge(items, with: result.items)
            query.page = result.page
            total = result.total
            hasMore = result.hasMore
            errorMessage = nil
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func syncSpecified(using api: APIClient?) async -> Bool {
        guard !syncIssue.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else {
            errorMessage = "请填写要同步的期号。"
            return false
        }
        return await sync(code: syncLottery.rawValue, issue: syncIssue, id: "specified", using: api)
    }

    func complete(_ draw: DrawResult, using api: APIClient?) async -> DrawResult? {
        guard await sync(code: draw.lotteryCode, issue: draw.issue, id: draw.id, using: api) else {
            return nil
        }
        return items.first { $0.id == draw.id }
    }

    private func sync(code: String, issue: String, id: String, using api: APIClient?) async -> Bool {
        guard let api, !syncingIDs.contains(id) else { return false }
        syncingIDs.insert(id)
        defer { syncingIDs.remove(id) }
        do {
            let result = try await api.syncDraw(code: code, issue: issue)
            guard result.syncedCount > 0 else {
                errorMessage = "第三方接口暂未返回这一期的完整开奖信息。"
                return false
            }
            await load(using: api, reset: true)
            return true
        } catch {
            errorMessage = error.localizedDescription
            return false
        }
    }

    private func merge(_ current: [DrawResult], with next: [DrawResult]) -> [DrawResult] {
        var ids = Set(current.map(\.id))
        return current + next.filter { ids.insert($0.id).inserted }
    }
}

struct DrawResultsView: View {
    @Environment(AppSession.self) private var session
    @State private var model = DrawResultsModel()
    @State private var mode = DrawDisplayMode.trend
    @State private var selectedDraw: DrawResult?
    @State private var showsFilters = false
    @State private var showsSync = false
    @State private var trendRefreshID = UUID()

    var body: some View {
        VStack(spacing: 0) {
            Picker("开奖视图", selection: $mode) {
                ForEach(DrawDisplayMode.allCases) { item in Text(item.rawValue).tag(item) }
            }
            .pickerStyle(.segmented)
            .padding(.horizontal)
            .padding(.bottom, 8)

            switch mode {
            case .trend:
                DrawTrendView(refreshID: trendRefreshID)
            case .list:
                drawList
            }
        }
        .navigationTitle("开奖")
        .toolbar {
            if mode == .list {
                Button("筛选", systemImage: "line.3.horizontal.decrease") { showsFilters = true }
            }
            Button("同步", systemImage: "arrow.triangle.2.circlepath") { showsSync = true }
        }
        .task { await model.load(using: session.api, reset: true) }
        .sheet(item: $selectedDraw) { draw in
            DrawDetailView(draw: draw) { current in
                let updated = await model.complete(current, using: session.api)
                if updated != nil { trendRefreshID = UUID() }
                return updated
            }
        }
        .sheet(isPresented: $showsFilters) {
            DrawFilterSheet(query: model.query) { query in
                model.query = query
                Task { await model.load(using: session.api, reset: true) }
            }
        }
        .sheet(isPresented: $showsSync) { syncSheet }
    }

    private var drawList: some View {
        List {
            Section {
                HStack(spacing: 8) {
                    if !model.query.lotteryCode.isEmpty {
                        filterPill(LotteryKind(rawValue: model.query.lotteryCode)?.name ?? model.query.lotteryCode)
                    }
                    if !model.query.issue.isEmpty { filterPill("第 \(model.query.issue) 期") }
                    if !model.query.drawDate.isEmpty { filterPill(model.query.drawDate) }
                    filterPill(model.query.sort == "oldest" ? "最早优先" : "最新优先")
                    Spacer(minLength: 0)
                }
                .font(.caption)
            }

            if let errorMessage = model.errorMessage {
                ErrorBanner(message: errorMessage).listRowBackground(Color.clear)
            }

            Section("\(model.total) 期开奖") {
                if model.items.isEmpty, !model.isLoading {
                    EmptyState(icon: "list.number", title: "暂无开奖记录", message: "调整筛选条件或同步指定期号。")
                        .listRowBackground(Color.clear)
                }
                ForEach(model.items) { draw in
                    HStack(spacing: 10) {
                        Button { selectedDraw = draw } label: {
                            DrawRow(draw: draw)
                                .frame(maxWidth: .infinity, alignment: .leading)
                        }
                        .buttonStyle(.plain)

                        Button("补全", systemImage: "arrow.clockwise") {
                            complete(draw)
                        }
                        .labelStyle(.iconOnly)
                        .disabled(model.syncingIDs.contains(draw.id))
                        .accessibilityHint("从数据源重新同步这一期")
                    }
                    .onAppear {
                        if draw.id == model.items.last?.id, model.hasMore {
                            Task { await model.load(using: session.api, reset: false) }
                        }
                    }
                }
                if model.isLoading { HStack { Spacer(); ProgressView(); Spacer() } }
            }
        }
        .refreshable { await model.load(using: session.api, reset: true) }
    }

    private func filterPill(_ text: String) -> some View {
        Text(text)
            .padding(.horizontal, 9)
            .padding(.vertical, 5)
            .background(Color(.secondarySystemGroupedBackground), in: .capsule)
    }

    private func complete(_ draw: DrawResult) {
        Task {
            if let updated = await model.complete(draw, using: session.api) {
                if selectedDraw?.id == draw.id { selectedDraw = updated }
                trendRefreshID = UUID()
                session.message = "第 \(draw.issue) 期开奖信息已补全。"
            }
        }
    }

    private var syncSheet: some View {
        NavigationStack {
            Form {
                Picker("彩票类型", selection: Binding(
                    get: { model.syncLottery },
                    set: { model.syncLottery = $0 }
                )) {
                    ForEach(LotteryKind.allCases) { kind in Text(kind.name).tag(kind) }
                }
                TextField("期号，例如 2026048", text: Binding(
                    get: { model.syncIssue },
                    set: { model.syncIssue = $0 }
                ))
                .keyboardType(.numberPad)
                if let errorMessage = model.errorMessage { ErrorBanner(message: errorMessage) }
            }
            .navigationTitle("同步指定期开奖")
            .toolbar {
                ToolbarItem(placement: .cancellationAction) { Button("取消") { showsSync = false } }
                ToolbarItem(placement: .confirmationAction) {
                    Button(model.syncingIDs.contains("specified") ? "同步中" : "同步") {
                        Task {
                            if await model.syncSpecified(using: session.api) {
                                model.query.lotteryCode = model.syncLottery.rawValue
                                model.query.issue = model.syncIssue
                                await model.load(using: session.api, reset: true)
                                trendRefreshID = UUID()
                                showsSync = false
                            }
                        }
                    }
                    .disabled(model.syncingIDs.contains("specified"))
                }
            }
        }
        .presentationDetents([.medium])
    }
}

private struct DrawRow: View {
    let draw: DrawResult

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack {
                Text(LotteryKind(rawValue: draw.lotteryCode)?.name ?? draw.lotteryCode)
                    .font(.headline)
                Text("第 \(draw.issue) 期").foregroundStyle(.secondary)
                Spacer()
                Text(String(draw.drawDate.prefix(10))).font(.caption).foregroundStyle(.secondary)
            }
            NumberBalls(redNumbers: draw.redNumbers, blueNumbers: draw.blueNumbers, compact: true)
        }
        .padding(.vertical, 5)
    }
}

private struct DrawFilterSheet: View {
    @Environment(\.dismiss) private var dismiss
    @State private var draft: DrawQuery
    @State private var usesDate: Bool
    @State private var date: Date
    let onApply: (DrawQuery) -> Void

    init(query: DrawQuery, onApply: @escaping (DrawQuery) -> Void) {
        _draft = State(initialValue: query)
        let parsedDate = LotteryFormatters.dateOnly.date(from: query.drawDate)
        _usesDate = State(initialValue: parsedDate != nil)
        _date = State(initialValue: parsedDate ?? Date())
        self.onApply = onApply
    }

    var body: some View {
        NavigationStack {
            Form {
                Picker("彩票类型", selection: $draft.lotteryCode) {
                    Text("全部彩种").tag("")
                    ForEach(LotteryKind.allCases) { kind in Text(kind.name).tag(kind.rawValue) }
                }
                TextField("期号", text: $draft.issue).keyboardType(.numberPad)
                Toggle("按开奖日期筛选", isOn: $usesDate)
                if usesDate { DatePicker("开奖日期", selection: $date, displayedComponents: .date) }
                Picker("排序", selection: $draft.sort) {
                    Text("最新优先").tag("latest")
                    Text("最早优先").tag("oldest")
                }
            }
            .navigationTitle("筛选开奖记录")
            .toolbar {
                ToolbarItem(placement: .cancellationAction) { Button("取消") { dismiss() } }
                ToolbarItem(placement: .confirmationAction) {
                    Button("应用") {
                        draft.page = 1
                        draft.pageSize = 20
                        draft.drawDate = usesDate ? LotteryFormatters.dateOnly.string(from: date) : ""
                        onApply(draft)
                        dismiss()
                    }
                }
            }
        }
        .presentationDetents([.medium, .large])
    }
}

private struct DrawDetailView: View {
    @Environment(\.dismiss) private var dismiss
    @State private var draw: DrawResult
    @State private var isCompleting = false
    @State private var errorMessage: String?
    let onComplete: (DrawResult) async -> DrawResult?

    init(draw: DrawResult, onComplete: @escaping (DrawResult) async -> DrawResult?) {
        _draw = State(initialValue: draw)
        self.onComplete = onComplete
    }

    var body: some View {
        NavigationStack {
            List {
                if let errorMessage { ErrorBanner(message: errorMessage).listRowBackground(Color.clear) }
                Section("开奖号码") {
                    NumberBalls(redNumbers: draw.redNumbers, blueNumbers: draw.blueNumbers)
                        .padding(.vertical, 8)
                }
                Section("开奖信息") {
                    LabeledContent("彩种", value: LotteryKind(rawValue: draw.lotteryCode)?.name ?? draw.lotteryCode)
                    LabeledContent("期号", value: draw.issue)
                    LabeledContent("开奖日期", value: String(draw.drawDate.prefix(10)))
                    LabeledContent("一等奖", value: LotteryFormatters.currency(draw.firstPrizeAmount ?? 0))
                    LabeledContent("二等奖", value: LotteryFormatters.currency(draw.secondPrizeAmount ?? 0))
                    LabeledContent("销售额", value: LotteryFormatters.currency(draw.saleAmount))
                    LabeledContent("奖池", value: LotteryFormatters.currency(draw.prizePoolAmount))
                }
                Section("同步信息") {
                    LabeledContent("来源", value: draw.source.isEmpty ? "未知" : draw.source)
                    LabeledContent("首次同步", value: displayTime(draw.createdAt))
                    LabeledContent("更新时间", value: displayTime(draw.updatedAt))
                }
                Section("完整奖级明细") {
                    if draw.prizeDetails.isEmpty { Text("暂无奖级明细").foregroundStyle(.secondary) }
                    ForEach(draw.prizeDetails) { prize in
                        VStack(alignment: .leading, spacing: 5) {
                            LabeledContent(prize.prizeName, value: LotteryFormatters.currency(prize.singleBonus))
                            HStack {
                                if !prize.prizeRule.isEmpty { Text(prize.prizeRule) }
                                Spacer()
                                Text("\(prize.winnerCount) 注")
                            }
                            .font(.caption)
                            .foregroundStyle(.secondary)
                        }
                    }
                }
                if !draw.rawPayload.isEmpty {
                    Section {
                        DisclosureGroup("原始同步数据") {
                            ScrollView(.horizontal) {
                                Text(draw.rawPayload)
                                    .font(.caption.monospaced())
                                    .textSelection(.enabled)
                            }
                        }
                    } footer: {
                        Text("技术信息默认折叠，仅用于排查数据同步问题。")
                    }
                }
            }
            .navigationTitle("第 \(draw.issue) 期")
            .toolbar {
                ToolbarItem(placement: .cancellationAction) { Button("完成") { dismiss() } }
                ToolbarItem(placement: .confirmationAction) {
                    Button(isCompleting ? "补全中" : "补全", systemImage: "arrow.clockwise") {
                        Task {
                            isCompleting = true
                            defer { isCompleting = false }
                            if let updated = await onComplete(draw) { draw = updated }
                            else { errorMessage = "未能取得这一期的完整开奖信息。" }
                        }
                    }
                    .disabled(isCompleting)
                }
            }
        }
    }

    private func displayTime(_ value: String) -> String {
        String(value.prefix(19)).replacingOccurrences(of: "T", with: " ")
    }
}
