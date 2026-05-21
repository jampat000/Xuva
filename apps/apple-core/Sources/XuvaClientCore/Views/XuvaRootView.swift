import SwiftUI

public struct XuvaRootView: View {
    @StateObject private var store = XuvaClientStore()

    public init() {}

    public var body: some View {
        ZStack {
            XuvaTheme.backgroundWash
            switch store.screen {
            case .connect, .pair:
                PairingScreen()
            case .home:
                HomeScreen()
            case .detail:
                DetailScreen()
            case .player:
                PlayerScreen()
            }
            if store.isBusy {
                ProgressView()
                    .controlSize(.large)
                    .tint(.white)
                    .padding(28)
                    .background(.black.opacity(0.48), in: RoundedRectangle(cornerRadius: 24, style: .continuous))
            }
        }
        .environmentObject(store)
        .preferredColorScheme(.dark)
        .task {
            await store.resumeSessionIfPossible()
            await store.autoConnectIfPossible()
        }
    }
}
