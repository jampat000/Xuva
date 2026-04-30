import SwiftUI

@main
struct VyrdenTVApp: App {
    @StateObject private var appState = VyrdenAppState()

    var body: some Scene {
        WindowGroup {
            RootView()
                .environmentObject(appState)
                .preferredColorScheme(.dark)
        }
    }
}
