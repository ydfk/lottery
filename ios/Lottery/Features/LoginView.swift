import SwiftUI

struct LoginView: View {
    @Environment(AppSession.self) private var session
    @State private var username = ""
    @State private var password = ""
    @State private var isLoading = false
    @State private var errorMessage: String?

    var body: some View {
        ScrollView {
            VStack(spacing: 28) {
                Spacer(minLength: 72)
                Image("BrandMark")
                    .resizable()
                    .scaledToFit()
                    .frame(width: 96, height: 96)
                    .accessibilityHidden(true)

                VStack(spacing: 8) {
                    Text("彩迹")
                        .font(.system(size: 38, weight: .bold, design: .rounded))
                    Text("记录每一次选号与开奖结果")
                        .foregroundStyle(.secondary)
                }

                VStack(spacing: 14) {
                    TextField("用户名", text: $username)
                        .textContentType(.username)
                        .textInputAutocapitalization(.never)
                        .autocorrectionDisabled()
                        .padding(14)
                        .background(.background, in: .rect(cornerRadius: 16))
                        .accessibilityIdentifier("username")
                    SecureField("密码", text: $password)
                        .textContentType(.password)
                        .padding(14)
                        .background(.background, in: .rect(cornerRadius: 16))
                        .accessibilityIdentifier("password")
                    if let errorMessage {
                        ErrorBanner(message: errorMessage)
                    }
                    Button(action: signIn) {
                        HStack {
                            if isLoading { ProgressView().tint(.white) }
                            Text(isLoading ? "登录中" : "登录")
                                .frame(maxWidth: .infinity)
                        }
                    }
                    .buttonStyle(.glassProminent)
                    .controlSize(.extraLarge)
                    .disabled(username.trimmingCharacters(in: .whitespaces).isEmpty || password.isEmpty || isLoading)
                    .accessibilityIdentifier("login")
                }
                .padding(20)
                .glassEffect(.regular, in: .rect(cornerRadius: 28))
            }
            .frame(maxWidth: 460)
            .padding(.horizontal, 24)
            .frame(maxWidth: .infinity)
        }
        .scrollDismissesKeyboard(.interactively)
    }

    private func signIn() {
        isLoading = true
        errorMessage = nil
        Task {
            defer { isLoading = false }
            do {
                try await session.login(
                    username: username.trimmingCharacters(in: .whitespacesAndNewlines),
                    password: password
                )
            } catch {
                errorMessage = error.localizedDescription
            }
        }
    }
}

