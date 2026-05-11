import SwiftUI

struct RootView: View {
    @EnvironmentObject private var appState: LorivoAppState

    var body: some View {
        ZStack {
            LorivoTheme.cinema.ignoresSafeArea()
            if !appState.isPaired {
                PairingView()
            } else {
                HomeView()
            }
        }
    }
}
