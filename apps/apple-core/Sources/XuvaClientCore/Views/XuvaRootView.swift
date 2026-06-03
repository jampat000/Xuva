import SwiftUI

public struct XuvaRootView: View {
    @StateObject private var store = XuvaClientStore()
    @StateObject private var watchlist = XuvaWatchlist()
    @Environment(\.scenePhase) private var scenePhase

    public init() {}

    public var body: some View {
        ZStack {
            // Solid background fills the entire screen including overscan area
            // so there's never a black band peeking through at the edges.
            XuvaTheme.background
                .ignoresSafeArea()
            XuvaTheme.backgroundWash
                .ignoresSafeArea()
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
                    #if !os(tvOS)
                    .controlSize(.large)
                    #endif
                    .tint(.white)
                    .padding(28)
                    .background(.black.opacity(0.48), in: RoundedRectangle(cornerRadius: 24, style: .continuous))
            }
            if let error = store.errorMessage, store.screen != .player {
                ErrorToast(message: error) { store.clearError() }
                    #if os(tvOS)
                    .padding(.top, 120)
                    #else
                    .padding(.top, 60)
                    #endif
                    .frame(maxHeight: .infinity, alignment: .top)
                    .transition(.move(edge: .top).combined(with: .opacity))
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .ignoresSafeArea()
        // Suppress the system focus halo (the giant white card-lift effect on
        // tvOS that was ballooning every focused button over its neighbours)
        // and rely on our own xuvaFocused() ring everywhere instead.
        .modifier(DisableSystemFocusEffect())
        .environmentObject(store)
        .environmentObject(watchlist)
        .preferredColorScheme(.dark)
        .task {
            await store.resumeSessionIfPossible()
            await store.autoConnectIfPossible()
        }
        .task(id: store.api?.baseURL) {
            // Three cases the watchlist needs to know about:
            //   - Fresh pair (api: nil → set): sync this account's items.
            //   - Unpair (api: set → nil): watchlist.didSet clears the cache
            //     so the next pairing doesn't show stale items.
            //   - Re-pair to a different server (baseURL changes): same as
            //     fresh pair; didSet drops the synced flag too.
            watchlist.api = store.api
            if store.api != nil {
                await watchlist.syncFromServer()
            }
        }
        .onReceive(NotificationCenter.default.publisher(for: deepLinkNotification)) { note in
            guard let kind = note.userInfo?["kind"] as? String,
                  let id = note.userInfo?["id"] as? String else { return }
            Task { await store.openDeepLink(kind: kind, id: id) }
        }
        .onChange(of: scenePhase) { _, newPhase in
            // Silently refresh home when the app returns to the foreground so
            // Continue Watching and pending-request rows always reflect live state.
            if newPhase == .active, store.connectionState == .paired {
                Task { await store.loadHome() }
            }
        }
    }
}

private let deepLinkNotification = Notification.Name("xuva.openDeepLink")

/// Wraps `.focusEffectDisabled()` in a back-compat guard. Available
/// iOS 17 / tvOS 17 / macOS 14+; on older targets just passes through.
private struct DisableSystemFocusEffect: ViewModifier {
    func body(content: Content) -> some View {
        if #available(iOS 17.0, tvOS 17.0, macOS 14.0, *) {
            content.focusEffectDisabled(true)
        } else {
            content
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
            #if os(tvOS)
            .buttonStyle(.card)
            #else
            .buttonStyle(.plain)
            #endif
            .foregroundStyle(XuvaTheme.mutedText)
        }
        .padding(.horizontal, 22)
        .padding(.vertical, 14)
        .background(XuvaTheme.surface.opacity(0.94), in: RoundedRectangle(cornerRadius: 16, style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: 16, style: .continuous).stroke(XuvaTheme.warn.opacity(0.45)))
        .shadow(color: .black.opacity(0.4), radius: 24, y: 12)
        .task {
            try? await Task.sleep(nanoseconds: 7_000_000_000)
            dismiss()
        }
    }
}
