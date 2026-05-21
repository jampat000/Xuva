import AVKit
import SwiftUI

public struct PlayerScreen: View {
    @EnvironmentObject private var store: XuvaClientStore

    public init() {}

    public var body: some View {
        if let playback = store.playback,
           let stream = playback.route?.streamURL,
           let url = store.api?.resolvedURL(stream) {
            XuvaVideoPlayer(url: url, authToken: store.api?.authToken, playback: playback) {
                store.closePlayer()
            }
        } else {
            VStack(spacing: 18) {
                Image(systemName: "exclamationmark.triangle.fill")
                    .font(.system(size: 54))
                    .foregroundStyle(XuvaTheme.warn)
                Text("Playback unavailable")
                    .font(.largeTitle.bold())
                Text(store.errorMessage ?? "The server did not return a playable direct or HLS route.")
                    .foregroundStyle(XuvaTheme.muted)
                    .multilineTextAlignment(.center)
                Button("Back") { store.closePlayer() }
                    .buttonStyle(XuvaSecondaryButtonStyle())
            }
            .padding(40)
        }
    }
}

struct XuvaVideoPlayer: View {
    @EnvironmentObject private var store: XuvaClientStore
    let url: URL
    let authToken: String?
    let playback: PlaybackStartResponse
    let close: () -> Void
    @State private var player: AVPlayer

    init(url: URL, authToken: String?, playback: PlaybackStartResponse, close: @escaping () -> Void) {
        self.url = url
        self.authToken = authToken
        self.playback = playback
        self.close = close
        _player = State(initialValue: AVPlayer(playerItem: Self.playerItem(url: url, authToken: authToken)))
    }

    var body: some View {
        ZStack(alignment: .topLeading) {
            VideoPlayer(player: player)
                .ignoresSafeArea()
                .onAppear { player.play() }
                .onDisappear { player.pause() }
            HStack(spacing: 14) {
                Button {
                    Task { await stopAndClose() }
                } label: {
                    Image(systemName: "chevron.left")
                }
                .buttonStyle(XuvaIconButtonStyle())
                RouteBadge(decision: playback.decision ?? playback.route?.decision)
            }
            .padding(32)
        }
        .background(.black)
    }

    private func stopAndClose() async {
        let seconds = CMTimeGetSeconds(player.currentTime())
        if seconds.isFinite {
            await store.stopPlayback(positionSeconds: max(0, Int(seconds)), completed: false)
        } else {
            close()
        }
    }

    private static func playerItem(url: URL, authToken: String?) -> AVPlayerItem {
        let token = authToken?.trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
        guard !token.isEmpty else {
            return AVPlayerItem(url: url)
        }
        let asset = AVURLAsset(url: url, options: [
            "AVURLAssetHTTPHeaderFieldsKey": ["X-Auth-Token": token]
        ])
        return AVPlayerItem(asset: asset)
    }
}
