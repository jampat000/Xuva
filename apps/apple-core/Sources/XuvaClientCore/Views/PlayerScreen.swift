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
                    player.play()
                    isPlaying = true
                }
                .onDisappear {
                    player.pause()
                    removeTimeObserver()
                }

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
                close: { Task { await stopAndClose() } },
                toggleInspector: { showInspector.toggle() }
            )
            .padding(.horizontal, 42)
            .padding(.bottom, 34)

            if showInspector {
                PlayerInspector(decision: playback.decision ?? playback.route?.decision)
                    .transition(.move(edge: .trailing).combined(with: .opacity))
                    .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .trailing)
                    .ignoresSafeArea(edges: .vertical)
            }
        }
        .background(.black)
        .task {
            while !Task.isCancelled {
                try? await Task.sleep(nanoseconds: UInt64(max(playback.heartbeatIntervalMs ?? 10_000, 5_000)) * 1_000_000)
                await sendHeartbeat()
            }
        }
    }

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

private struct PlayerChrome: View {
    let decision: PlaybackDecision?
    let currentSeconds: Double
    let durationSeconds: Double
    let isPlaying: Bool
    let togglePlay: () -> Void
    let skipBackward: () -> Void
    let skipForward: () -> Void
    let close: () -> Void
    let toggleInspector: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 20) {
            HStack(spacing: 14) {
                Button(action: close) {
                    Image(systemName: "chevron.left")
                }
                .buttonStyle(XuvaIconButtonStyle())

                RouteBadge(decision: decision)
                Spacer()
                MediaPill(text: "Quality", systemImage: "slider.horizontal.3", tint: XuvaTheme.secondaryText)
                MediaPill(text: "Audio", systemImage: "speaker.wave.2", tint: XuvaTheme.secondaryText)
                MediaPill(text: "Subtitles", systemImage: "captions.bubble", tint: XuvaTheme.secondaryText)
                Button(action: toggleInspector) {
                    Image(systemName: "info.circle")
                }
                .buttonStyle(XuvaIconButtonStyle())
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
        .padding(22)
        .background(.black.opacity(0.20), in: RoundedRectangle(cornerRadius: 10, style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: 10, style: .continuous).stroke(.white.opacity(0.10)))
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

private struct PlayerInspector: View {
    let decision: PlaybackDecision?

    var body: some View {
        VStack(alignment: .leading, spacing: 20) {
            VStack(alignment: .leading, spacing: 5) {
                Text("INSPECTOR")
                    .font(.caption.weight(.bold))
                    .tracking(2.8)
                    .foregroundStyle(XuvaTheme.muted)
                Text(decision?.badgeLabel ?? "Playback route")
                    .font(.title2.weight(.bold))
                    .foregroundStyle(XuvaTheme.text)
            }

            InspectorRow(label: "Container", value: decision?.containerAction)
            InspectorRow(label: "Video", value: decision?.videoAction)
            InspectorRow(label: "Audio", value: decision?.audioAction)
            InspectorRow(label: "Subtitles", value: decision?.subtitleAction)

            if let reason = decision?.reasonText ?? decision?.serverImpact, !reason.isEmpty {
                VStack(alignment: .leading, spacing: 8) {
                    Text("Reason")
                        .font(.caption.weight(.bold))
                        .tracking(1.8)
                        .foregroundStyle(XuvaTheme.muted)
                    Text(reason)
                        .font(.callout)
                        .foregroundStyle(XuvaTheme.secondaryText)
                        .lineLimit(5)
                }
                .padding(.top, 8)
            }
            Spacer()
        }
        .padding(.top, 76)
        .padding(.horizontal, 24)
        .frame(width: 360, alignment: .topLeading)
        .background(.black.opacity(0.90))
        .overlay(Rectangle().fill(.white.opacity(0.10)).frame(width: 1), alignment: .leading)
    }
}

private struct InspectorRow: View {
    let label: String
    let value: String?

    var body: some View {
        HStack {
            Text(label)
                .foregroundStyle(XuvaTheme.muted)
            Spacer()
            Text(displayValue)
                .font(.callout.monospaced().weight(.semibold))
                .foregroundStyle(XuvaTheme.text)
        }
        .padding(12)
        .background(.white.opacity(0.045), in: RoundedRectangle(cornerRadius: 8, style: .continuous))
    }

    private var displayValue: String {
        guard let value, !value.isEmpty else { return "Auto" }
        return value.replacingOccurrences(of: "_", with: " ").capitalized
    }
}
