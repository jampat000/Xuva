import SwiftUI

struct RootView: View {
    @EnvironmentObject private var appState: XuvaAppState

    var body: some View {
        ZStack {
            XuvaTheme.cinema.ignoresSafeArea()
            if !appState.isPaired {
                PairingView()
            } else {
                HomeView()
            }
        }
    }
}
