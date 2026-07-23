import PhotosUI
import SwiftUI
import UIKit

struct TicketEditorView: View {
    @Environment(AppSession.self) private var session
    @State private var photoItem: PhotosPickerItem?
    @State private var imageData: Data?
    @State private var upload: TicketUpload?
    @State private var confidence: Double?
    @State private var selectedEntryID: UUID?
    @State private var showsCamera = false
    @State private var isRecognizing = false
    @State private var errorMessage: String?
    @State private var successMessage: String?
    @State private var costManuallyEdited = false
    private let draftStore = TicketDraftStore()

    var body: some View {
        @Bindable var session = session
        Form {
            if let errorMessage { Section { ErrorBanner(message: errorMessage) } }
            if let successMessage {
                Section { Label(successMessage, systemImage: "checkmark.circle.fill").foregroundStyle(LotteryPalette.jade) }
            }
            Section("基本信息") {
                Picker("彩票类型", selection: $session.ticketDraft.lottery) {
                    ForEach(LotteryKind.allCases) { kind in Text(kind.name).tag(kind) }
                }
                TextField("期号", text: $session.ticketDraft.issue)
                    .keyboardType(.numberPad)
                DatePicker("开奖日期", selection: $session.ticketDraft.drawDate, displayedComponents: .date)
                DatePicker("购买时间", selection: $session.ticketDraft.purchasedAt)
                HStack {
                    Text("金额")
                    Spacer()
                    TextField("0.00", value: $session.ticketDraft.costAmount, format: .number.precision(.fractionLength(2)))
                        .keyboardType(.decimalPad)
                        .multilineTextAlignment(.trailing)
                        .onChange(of: session.ticketDraft.costAmount) { _, _ in costManuallyEdited = true }
                    Text("元").foregroundStyle(.secondary)
                }
            }

            Section {
                ForEach($session.ticketDraft.entries) { $entry in
                    entryRow(entry: $entry, kind: session.ticketDraft.lottery)
                }
                .onDelete { indexes in session.ticketDraft.entries.remove(atOffsets: indexes) }
                Button("新增一注", systemImage: "plus.circle") {
                    session.ticketDraft.entries.insert(TicketEntryDraft(), at: 0)
                }
            } header: {
                Text("号码")
            } footer: {
                Text("号码按规则自动排序；金额会根据注数、倍数和追加自动计算。")
            }

            imageSection

            Section("备注") {
                TextField("可选备注", text: $session.ticketDraft.notes, axis: .vertical)
                    .lineLimit(3...6)
            }

            Section {
                Button(session.editingTicket == nil ? "保存票据" : "保存修改", systemImage: "checkmark.circle.fill") {
                    save()
                }
                .disabled(!session.ticketDraft.isValid || session.editorSaving)
                if session.editingTicket != nil {
                    Button("取消编辑", role: .cancel) { resetEditor() }
                }
            }
        }
        .navigationTitle(session.editingTicket == nil ? "记录票据" : "编辑票据")
        .scrollDismissesKeyboard(.interactively)
        .sheet(isPresented: $showsCamera) {
            CameraPicker { data in setImageData(data) }.ignoresSafeArea()
        }
        .sheet(isPresented: Binding(get: { selectedEntryID != nil }, set: { if !$0 { selectedEntryID = nil } })) {
            if let id = selectedEntryID,
               let index = session.ticketDraft.entries.firstIndex(where: { $0.id == id }) {
                NumberPickerSheet(entry: $session.ticketDraft.entries[index], kind: session.ticketDraft.lottery)
            }
        }
        .onChange(of: photoItem) { _, item in loadPhoto(item) }
        .onChange(of: session.ticketDraft.entries) { _, _ in updateCalculatedCost() }
        .onChange(of: session.ticketDraft.lottery) { _, _ in normalizeEntries() }
        .onChange(of: session.ticketDraft) { _, draft in
            session.editorCanSave = draft.isValid
            if session.editingTicket == nil { draftStore.save(draft) }
        }
        .onChange(of: session.saveTicketTrigger) { _, _ in save() }
        .task {
            if session.editingTicket == nil,
               session.ticketDraft.issue.isEmpty,
               let stored = draftStore.load() {
                session.ticketDraft = stored
            }
            session.editorCanSave = session.ticketDraft.isValid
        }
    }

    private func entryRow(entry: Binding<TicketEntryDraft>, kind: LotteryKind) -> some View {
        VStack(alignment: .leading, spacing: 12) {
            Button {
                selectedEntryID = entry.wrappedValue.id
            } label: {
                VStack(alignment: .leading, spacing: 9) {
                    HStack {
                        Text(entry.wrappedValue.isValid(for: kind) ? "号码已完成" : "选择号码")
                            .font(.subheadline.weight(.semibold))
                        Spacer()
                        Image(systemName: "chevron.right").foregroundStyle(.tertiary)
                    }
                    if !entry.wrappedValue.red.isEmpty || !entry.wrappedValue.blue.isEmpty {
                        NumberBalls(
                            redNumbers: entry.wrappedValue.red.sorted().formattedNumbers,
                            blueNumbers: entry.wrappedValue.blue.sorted().formattedNumbers,
                            compact: true
                        )
                    }
                }
            }
            .buttonStyle(.plain)
            Stepper("倍数：\(entry.wrappedValue.multiple)", value: entry.multiple, in: 1...99)
            if kind == .dlt {
                Toggle("追加投注", isOn: entry.isAdditional)
            }
        }
        .padding(.vertical, 5)
    }

    private var imageSection: some View {
        Section {
            if let imageData, let image = UIImage(data: imageData) {
                Image(uiImage: image)
                    .resizable()
                    .scaledToFill()
                    .frame(height: 190)
                    .clipShape(.rect(cornerRadius: 18))
                    .accessibilityLabel("彩票照片预览")
            }
            HStack {
                PhotosPicker(selection: $photoItem, matching: .images) {
                    Label("相册", systemImage: "photo.on.rectangle")
                }
                Spacer()
                Button("拍照", systemImage: "camera") { showsCamera = true }
            }
            if imageData != nil, session.editingTicket == nil {
                Button(isRecognizing ? "识别中" : "图片识别", systemImage: "viewfinder") {
                    recognize()
                }
                .disabled(isRecognizing)
            }
            if let confidence {
                LabeledContent("识别置信度", value: "\(Int(confidence * 100))%")
            }
        } header: {
            Text("图片识别辅助")
        } footer: {
            Text(session.editingTicket == nil ? "识别结果只会回填表单，保存前仍可完整修正。" : "编辑时沿用原图，可修改其他信息。")
        }
    }

    private func loadPhoto(_ item: PhotosPickerItem?) {
        guard let item else { return }
        Task {
            do {
                guard let data = try await item.loadTransferable(type: Data.self) else { return }
                setImageData(data)
            } catch {
                errorMessage = "无法读取所选照片：\(error.localizedDescription)"
            }
        }
    }

    private func setImageData(_ data: Data) {
        imageData = data
        upload = nil
        confidence = nil
        errorMessage = nil
    }

    private func recognize() {
        guard let api = session.api, let imageData else { return }
        isRecognizing = true
        errorMessage = nil
        Task {
            defer { isRecognizing = false }
            do {
                let uploaded: TicketUpload
                if let upload {
                    uploaded = upload
                } else {
                    uploaded = try await api.uploadTicketImage(data: imageData, filename: "ticket.jpg")
                }
                upload = uploaded
                let recognized = try await api.recognizeTicket(uploadId: uploaded.id)
                apply(recognized)
                successMessage = "识别完成，请检查号码后保存。"
            } catch {
                errorMessage = error.localizedDescription
            }
        }
    }

    private func apply(_ recognized: RecognitionDraft) {
        session.ticketDraft.lottery = LotteryKind(rawValue: recognized.lotteryCode) ?? session.ticketDraft.lottery
        session.ticketDraft.issue = recognized.issue
        session.ticketDraft.drawDate = LotteryFormatters.dateOnly.date(from: recognized.drawDate) ?? session.ticketDraft.drawDate
        session.ticketDraft.entries = recognized.entries.map {
            TicketEntryDraft(red: Set($0.red), blue: Set($0.blue), multiple: max(1, $0.multiple), isAdditional: $0.isAdditional)
        }
        session.ticketDraft.uploadId = recognized.upload.id
        session.ticketDraft.costAmount = recognized.costAmount > 0 ? recognized.costAmount : session.ticketDraft.calculatedCost
        confidence = recognized.confidence
        costManuallyEdited = recognized.costAmount > 0
    }

    private func updateCalculatedCost() {
        guard !costManuallyEdited else { return }
        session.ticketDraft.costAmount = session.ticketDraft.calculatedCost
    }

    private func normalizeEntries() {
        let rules = session.ticketDraft.lottery.rules
        for index in session.ticketDraft.entries.indices {
            session.ticketDraft.entries[index].red = Set(session.ticketDraft.entries[index].red.filter { rules.redMin...rules.redMax ~= $0 })
            session.ticketDraft.entries[index].blue = Set(session.ticketDraft.entries[index].blue.filter { rules.blueMin...rules.blueMax ~= $0 })
            if session.ticketDraft.lottery == .ssq { session.ticketDraft.entries[index].isAdditional = false }
        }
        costManuallyEdited = false
        updateCalculatedCost()
    }

    private func save() {
        guard !session.editorSaving, session.ticketDraft.isValid, let api = session.api else { return }
        session.editorSaving = true
        errorMessage = nil
        let draft = session.ticketDraft
        let editingID = session.editingTicket?.id
        Task {
            defer { session.editorSaving = false }
            do {
                let payload = TicketPayload(
                    lotteryCode: draft.lottery.rawValue,
                    uploadId: draft.uploadId,
                    recommendationId: draft.recommendationId,
                    issue: draft.issue,
                    drawDate: LotteryFormatters.dateOnly.string(from: draft.drawDate),
                    purchasedAt: ISO8601DateFormatter().string(from: draft.purchasedAt),
                    costAmount: draft.costAmount,
                    notes: draft.notes,
                    entries: draft.entries.map { $0.payload() }
                )
                if let editingID { _ = try await api.updateTicket(id: editingID, payload: payload) }
                else { _ = try await api.createTicket(payload) }
                draftStore.clear()
                resetEditor()
                session.selectedTab = .history
                session.message = editingID == nil ? "票据已保存并完成判奖检查。" : "票据修改已保存。"
            } catch {
                errorMessage = error.localizedDescription
            }
        }
    }

    private func resetEditor() {
        draftStore.clear()
        imageData = nil
        upload = nil
        confidence = nil
        costManuallyEdited = false
        session.resetEditor()
    }
}

private struct NumberPickerSheet: View {
    @Environment(\.dismiss) private var dismiss
    @Binding var entry: TicketEntryDraft
    let kind: LotteryKind

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(alignment: .leading, spacing: 24) {
                    numberGrid(
                        title: kind == .dlt ? "前区 · 选 \(kind.rules.redCount) 个" : "红球 · 选 \(kind.rules.redCount) 个",
                        range: kind.rules.redMin...kind.rules.redMax,
                        selection: $entry.red,
                        limit: kind.rules.redCount,
                        color: LotteryPalette.red
                    )
                    numberGrid(
                        title: kind == .dlt ? "后区 · 选 \(kind.rules.blueCount) 个" : "蓝球 · 选 \(kind.rules.blueCount) 个",
                        range: kind.rules.blueMin...kind.rules.blueMax,
                        selection: $entry.blue,
                        limit: kind.rules.blueCount,
                        color: LotteryPalette.blue
                    )
                }
                .padding()
            }
            .navigationTitle("选择号码")
            .toolbar { Button("完成") { dismiss() }.disabled(!entry.isValid(for: kind)) }
        }
        .presentationDetents([.large])
    }

    private func numberGrid(
        title: String,
        range: ClosedRange<Int>,
        selection: Binding<Set<Int>>,
        limit: Int,
        color: Color
    ) -> some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Text(title).font(.headline)
                Spacer()
                Text("已选 \(selection.wrappedValue.count)").font(.caption).foregroundStyle(.secondary)
            }
            LazyVGrid(columns: Array(repeating: GridItem(.flexible()), count: 7), spacing: 10) {
                ForEach(Array(range), id: \.self) { number in
                    let selected = selection.wrappedValue.contains(number)
                    Button {
                        if selected { selection.wrappedValue.remove(number) }
                        else if selection.wrappedValue.count < limit { selection.wrappedValue.insert(number) }
                    } label: {
                        Text(String(format: "%02d", number))
                            .font(.system(.subheadline, design: .rounded, weight: .bold))
                            .foregroundStyle(selected ? .white : color)
                            .frame(maxWidth: .infinity)
                            .aspectRatio(1, contentMode: .fit)
                            .background(selected ? color : color.opacity(0.1), in: .circle)
                    }
                    .buttonStyle(.plain)
                    .accessibilityLabel("号码 \(number)")
                    .accessibilityValue(selected ? "已选择" : "未选择")
                }
            }
        }
    }
}
