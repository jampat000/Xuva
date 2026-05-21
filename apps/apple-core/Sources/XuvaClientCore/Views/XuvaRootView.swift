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
            if let error = store.errorMessage, store.screen != .player {
                ErrorToast(message: error) { store.clearError() }
                    .padding(.top, 60)
                    .frame(maxHeight: .infinity, alignment: .top)
                    .transition(.move(edge: .top).combined(with: .opacity))
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

private struct ErrorToast: View {
    let message: String
    let dismiss: () -> Void

    var body: some View {
        HStack(spacing: 14) {
            Image(systemName: "exclamationmark.triangle.fill")
                .foregroundStyle(XuvaTheme.warn)
            Text(message)
                .font(.system(size: 17, weight: .medium))
                .foregroundStyle(XuvaTheme.text)
                .lineLimit(3)
                .frame(maxWidth: 720, alignment: .leading)
            Button {
                dismiss()
            } label: {
                Image(systemName: "xmark")
            }
            .buttonStyle(.plain)
            .foregroundStyle(XuvaTheme.mutedText)
        }
        .padding(.horizontal, 22)
        .padding(.vertical, 14)
        .background(XuvaTheme.surface.opacity(0.94), in: RoundedRectangle(cornerRadius: 16, style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: 16, style: .continuous).stroke(XuvaTheme.warn.opacity(0.45)))
        .shadow(color: .black.opacity(0.4), radius: 24, y: 12)
    }
}
