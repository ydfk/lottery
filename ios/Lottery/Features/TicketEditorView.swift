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
    @State private var recognizedCost: Double?
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
                LabeledContent("应付金额", value: LotteryFormatters.currency(session.ticketDraft.calculatedCost))
                    .accessibilityIdentifier("ticket-cost")
                if let recognizedCost,
                   session.ticketDraft.recognizedCostDiffers(from: recognizedCost) {
                    Label {
                        Text("票面识别为 \(LotteryFormatters.currency(recognizedCost))，以规则金额为准。")
                    } icon: {
                        Image(systemName: "exclamationmark.triangle.fill")
                    }
                    .font(.footnote)
                    .foregroundStyle(LotteryPalette.amber)
                    .accessibilityLabel("识别金额与规则金额不一致，以规则金额为准")
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

            Section("投注合计") {
                LabeledContent("有效注数", value: "\(session.ticketDraft.validEntryCount) 注")
                LabeledContent("总投注倍数", value: "\(session.ticketDraft.totalMultiple) 倍")
                LabeledContent {
                    Text(LotteryFormatters.currency(session.ticketDraft.calculatedCost))
                        .font(.title3.monospacedDigit().weight(.bold))
                } label: {
                    Text("最终应付")
                }
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
            MultipleControl(multiple: entry.multiple)
            if kind == .dlt {
                Toggle(isOn: entry.isAdditional) {
                    HStack(spacing: 8) {
                        Label("追加投注", systemImage: "plus.circle.fill")
                        Spacer()
                        Text("+1 元/注")
                            .font(.caption.weight(.semibold))
                            .foregroundStyle(.secondary)
                    }
                    .frame(maxWidth: .infinity, alignment: .leading)
                }
                .toggleStyle(.button)
                .buttonStyle(.bordered)
                .tint(entry.wrappedValue.isAdditional ? LotteryPalette.blue : .secondary)
                .accessibilityHint("开启后每注由 2 元变为 3 元")
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
        recognizedCost = nil
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
            TicketEntryDraft(
                red: Set($0.red),
                blue: Set($0.blue),
                multiple: TicketEntryDraft.clampedMultiple($0.multiple),
                isAdditional: $0.isAdditional
            )
        }
        session.ticketDraft.uploadId = recognized.upload.id
        recognizedCost = recognized.costAmount > 0 ? recognized.costAmount : nil
        confidence = recognized.confidence
    }

    private func normalizeEntries() {
        let rules = session.ticketDraft.lottery.rules
        for index in session.ticketDraft.entries.indices {
            session.ticketDraft.entries[index].red = Set(session.ticketDraft.entries[index].red.filter { rules.redMin...rules.redMax ~= $0 })
            session.ticketDraft.entries[index].blue = Set(session.ticketDraft.entries[index].blue.filter { rules.blueMin...rules.blueMax ~= $0 })
            if session.ticketDraft.lottery == .ssq { session.ticketDraft.entries[index].isAdditional = false }
        }
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
                    costAmount: draft.calculatedCost,
                    notes: draft.notes,
                    entries: draft.entries.map { $0.payload() }
                )
                let saved: Ticket
                if let editingID { saved = try await api.updateTicket(id: editingID, payload: payload) }
                else { saved = try await api.createTicket(payload) }
                draftStore.clear()
                resetEditor()
                session.selectedTab = .history
                session.message = "已保存，服务端确认金额为 \(LotteryFormatters.currency(saved.costAmount))。"
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
        recognizedCost = nil
        session.resetEditor()
    }
}

private struct MultipleControl: View {
    @Binding var multiple: Int

    var body: some View {
        HStack {
            Label("投注倍数", systemImage: "multiply.circle")
                .font(.subheadline.weight(.semibold))
            Spacer()
            ControlGroup {
                Button("减少倍数", systemImage: "minus") {
                    multiple = TicketEntryDraft.clampedMultiple(multiple - 1)
                }
                .disabled(multiple <= 1)

                Text("\(multiple) 倍")
                    .font(.body.monospacedDigit().weight(.semibold))
                    .frame(minWidth: 52)

                Button("增加倍数", systemImage: "plus") {
                    multiple = TicketEntryDraft.clampedMultiple(multiple + 1)
                }
                .disabled(multiple >= 99)
            }
        }
        .accessibilityElement(children: .contain)
        .accessibilityLabel("投注倍数")
        .accessibilityValue("\(multiple) 倍")
        .accessibilityAdjustableAction { direction in
            switch direction {
            case .increment: multiple = TicketEntryDraft.clampedMultiple(multiple + 1)
            case .decrement: multiple = TicketEntryDraft.clampedMultiple(multiple - 1)
            @unknown default: break
            }
        }
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
