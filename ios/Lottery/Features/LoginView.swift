import SwiftUI

struct LoginView: View {
    private enum Field: Hashable {
        case username
        case password
    }

    @Environment(AppSession.self) private var session
    @Environment(\.accessibilityReduceTransparency) private var reduceTransparency
    @State private var username = ""
    @State private var password = ""
    @State private var showsPassword = false
    @State private var isLoading = false
    @State private var errorMessage: String?
    @FocusState private var focusedField: Field?

    var body: some View {
        ScrollView {
            VStack(spacing: 0) {
                brandTicket
                loginForm
            }
            .frame(maxWidth: 460)
            .padding(.horizontal, 20)
            .padding(.vertical, 28)
            .frame(maxWidth: .infinity)
        }
        .background(Color(.systemGroupedBackground))
        .scrollDismissesKeyboard(.interactively)
    }

    private var brandTicket: some View {
        VStack(alignment: .leading, spacing: 20) {
            HStack(alignment: .top) {
                Image("BrandMark")
                    .resizable()
                    .scaledToFit()
                    .frame(width: 62, height: 62)
                    .accessibilityHidden(true)
                Spacer()
                Text("LOTTERY TRACE")
                    .font(.caption2.monospaced().weight(.semibold))
                    .tracking(1.5)
                    .foregroundStyle(.white.opacity(0.55))
            }

            LoginTrail()
                .stroke(
                    LinearGradient(
                        colors: [LotteryPalette.red, .white.opacity(0.55), LotteryPalette.blue],
                        startPoint: .leading,
                        endPoint: .trailing
                    ),
                    style: StrokeStyle(lineWidth: 7, lineCap: .square)
                )
                .frame(height: 46)
                .accessibilityHidden(true)

            VStack(alignment: .leading, spacing: 6) {
                Text("彩迹")
                    .font(.system(.largeTitle, design: .rounded, weight: .bold))
                Text("每一张票，都有迹可循")
                    .font(.subheadline)
                    .foregroundStyle(.white.opacity(0.68))
            }
        }
        .foregroundStyle(.white)
        .padding(26)
        .background(Color(red: 0.035, green: 0.075, blue: 0.12))
        .clipShape(.rect(topLeadingRadius: 30, topTrailingRadius: 30))
        .accessibilityElement(children: .combine)
        .accessibilityLabel("彩迹，每一张票，都有迹可循")
    }

    private var loginForm: some View {
        VStack(alignment: .leading, spacing: 20) {
            VStack(alignment: .leading, spacing: 6) {
                Text("欢迎回来")
                    .font(.title2.weight(.bold))
                Text("登录后继续记录选号与开奖轨迹")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
            }

            VStack(alignment: .leading, spacing: 8) {
                Text("用户名").font(.subheadline.weight(.semibold))
                HStack(spacing: 12) {
                    Image(systemName: "person.fill").foregroundStyle(.secondary)
                    TextField("请输入用户名", text: $username)
                        .textContentType(.username)
                        .textInputAutocapitalization(.never)
                        .autocorrectionDisabled()
                        .focused($focusedField, equals: .username)
                        .submitLabel(.next)
                        .onSubmit { focusedField = .password }
                        .accessibilityIdentifier("username")
                }
                .padding(.horizontal, 16)
                .frame(minHeight: 54)
                .background(Color(.secondarySystemGroupedBackground), in: .rect(cornerRadius: 16))
            }

            VStack(alignment: .leading, spacing: 8) {
                Text("密码").font(.subheadline.weight(.semibold))
                HStack(spacing: 12) {
                    Image(systemName: "lock.fill").foregroundStyle(.secondary)
                    Group {
                        if showsPassword {
                            TextField("请输入密码", text: $password)
                        } else {
                            SecureField("请输入密码", text: $password)
                        }
                    }
                    .textContentType(.password)
                    .focused($focusedField, equals: .password)
                    .submitLabel(.go)
                    .onSubmit(signIn)
                    .accessibilityIdentifier("password")

                    Button(showsPassword ? "隐藏密码" : "显示密码", systemImage: showsPassword ? "eye.slash" : "eye") {
                        showsPassword.toggle()
                    }
                    .labelStyle(.iconOnly)
                    .foregroundStyle(.secondary)
                }
                .padding(.horizontal, 16)
                .frame(minHeight: 54)
                .background(Color(.secondarySystemGroupedBackground), in: .rect(cornerRadius: 16))
            }

            if let errorMessage {
                ErrorBanner(message: errorMessage)
                    .accessibilityAddTraits(.updatesFrequently)
            }

            Button(action: signIn) {
                HStack(spacing: 10) {
                    if isLoading { ProgressView() }
                    Text(isLoading ? "登录中" : "登录")
                        .frame(maxWidth: .infinity)
                }
            }
            .buttonStyle(.glassProminent)
            .controlSize(.extraLarge)
            .disabled(!canSignIn)
            .accessibilityIdentifier("login")
        }
        .padding(24)
        .background(
            reduceTransparency ? Color(.systemBackground) : Color(.systemBackground).opacity(0.96),
            in: .rect(bottomLeadingRadius: 30, bottomTrailingRadius: 30)
        )
    }

    private var canSignIn: Bool {
        !username.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
            && !password.isEmpty
            && !isLoading
    }

    private func signIn() {
        guard canSignIn else { return }
        focusedField = nil
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

private struct LoginTrail: Shape {
    func path(in rect: CGRect) -> Path {
        var path = Path()
        path.move(to: CGPoint(x: rect.minX, y: rect.height * 0.7))
        path.addCurve(
            to: CGPoint(x: rect.maxX, y: rect.height * 0.38),
            control1: CGPoint(x: rect.width * 0.26, y: -rect.height * 0.15),
            control2: CGPoint(x: rect.width * 0.68, y: rect.height * 1.25)
        )
        return path
    }
}
