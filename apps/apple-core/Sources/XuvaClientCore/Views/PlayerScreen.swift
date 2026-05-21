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
                Text(store.errorMessage ?? "Xuva could not prepare a playable stream.")
                    .foregroundStyle(XuvaTheme.muted)
                    .multilineTextAlignment(.center)
                Button("Back") { store.closePlayer() }
                    .buttonStyle(XuvaSecondaryButtonStyle())
            }
            .padding(40)
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .background(XuvaTheme.background)
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
    @State private var timeObserver: Any?
    @State private var currentSeconds: Double = 0
    @State private var durationSeconds: Double = 0
    @State private var isPlaying = true
    @State private var showInspector = false
    @State private var didFinish = false
    @State private var loadError: String?
    @State private var observers: [NSObjectProtocol] = []

    init(url: URL, authToken: String?, playback: PlaybackStartResponse, close: @escaping () -> Void) {
        self.url = url
        self.authToken = authToken
        self.playback = playback
        self.close = close
        _player = State(initialValue: AVPlayer(playerItem: Self.playerItem(url: url, authToken: authToken)))
    }

    var body: some View {
        ZStack(alignment: .bottom) {
            VideoPlayer(player: player)
                .ignoresSafeArea()
                .onAppear {
                    addTimeObserver()
                    addErrorObservers()
                    player.play()
                    isPlaying = true
                }
                .onDisappear {
                    player.pause()
                    removeTimeObserver()
                    removeErrorObservers()
                }

            if let err = loadError {
                ErrorOverlay(message: err, url: url, close: { Task { await stopAndClose() } })
            }

            #if !os(tvOS)
            if loadError == nil {
                customChrome
            }
            #endif
        }
        .background(.black)
        .task {
            while !Task.isCancelled {
                let intervalMs = max(playback.heartbeatIntervalMs ?? 10_000, 2_000)
                try? await Task.sleep(nanoseconds: UInt64(intervalMs) * 1_000_000)
                await sendHeartbeat()
            }
        }
    }

    private func addErrorObservers() {
        let center = NotificationCenter.default
        // Only treat real "could not play this stream" failures as fatal.
        // Transient error log entries (network retries, codec warm-up) are not
        // shown to the user — the player recovers and keeps playing.
        let failed = center.addObserver(forName: .AVPlayerItemFailedToPlayToEndTime, object: nil, queue: .main) { note in
            if let err = note.userInfo?[AVPlayerItemFailedToPlayToEndTimeErrorKey] as? NSError {
                loadError = "Playback failed: \(err.localizedDescription) (code \(err.code))"
            } else {
                loadError = "Playback failed to reach end of stream."
            }
        }
        observers = [failed]
    }

    private func removeErrorObservers() {
        for ob in observers { NotificationCenter.default.removeObserver(ob) }
        observers.removeAll()
    }

    #if !os(tvOS)
    private var customChrome: some View {
        ZStack(alignment: .bottom) {
            LinearGradient(
                colors: [.clear, .black.opacity(0.32), .black.opacity(0.82)],
                startPoint: .top,
                endPoint: .bottom
            )
            .frame(maxHeight: 360)
            .frame(maxHeight: .infinity, alignment: .bottom)
            .ignoresSafeArea()

            PlayerChrome(
                decision: playback.decision ?? playback.route?.decision,
                currentSeconds: currentSeconds,
                durationSeconds: durationSeconds,
                isPlaying: isPlaying,
                togglePlay: togglePlay,
                skipBackward: { skip(by: -10) },
                skipForward: { skip(by: 30) },
                close: { Task { await stopAndClose() } }
            )
            .padding(.horizontal, 28)
            .padding(.bottom, 28)
        }
    }
    #endif

    private func togglePlay() {
        if isPlaying {
            player.pause()
        } else {
            player.play()
        }
        isPlaying.toggle()
    }

    private func skip(by seconds: Double) {
        let next = max(0, min(durationSeconds > 0 ? durationSeconds : .greatestFiniteMagnitude, currentSeconds + seconds))
        player.seek(to: CMTime(seconds: next, preferredTimescale: 600))
    }

    private func addTimeObserver() {
        guard timeObserver == nil else { return }
        timeObserver = player.addPeriodicTimeObserver(forInterval: CMTime(seconds: 0.5, preferredTimescale: 600), queue: .main) { time in
            currentSeconds = max(0, CMTimeGetSeconds(time))
            if let duration = player.currentItem?.duration {
                let value = CMTimeGetSeconds(duration)
                if value.isFinite { durationSeconds = value }
            }
        }
    }

    private func removeTimeObserver() {
        if let timeObserver {
            player.removeTimeObserver(timeObserver)
            self.timeObserver = nil
        }
    }

    private func stopAndClose() async {
        let seconds = CMTimeGetSeconds(player.currentTime())
        if seconds.isFinite {
            await store.stopPlayback(positionSeconds: max(0, Int(seconds)), completed: false)
        } else {
            close()
        }
    }

    private func sendHeartbeat() async {
        guard let heartbeatUrl = playback.heartbeatUrl else { return }
        let seconds = CMTimeGetSeconds(player.currentTime())
        guard seconds.isFinite else { return }
        try? await store.api?.heartbeat(path: heartbeatUrl, positionSeconds: max(0, Int(seconds)), isPaused: !isPlaying)
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

struct ErrorOverlay: View {
    let message: String
    let url: URL
    let close: () -> Void

    var body: some View {
        VStack(spacing: 18) {
            Image(systemName: "exclamationmark.triangle.fill")
                .font(.system(size: 60))
                .foregroundStyle(XuvaTheme.warn)
            Text("Playback failed")
                .font(.system(size: 32, weight: .bold))
                .foregroundStyle(XuvaTheme.text)
            Text(message)
                .font(.system(size: 18))
                .foregroundStyle(XuvaTheme.secondaryText)
                .multilineTextAlignment(.center)
                .frame(maxWidth: 720)
            Text(url.absoluteString)
                .font(.system(size: 13, design: .monospaced))
                .foregroundStyle(XuvaTheme.mutedText)
                .lineLimit(2)
                .truncationMode(.middle)
                .frame(maxWidth: 720)
            Button("Back to detail") { close() }
                .buttonStyle(XuvaSecondaryButtonStyle())
                .padding(.top, 12)
        }
        .padding(40)
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(Color.black.opacity(0.85))
    }
}

#if !os(tvOS)
private struct PlayerChrome: View {
    let decision: PlaybackDecision?
    let currentSeconds: Double
    let durationSeconds: Double
    let isPlaying: Bool
    let togglePlay: () -> Void
    let skipBackward: () -> Void
    let skipForward: () -> Void
    let close: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 18) {
            HStack(spacing: 12) {
                Button(action: close) {
                    Image(systemName: "chevron.left")
                }
                .buttonStyle(XuvaIconButtonStyle())

                RouteBadge(decision: decision)
                Spacer()
            }

            VStack(spacing: 10) {
                GeometryReader { proxy in
                    ZStack(alignment: .leading) {
                        Capsule().fill(.white.opacity(0.20))
                        Capsule()
                            .fill(XuvaTheme.text)
                            .frame(width: proxy.size.width * progress)
                    }
                }
                .frame(height: 5)

                HStack {
                    Text(format(currentSeconds))
                    Spacer()
                    Text(durationSeconds > 0 ? "-\(format(max(0, durationSeconds - currentSeconds)))" : "--:--")
                }
                .font(.caption.monospacedDigit().weight(.semibold))
                .foregroundStyle(.white.opacity(0.72))
            }

            HStack(spacing: 16) {
                Button(action: skipBackward) {
                    Image(systemName: "gobackward.10")
                }
                .buttonStyle(XuvaIconButtonStyle())
                Button(action: togglePlay) {
                    Image(systemName: isPlaying ? "pause.fill" : "play.fill")
                        .font(.title2.weight(.bold))
                }
                .buttonStyle(XuvaIconButtonStyle())
                Button(action: skipForward) {
                    Image(systemName: "goforward.30")
                }
                .buttonStyle(XuvaIconButtonStyle())
            }
        }
        .padding(20)
        .background(.black.opacity(0.30), in: RoundedRectangle(cornerRadius: 18, style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: 18, style: .continuous).stroke(.white.opacity(0.10)))
    }

    private var progress: CGFloat {
        guard durationSeconds > 0 else { return 0 }
        return CGFloat(min(max(currentSeconds / durationSeconds, 0), 1))
    }

    private func format(_ seconds: Double) -> String {
        guard seconds.isFinite else { return "--:--" }
        let value = max(0, Int(seconds.rounded()))
        let hours = value / 3600
        let minutes = (value % 3600) / 60
        let secs = value % 60
        if hours > 0 { return "\(hours):\(String(format: "%02d", minutes)):\(String(format: "%02d", secs))" }
        return "\(minutes):\(String(format: "%02d", secs))"
    }
}
#endif
