import SwiftUI

struct RootView: View {
    @EnvironmentObject private var appState: VyrdenAppState

    var body: some View {
        ZStack {
            VyrdenTheme.cinema.ignoresSafeArea()
            if !appState.isPaired {
                PairingView()
            } else {
                HomeView()
            }
        }
    }
}
