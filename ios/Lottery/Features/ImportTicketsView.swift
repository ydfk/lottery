import SwiftUI
import UniformTypeIdentifiers

struct ImportTicketsView: View {
    @Environment(AppSession.self) private var session
    @Environment(\.dismiss) private var dismiss
    let onComplete: () -> Void

    @State private var workbookURL: URL?
    @State private var archiveURL: URL?
    @State private var selectingWorkbook = false
    @State private var selectingArchive = false
    @State private var isImporting = false
    @State private var result: ImportResult?
    @State private var errorMessage: String?

    private var workbookType: UTType { UTType(filenameExtension: "xlsx") ?? .data }

    var body: some View {
        NavigationStack {
            Form {
                Section("Excel 工作簿") {
                    fileRow(url: workbookURL, placeholder: "选择 .xlsx 文件") { selectingWorkbook = true }
                }
                Section("图片压缩包（可选）") {
                    fileRow(url: archiveURL, placeholder: "选择 .zip 文件") { selectingArchive = true }
                }
                if let errorMessage { Section { ErrorBanner(message: errorMessage) } }
                if let result {
                    Section("导入结果") {
                        LabeledContent("总计", value: "\(result.totalCount)")
                        LabeledContent("成功", value: "\(result.successCount)")
                        LabeledContent("失败", value: "\(result.failedCount)")
                        ForEach(result.rows.filter { $0.status != "success" }) { row in
                            VStack(alignment: .leading) {
                                Text("第 \(row.row) 行").font(.subheadline.weight(.semibold))
                                Text(row.message ?? "导入失败").font(.caption).foregroundStyle(.secondary)
                            }
                        }
                    }
                }
                Section {
                    Button(isImporting ? "导入中" : "开始导入", systemImage: "square.and.arrow.down") { importFiles() }
                        .disabled(workbookURL == nil || isImporting)
                }
            }
            .navigationTitle("导入历史记录")
            .toolbar {
                ToolbarItem(placement: .cancellationAction) { Button("关闭") { dismiss() } }
                if result?.successCount ?? 0 > 0 {
                    ToolbarItem(placement: .confirmationAction) { Button("完成") { onComplete() } }
                }
            }
        }
        .fileImporter(isPresented: $selectingWorkbook, allowedContentTypes: [workbookType]) { result in
            if case let .success(url) = result { workbookURL = url }
            else if case let .failure(error) = result { errorMessage = error.localizedDescription }
        }
        .fileImporter(isPresented: $selectingArchive, allowedContentTypes: [.zip]) { result in
            if case let .success(url) = result { archiveURL = url }
            else if case let .failure(error) = result { errorMessage = error.localizedDescription }
        }
    }

    private func fileRow(url: URL?, placeholder: String, action: @escaping () -> Void) -> some View {
        Button(action: action) {
            HStack {
                Image(systemName: url == nil ? "doc.badge.plus" : "doc.fill")
                Text(url?.lastPathComponent ?? placeholder)
                Spacer()
                Image(systemName: "chevron.right").foregroundStyle(.tertiary)
            }
        }
    }

    private func importFiles() {
        guard let api = session.api, let workbookURL else { return }
        isImporting = true
        errorMessage = nil
        Task {
            defer { isImporting = false }
            do {
                let workbookData = try readSecurityScoped(workbookURL)
                let workbook = MultipartPart(
                    name: "workbook",
                    filename: workbookURL.lastPathComponent,
                    mimeType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
                    data: workbookData
                )
                var archive: MultipartPart?
                if let archiveURL {
                    archive = MultipartPart(
                        name: "imagesZip",
                        filename: archiveURL.lastPathComponent,
                        mimeType: "application/zip",
                        data: try readSecurityScoped(archiveURL)
                    )
                }
                result = try await api.importTickets(workbook: workbook, imagesZip: archive)
            } catch {
                errorMessage = error.localizedDescription
            }
        }
    }

    private func readSecurityScoped(_ url: URL) throws -> Data {
        let accessed = url.startAccessingSecurityScopedResource()
        defer { if accessed { url.stopAccessingSecurityScopedResource() } }
        return try Data(contentsOf: url)
    }
}

