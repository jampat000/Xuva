import SwiftUI

@main
struct LorivoTVApp: App {
    @StateObject private var appState = LorivoAppState()

    var body: some Scene {
        WindowGroup {
            RootView()
                .environmentObject(appState)
                .preferredColorScheme(.dark)
        }
    }
}
