import SwiftUI

@main
struct XuvaTVApp: App {
    @StateObject private var appState = XuvaAppState()

    var body: some Scene {
        WindowGroup {
            RootView()
                .environmentObject(appState)
                .preferredColorScheme(.dark)
        }
    }
}
