import Observation
import SwiftUI

@MainActor
@Observable
private final class HistoryModel {
    var items: [Ticket] = []
    var lottery = ""
    var status = ""
    var sort = "latest"
    var page = 1
    var total = 0
    var hasMore = false
    var isLoading = false
    var isMutating = false
    var errorMessage: String?

    func load(using api: APIClient?, reset: Bool) async {
        guard let api, !isLoading else { return }
        isLoading = true
        defer { isLoading = false }
        do {
            let targetPage = reset ? 1 : page + 1
            let result = try await api.tickets(page: targetPage, lottery: lottery, status: status, sort: sort)
            items = reset ? result.items : items + result.items
            page = result.page
            total = result.total
            hasMore = result.hasMore
            errorMessage = nil
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func recheck(_ ticket: Ticket, using api: APIClient?) async -> Ticket? {
        guard let api, !isMutating else { return nil }
        isMutating = true
        defer { isMutating = false }
        do {
            let updated = try await api.recheckTicket(id: ticket.id)
            replace(updated)
            return updated
        } catch {
            errorMessage = error.localizedDescription
            return nil
        }
    }

    func delete(_ ticket: Ticket, using api: APIClient?) async -> Bool {
        guard let api, !isMutating else { return false }
        isMutating = true
        defer { isMutating = false }
        do {
            _ = try await api.deleteTicket(id: ticket.id)
            items.removeAll { $0.id == ticket.id }
            total = max(0, total - 1)
            return true
        } catch {
            errorMessage = error.localizedDescription
            return false
        }
    }

    private func replace(_ ticket: Ticket) {
        guard let index = items.firstIndex(where: { $0.id == ticket.id }) else { return }
        items[index] = ticket
    }
}

struct HistoryView: View {
    @Environment(AppSession.self) private var session
    @State private var model = HistoryModel()
    @State private var selected: Ticket?
    @State private var deleteCandidate: Ticket?

    var body: some View {
        List {
            Section {
                Picker("彩种", selection: $model.lottery) {
                    Text("全部").tag("")
                    ForEach(LotteryKind.allCases) { kind in Text(kind.name).tag(kind.rawValue) }
                }
                .pickerStyle(.segmented)
                HStack {
                    Menu {
                        Picker("状态", selection: $model.status) {
                            Text("全部状态").tag("")
                            Text("待开奖").tag("pending")
                            Text("已中奖").tag("won")
                            Text("未中奖").tag("not_won")
                        }
                    } label: { Label("状态", systemImage: "line.3.horizontal.decrease") }
                    Spacer()
                    Menu {
                        Picker("排序", selection: $model.sort) {
                            Text("最新购买").tag("latest")
                            Text("最早购买").tag("oldest")
                            Text("奖金最高").tag("prize_high")
                            Text("花费最高").tag("cost_high")
                        }
                    } label: { Label("排序", systemImage: "arrow.up.arrow.down") }
                }
            }
            if let errorMessage = model.errorMessage {
                ErrorBanner(message: errorMessage).listRowBackground(Color.clear)
            }
            Section("\(model.total) 条记录") {
                if model.items.isEmpty, !model.isLoading {
                    EmptyState(icon: "ticket", title: "暂无票据", message: "从记录页添加第一张彩票。")
                        .listRowBackground(Color.clear)
                }
                ForEach(model.items) { ticket in
                    Button { selected = ticket } label: { TicketRow(ticket: ticket) }
                        .buttonStyle(.plain)
                        .swipeActions(edge: .trailing) {
                            Button("删除", systemImage: "trash", role: .destructive) { deleteCandidate = ticket }
                            Button("编辑", systemImage: "pencil") { session.edit(ticket) }.tint(.blue)
                        }
                        .swipeActions(edge: .leading) {
                            Button("重判", systemImage: "arrow.clockwise") {
                                Task { _ = await model.recheck(ticket, using: session.api) }
                            }
                            .tint(LotteryPalette.amber)
                        }
                        .onAppear {
                            if ticket.id == model.items.last?.id, model.hasMore {
                                Task { await model.load(using: session.api, reset: false) }
                            }
                        }
                }
                if model.isLoading { HStack { Spacer(); ProgressView(); Spacer() } }
            }
        }
        .navigationTitle("历史")
        .refreshable { await model.load(using: session.api, reset: true) }
        .task { await model.load(using: session.api, reset: true) }
        .onChange(of: model.lottery) { _, _ in reload() }
        .onChange(of: model.status) { _, _ in reload() }
        .onChange(of: model.sort) { _, _ in reload() }
        .onChange(of: session.selectedTab) { _, tab in if tab == .history { reload() } }
        .sheet(item: $selected) { ticketDetail($0) }
        .confirmationDialog(
            "删除这条票据记录？",
            isPresented: Binding(get: { deleteCandidate != nil }, set: { if !$0 { deleteCandidate = nil } }),
            titleVisibility: .visible
        ) {
            Button("删除票据和关联图片", role: .destructive) {
                guard let ticket = deleteCandidate else { return }
                Task {
                    if await model.delete(ticket, using: session.api) { selected = nil }
                    deleteCandidate = nil
                }
            }
        } message: {
            Text("号码、上传记录和未被其他记录使用的原图会一并删除。")
        }
    }

    private func ticketDetail(_ ticket: Ticket) -> some View {
        NavigationStack {
            List {
                Section {
                    HStack {
                        Text(LotteryKind(rawValue: ticket.lotteryCode)?.name ?? ticket.lotteryCode)
                        Text("第 \(ticket.issue) 期").foregroundStyle(.secondary)
                        Spacer()
                        StatusPill(status: ticket.status)
                    }
                    if !ticket.drawRedNumbers.isEmpty {
                        NumberBalls(redNumbers: ticket.drawRedNumbers, blueNumbers: ticket.drawBlueNumbers)
                    }
                }
                Section("日期") {
                    LabeledContent("生成时间", value: LotteryFormatters.displayDateTime(ticket.createdAt))
                    LabeledContent("购买时间", value: LotteryFormatters.displayDateTime(ticket.purchasedAt))
                    LabeledContent(
                        "开奖日期",
                        value: LotteryFormatters.displayDate(ticket.drawDate ?? ticket.manualDrawDate)
                    )
                }
                Section("购买号码") {
                    ForEach(ticket.entries) { entry in
                        VStack(alignment: .leading, spacing: 8) {
                            NumberBalls(redNumbers: entry.redNumbers, blueNumbers: entry.blueNumbers, compact: true)
                            HStack {
                                Text("\(entry.multiple) 倍\(entry.isAdditional ? " · 追加" : "")")
                                Spacer()
                                if !entry.prizeName.isEmpty { Text(entry.prizeName).foregroundStyle(LotteryPalette.jade) }
                            }
                            .font(.caption)
                            .foregroundStyle(.secondary)
                        }
                        .padding(.vertical, 4)
                    }
                }
                Section("金额") {
                    LabeledContent("花费", value: LotteryFormatters.currency(ticket.costAmount))
                    LabeledContent("中奖", value: LotteryFormatters.currency(ticket.prizeAmount))
                }
                if let recommendation = ticket.recommendation {
                    recommendationSection(recommendation)
                } else if ticket.recommendationId != nil {
                    Section("关联推荐") {
                        Label("关联推荐暂不可用", systemImage: "link.badge.plus")
                            .foregroundStyle(.secondary)
                    }
                }
                if !ticket.imageUrl.isEmpty {
                    Section("票据原图") { RemoteTicketImage(path: ticket.imageUrl) }
                }
                if !ticket.notes.isEmpty { Section("备注") { Text(ticket.notes) } }
            }
            .navigationTitle("票据详情")
            .toolbar {
                ToolbarItemGroup(placement: .bottomBar) {
                    Button("编辑", systemImage: "pencil") { selected = nil; session.edit(ticket) }
                    Button("重判", systemImage: "arrow.clockwise") {
                        Task { if let updated = await model.recheck(ticket, using: session.api) { selected = updated } }
                    }
                    Button("删除", systemImage: "trash", role: .destructive) { deleteCandidate = ticket }
                }
            }
        }
        .presentationDetents([.medium, .large])
    }

    private func recommendationSection(_ recommendation: TicketRecommendation) -> some View {
        Section("关联推荐") {
            HStack {
                Label("已关联推荐", systemImage: "link.circle.fill")
                    .font(.subheadline.weight(.semibold))
                    .foregroundStyle(LotteryPalette.blue)
                Spacer()
                Text("第 \(recommendation.issue) 期").foregroundStyle(.secondary)
            }
            LotteryDateStrip(createdAt: recommendation.createdAt, drawDate: recommendation.drawDate)
            if !recommendation.summary.isEmpty {
                Text(recommendation.summary).font(.subheadline)
            }
            ForEach(recommendation.entries) { entry in
                VStack(alignment: .leading, spacing: 7) {
                    HStack {
                        Text("推荐 \(entry.sequence)").font(.caption.weight(.semibold))
                        Spacer()
                        if !entry.prizeName.isEmpty {
                            Text(entry.prizeName).foregroundStyle(LotteryPalette.jade)
                        }
                    }
                    NumberBalls(
                        redNumbers: entry.redNumbers,
                        blueNumbers: entry.blueNumbers,
                        compact: true
                    )
                }
                .padding(.vertical, 3)
            }
        }
    }

    private func reload() {
        Task { await model.load(using: session.api, reset: true) }
    }
}

private struct TicketRow: View {
    let ticket: Ticket

    var body: some View {
        VStack(alignment: .leading, spacing: 11) {
            HStack {
                Text(LotteryKind(rawValue: ticket.lotteryCode)?.name ?? ticket.lotteryCode).font(.headline)
                Text("第 \(ticket.issue) 期").foregroundStyle(.secondary)
                Spacer()
                StatusPill(status: ticket.status)
            }
            LotteryDateStrip(
                createdAt: ticket.createdAt,
                drawDate: ticket.drawDate ?? ticket.manualDrawDate
            )
            if ticket.recommendation != nil || ticket.recommendationId != nil {
                Label("关联推荐", systemImage: "link.circle.fill")
                    .font(.caption.weight(.semibold))
                    .foregroundStyle(LotteryPalette.blue)
                    .padding(.horizontal, 9)
                    .padding(.vertical, 5)
                    .background(LotteryPalette.blue.opacity(0.1), in: .capsule)
            }
            if !ticket.drawRedNumbers.isEmpty {
                NumberBalls(redNumbers: ticket.drawRedNumbers, blueNumbers: ticket.drawBlueNumbers, compact: true)
            } else if let first = ticket.entries.first {
                NumberBalls(redNumbers: first.redNumbers, blueNumbers: first.blueNumbers, compact: true)
            }
            HStack {
                Text("花费 \(LotteryFormatters.currency(ticket.costAmount))")
                Spacer()
                Text("中奖 \(LotteryFormatters.currency(ticket.prizeAmount))")
                    .foregroundStyle(ticket.prizeAmount > 0 ? LotteryPalette.jade : .secondary)
            }
            .font(.caption)
            .foregroundStyle(.secondary)
        }
        .padding(.vertical, 5)
    }
}

private struct RemoteTicketImage: View {
    @Environment(AppSession.self) private var session
    let path: String
    @State private var url: URL?

    var body: some View {
        AsyncImage(url: url) { phase in
            switch phase {
            case let .success(image): image.resizable().scaledToFit().clipShape(.rect(cornerRadius: 16))
            case .failure: ContentUnavailableView("图片加载失败", systemImage: "photo.badge.exclamationmark")
            default: HStack { Spacer(); ProgressView(); Spacer() }.frame(height: 160)
            }
        }
        .task { url = await session.api?.absoluteImageURL(path) }
    }
}
