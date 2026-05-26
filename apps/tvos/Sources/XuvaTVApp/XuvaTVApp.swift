import SwiftUI
import XuvaClientCore

@main
struct XuvaTVApp: App {
    var body: some Scene {
        WindowGroup {
            XuvaRootView()
                .onOpenURL { url in
                    handleDeepLink(url)
                }
        }
    }

    private func handleDeepLink(_ url: URL) {
        guard url.scheme == "xuva", url.host == "open" else { return }
        let parts = url.pathComponents.filter { $0 != "/" }
        guard parts.count >= 2 else { return }
        let kind = parts[0]
        let id = parts[1]
        // Post a notification; XuvaRootView's store receives it via onReceive.
        NotificationCenter.default.post(
            name: .xuvaOpenDeepLink,
            object: nil,
            userInfo: ["kind": kind, "id": id]
        )
    }
}

extension Notification.Name {
    static let xuvaOpenDeepLink = Notification.Name("xuva.openDeepLink")
}
