import SwiftUI

struct PasskeyManagementView: View {
    @Environment(AppSession.self) private var session
    @Environment(\.dismiss) private var dismiss
    @State private var passkeys: [PasskeyItem] = []
    @State private var isLoading = true
    @State private var isRegistering = false
    @State private var deletingID: String?
    @State private var deleteCandidate: PasskeyItem?
    @State private var message: String?

    var body: some View {
        NavigationStack {
            Group {
                if isLoading {
                    ProgressView("正在读取 Passkey")
                } else if passkeys.isEmpty {
                    ContentUnavailableView(
                        "尚未添加 Passkey",
                        systemImage: "person.badge.key",
                        description: Text("添加后可直接使用 Face ID 或设备密码登录，无需输入用户名。")
                    )
                } else {
                    List(passkeys) { passkey in
                        passkeyRow(passkey)
                    }
                }
            }
            .navigationTitle("Passkey")
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("完成") { dismiss() }
                }
                ToolbarItem(placement: .primaryAction) {
                    Button(isRegistering ? "添加中" : "添加", systemImage: "plus") {
                        registerPasskey()
                    }
                    .disabled(isRegistering || deletingID != nil)
                }
            }
            .safeAreaInset(edge: .bottom) {
                if let message {
                    Text(message)
                        .font(.footnote)
                        .foregroundStyle(.secondary)
                        .padding(.horizontal, 16)
                        .padding(.vertical, 10)
                        .frame(maxWidth: .infinity)
                        .background(.bar)
                }
            }
            .task { await loadPasskeys() }
            .alert("删除 Passkey？", isPresented: deleteAlertPresented) {
                Button("取消", role: .cancel) {}
                Button("删除", role: .destructive) {
                    guard let deleteCandidate else { return }
                    deletePasskey(deleteCandidate)
                }
            } message: {
                Text("删除后，这个 Passkey 将不能再用于登录。")
            }
        }
    }

    private func passkeyRow(_ passkey: PasskeyItem) -> some View {
        HStack(spacing: 14) {
            Image(systemName: "key.fill")
                .foregroundStyle(.blue)
                .frame(width: 36, height: 36)
                .background(.blue.opacity(0.1), in: .circle)
            VStack(alignment: .leading, spacing: 4) {
                Text("Passkey")
                    .font(.headline)
                Text(createdAtText(passkey.createdAt))
                    .font(.caption)
                    .foregroundStyle(.secondary)
                Text(String(passkey.id.suffix(8)))
                    .font(.caption2.monospaced())
                    .foregroundStyle(.tertiary)
            }
            Spacer()
            if deletingID == passkey.id {
                ProgressView()
            }
        }
        .swipeActions {
            Button("删除", systemImage: "trash", role: .destructive) {
                deleteCandidate = passkey
            }
        }
    }

    private var deleteAlertPresented: Binding<Bool> {
        Binding(
            get: { deleteCandidate != nil },
            set: { isPresented in
                if !isPresented { deleteCandidate = nil }
            }
        )
    }

    private func loadPasskeys() async {
        isLoading = true
        defer { isLoading = false }
        do {
            passkeys = try await session.passkeys()
            message = nil
        } catch {
            message = "读取失败：\(error.localizedDescription)"
        }
    }

    private func registerPasskey() {
        isRegistering = true
        message = nil
        Task {
            defer { isRegistering = false }
            do {
                try await session.registerPasskey()
                passkeys = try await session.passkeys()
                message = "Passkey 已添加"
            } catch {
                message = "添加失败：\(error.localizedDescription)"
            }
        }
    }

    private func deletePasskey(_ passkey: PasskeyItem) {
        deleteCandidate = nil
        deletingID = passkey.id
        message = nil
        Task {
            defer { deletingID = nil }
            do {
                try await session.deletePasskey(id: passkey.id)
                passkeys.removeAll { $0.id == passkey.id }
                message = "Passkey 已删除"
            } catch {
                message = "删除失败：\(error.localizedDescription)"
            }
        }
    }

    private func createdAtText(_ value: String) -> String {
        guard let date = DateParser.date(value) else { return "创建时间未知" }
        return "创建于 \(date.formatted(date: .abbreviated, time: .shortened))"
    }
}
