import Observation
import SwiftUI
import UIKit

@MainActor
@Observable
private final class RecommendationsModel {
    var items: [Recommendation] = []
    var lottery = ""
    var status = ""
    var sort = "latest"
    var page = 1
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
            let result = try await api.recommendations(
                page: targetPage,
                lottery: lottery,
                status: status,
                sort: sort
            )
            items = reset ? result.items : items + result.items
            page = result.page
            hasMore = result.hasMore
            errorMessage = nil
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func detail(_ recommendation: Recommendation, using api: APIClient?) async -> Recommendation? {
        guard let api else { return nil }
        do {
            return try await api.recommendation(code: recommendation.lotteryCode, id: recommendation.id)
        } catch {
            errorMessage = error.localizedDescription
            return nil
        }
    }

    func generate(_ kind: LotteryKind, using api: APIClient?) async -> Recommendation? {
        guard let api, !isMutating else { return nil }
        isMutating = true
        defer { isMutating = false }
        do {
            let item = try await api.generateRecommendation(code: kind.rawValue)
            await load(using: api, reset: true)
            return item
        } catch {
            errorMessage = error.localizedDescription
            return nil
        }
    }

    func recheck(_ recommendation: Recommendation, using api: APIClient?) async -> Recommendation? {
        guard let api, !isMutating else { return nil }
        isMutating = true
        defer { isMutating = false }
        do {
            let item = try await api.recheckRecommendation(code: recommendation.lotteryCode, id: recommendation.id)
            replace(item)
            return item
        } catch {
            errorMessage = error.localizedDescription
            return nil
        }
    }

    func delete(_ recommendation: Recommendation, using api: APIClient?) async -> Bool {
        guard let api, !isMutating else { return false }
        isMutating = true
        defer { isMutating = false }
        do {
            _ = try await api.deleteRecommendation(code: recommendation.lotteryCode, id: recommendation.id)
            items.removeAll { $0.id == recommendation.id }
            return true
        } catch {
            errorMessage = error.localizedDescription
            return false
        }
    }

    private func replace(_ item: Recommendation) {
        guard let index = items.firstIndex(where: { $0.id == item.id }) else { return }
        items[index] = item
    }
}

struct RecommendationsView: View {
    @Environment(AppSession.self) private var session
    @State private var model = RecommendationsModel()
    @State private var selected: Recommendation?
    @State private var stealth: Recommendation?
    @State private var deleteCandidate: Recommendation?
    @State private var showsGenerate = false

    var body: some View {
        List {
            Section { filters }
            if let errorMessage = model.errorMessage {
                ErrorBanner(message: errorMessage)
                    .listRowBackground(Color.clear)
            }
            Section("推荐记录") {
                if model.items.isEmpty, !model.isLoading {
                    EmptyState(icon: "sparkles", title: "暂无推荐", message: "从右上角生成一组新的推荐号码。")
                        .listRowBackground(Color.clear)
                }
                ForEach(model.items) { item in
                    Button {
                        Task { selected = await model.detail(item, using: session.api) ?? item }
                    } label: {
                        RecommendationRow(recommendation: item)
                    }
                    .buttonStyle(.plain)
                    .onAppear {
                        if item.id == model.items.last?.id, model.hasMore {
                            Task { await model.load(using: session.api, reset: false) }
                        }
                    }
                }
                if model.isLoading { HStack { Spacer(); ProgressView(); Spacer() } }
            }
        }
        .navigationTitle("推荐")
        .toolbar {
            Button("生成", systemImage: "wand.and.sparkles") { showsGenerate = true }
        }
        .refreshable { await model.load(using: session.api, reset: true) }
        .task { await model.load(using: session.api, reset: true) }
        .onChange(of: model.lottery) { _, _ in reload() }
        .onChange(of: model.status) { _, _ in reload() }
        .onChange(of: model.sort) { _, _ in reload() }
        .sheet(item: $selected) { detailSheet($0) }
        .fullScreenCover(item: $stealth) { StealthRecommendationView(recommendation: $0) }
        .confirmationDialog("生成推荐", isPresented: $showsGenerate) {
            ForEach(LotteryKind.allCases) { kind in
                Button("生成\(kind.name)推荐") {
                    Task {
                        if let item = await model.generate(kind, using: session.api) {
                            selected = await model.detail(item, using: session.api) ?? item
                        }
                    }
                }
            }
        }
        .confirmationDialog(
            "删除这条推荐？",
            isPresented: Binding(get: { deleteCandidate != nil }, set: { if !$0 { deleteCandidate = nil } }),
            titleVisibility: .visible
        ) {
            Button("删除推荐", role: .destructive) {
                guard let candidate = deleteCandidate else { return }
                Task {
                    if await model.delete(candidate, using: session.api) { selected = nil }
                    deleteCandidate = nil
                }
            }
        } message: {
            Text("已录入的购买记录会保留，但不再关联这条推荐。")
        }
    }

    private var filters: some View {
        VStack(spacing: 10) {
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
                } label: { Label("状态筛选", systemImage: "line.3.horizontal.decrease") }
                Spacer()
                Menu {
                    Picker("排序", selection: $model.sort) {
                        Text("最新生成").tag("latest")
                        Text("最早生成").tag("oldest")
                        Text("奖金最高").tag("prize_high")
                    }
                } label: { Label("排序", systemImage: "arrow.up.arrow.down") }
            }
            .font(.subheadline)
        }
        .padding(.vertical, 4)
    }

    private func detailSheet(_ item: Recommendation) -> some View {
        NavigationStack {
            List {
                Section {
                    HStack {
                        Text(LotteryKind(rawValue: item.lotteryCode)?.name ?? item.lotteryCode)
                        Text("第 \(item.issue) 期").foregroundStyle(.secondary)
                        Spacer()
                        StatusPill(status: item.checkedAt == nil ? "pending" : item.prizeAmount > 0 ? "won" : "not_won")
                    }
                    if !item.summary.isEmpty { Text(item.summary).font(.subheadline) }
                }
                Section("推荐号码") {
                    ForEach(item.entries) { entry in
                        VStack(alignment: .leading, spacing: 8) {
                            HStack {
                                Text("推荐 \(entry.sequence)").font(.subheadline.weight(.semibold))
                                Spacer()
                                Text("\(Int(entry.confidence * 100))%").font(.caption).foregroundStyle(.secondary)
                            }
                            NumberBalls(redNumbers: entry.redNumbers, blueNumbers: entry.blueNumbers, compact: true)
                            if !entry.reason.isEmpty { Text(entry.reason).font(.caption).foregroundStyle(.secondary) }
                        }
                        .padding(.vertical, 5)
                    }
                }
                if let red = item.drawRedNumbers, let blue = item.drawBlueNumbers, !red.isEmpty {
                    Section("开奖号码") { NumberBalls(redNumbers: red, blueNumbers: blue) }
                }
            }
            .navigationTitle("推荐详情")
            .toolbar {
                ToolbarItemGroup(placement: .bottomBar) {
                    Button("隐览", systemImage: "eye.slash") { selected = nil; stealth = item }
                    Button(item.isPurchased == true ? "续购" : "购买", systemImage: "cart") {
                        selected = nil
                        session.recordPurchase(item)
                    }
                    Button("重判", systemImage: "arrow.clockwise") {
                        Task { if let updated = await model.recheck(item, using: session.api) { selected = updated } }
                    }
                    Button("删除", systemImage: "trash", role: .destructive) { deleteCandidate = item }
                }
            }
        }
        .presentationDetents([.medium, .large])
    }

    private func reload() {
        Task { await model.load(using: session.api, reset: true) }
    }
}

private struct RecommendationRow: View {
    let recommendation: Recommendation

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Text(LotteryKind(rawValue: recommendation.lotteryCode)?.name ?? recommendation.lotteryCode)
                    .font(.headline)
                Text("第 \(recommendation.issue) 期").foregroundStyle(.secondary)
                Spacer()
                StatusPill(status: recommendation.checkedAt == nil ? "pending" : recommendation.prizeAmount > 0 ? "won" : "not_won")
            }
            if let first = recommendation.entries.first {
                NumberBalls(redNumbers: first.redNumbers, blueNumbers: first.blueNumbers, compact: true)
            }
            HStack {
                Text(recommendation.isPurchased == true ? "已购买 \(recommendation.purchasedCount ?? 1) 次" : "未购买")
                Spacer()
                if recommendation.prizeAmount > 0 { Text(LotteryFormatters.currency(recommendation.prizeAmount)).foregroundStyle(LotteryPalette.jade) }
            }
            .font(.caption)
            .foregroundStyle(.secondary)
        }
        .padding(.vertical, 5)
    }
}

private struct StealthRecommendationView: View {
    @Environment(\.dismiss) private var dismiss
    let recommendation: Recommendation
    @State private var exitCount = 0

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
                        Divider()
                    }
                }
                .overlay(RoundedRectangle(cornerRadius: 2).stroke(Color(.separator)))
                .padding(.horizontal, 12)
            }
        }
        .background(Color(.systemBackground).ignoresSafeArea())
        .statusBarHidden()
    }

    private func registerExitTap() {
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

