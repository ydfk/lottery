import SwiftUI

@main
struct LotteryApp: App {
    @State private var session = AppSession()

    var body: some Scene {
        WindowGroup {
            RootView()
                .environment(session)
                .onReceive(NotificationCenter.default.publisher(for: .sessionUnauthorized)) { _ in
                    if !session.isUITestSession {
                        session.logout(message: "登录已失效，请重新登录。")
                    }
                }
        }
    }
}
