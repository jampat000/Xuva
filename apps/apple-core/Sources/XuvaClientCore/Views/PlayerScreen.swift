import AVKit
import SwiftUI
import os

private let logger = Logger(subsystem: "com.xuva.client", category: "player")

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
            .ignoresSafeArea()
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .background(Color.black)
            #if os(tvOS)
            .onExitCommand {
                Task { await store.stopPlayback() }
            }
            #endif
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

#if os(tvOS)
struct XuvaVideoPlayer: UIViewControllerRepresentable {
    @EnvironmentObject private var store: XuvaClientStore
    let url: URL
    let authToken: String?
    let playback: PlaybackStartResponse
    let close: () -> Void

    func makeCoordinator() -> Coordinator {
        Coordinator(playback: playback, close: close, sendHeartbeat: { pos, paused in
            await sendHeartbeat(positionSeconds: pos, isPaused: paused, store: store)
        }, stopPlayback: { pos in
            await store.stopPlayback(positionSeconds: pos, completed: false)
        })
    }

    func makeUIViewController(context: Context) -> AVPlayerViewController {
        // tvOS plays nicely with a .playback audio session — without this,
        // background audio routing can cut out when the AVPlayerViewController's
        // own session interacts with the SwiftUI host scene.
        do {
            try AVAudioSession.sharedInstance().setCategory(.playback, mode: .moviePlayback, options: [])
            try AVAudioSession.sharedInstance().setActive(true, options: [])
        } catch {
            logger.error("AVAudioSession setup failed: \(error)")
        }

        let item = makePlayerItem(url: url, authToken: authToken)
        let player = AVPlayer(playerItem: item)
        player.automaticallyWaitsToMinimizeStalling = true

        let vc = AVPlayerViewController()
        vc.player = player
        vc.allowsPictureInPicturePlayback = true
        vc.showsPlaybackControls = true
        vc.requiresLinearPlayback = false
        vc.delegate = context.coordinator
        vc.view.backgroundColor = .black
        context.coordinator.attach(to: player, resumeAt: playback.clientStartPositionSeconds)
        player.play()
        logger.debug("AVPlayerViewController created resume=\(playback.clientStartPositionSeconds ?? 0)")
        return vc
    }

    func updateUIViewController(_ vc: AVPlayerViewController, context: Context) {}

    static func dismantleUIViewController(_ vc: AVPlayerViewController, coordinator: Coordinator) {
        vc.player?.pause()
        coordinator.detach()
    }

    @MainActor
    func sendHeartbeat(positionSeconds: Int, isPaused: Bool, store: XuvaClientStore) async {
        guard let heartbeatUrl = playback.heartbeatUrl else { return }
        try? await store.api?.heartbeat(path: heartbeatUrl, positionSeconds: positionSeconds, isPaused: isPaused)
    }

    final class Coordinator: NSObject, AVPlayerViewControllerDelegate {
        let playback: PlaybackStartResponse
        let close: () -> Void
        let sendHeartbeat: (Int, Bool) async -> Void
        let stopPlayback: (Int) async -> Void
        private weak var player: AVPlayer?
        private var heartbeatTask: Task<Void, Never>?
        private var errorObserver: NSObjectProtocol?
        private var statusObservation: NSKeyValueObservation?
        private var timeControlObservation: NSKeyValueObservation?

        init(playback: PlaybackStartResponse, close: @escaping () -> Void, sendHeartbeat: @escaping (Int, Bool) async -> Void, stopPlayback: @escaping (Int) async -> Void) {
            self.playback = playback
            self.close = close
            self.sendHeartbeat = sendHeartbeat
            self.stopPlayback = stopPlayback
        }

        func attach(to player: AVPlayer, resumeAt seconds: Int? = nil) {
            self.player = player
            heartbeatTask = Task { [weak self] in
                while !Task.isCancelled {
                    let intervalMs = max(self?.playback.heartbeatIntervalMs ?? 10_000, 2_000)
                    try? await Task.sleep(nanoseconds: UInt64(intervalMs) * 1_000_000)
                    guard let self, let p = self.player else { break }
                    let seconds = CMTimeGetSeconds(p.currentTime())
                    if seconds.isFinite {
                        let paused = p.timeControlStatus != .playing
                        await self.sendHeartbeat(max(0, Int(seconds)), paused)
                    }
                }
            }
            errorObserver = NotificationCenter.default.addObserver(forName: .AVPlayerItemFailedToPlayToEndTime, object: nil, queue: .main) { note in
                if let err = note.userInfo?[AVPlayerItemFailedToPlayToEndTimeErrorKey] as? NSError {
                    logger.error("AVPlayer FAILED: \(err) code=\(err.code)")
                }
            }
            statusObservation = player.currentItem?.observe(\.status, options: [.new]) { [weak self] item, _ in
                guard item.status == .readyToPlay else { return }
                if let s = seconds, s > 0 {
                    let target = CMTime(seconds: Double(s), preferredTimescale: 600)
                    item.seek(to: target, toleranceBefore: .positiveInfinity, toleranceAfter: .positiveInfinity) { _ in
                        logger.debug("resume seek to \(s)s done")
                    }
                }
                self?.applySubtitleSelection(to: item)
                self?.statusObservation?.invalidate()
                self?.statusObservation = nil
            }
            // Log stall events so we can diagnose seek-induced buffer exhaustion.
            // AVPlayer pauses automatically (automaticallyWaitsToMinimizeStalling)
            // and resumes once the buffer refills — no manual recovery needed.
            timeControlObservation = player.observe(\.timeControlStatus, options: [.new]) { p, _ in
                if p.timeControlStatus == .waitingToPlayAtSpecifiedRate {
                    logger.debug("AVPlayer stalling reason=\(p.reasonForWaitingToPlay?.rawValue ?? "unknown")")
                }
            }
        }

        func detach() {
            heartbeatTask?.cancel()
            heartbeatTask = nil
            if let errorObserver { NotificationCenter.default.removeObserver(errorObserver) }
            errorObserver = nil
            statusObservation?.invalidate()
            statusObservation = nil
            timeControlObservation?.invalidate()
            timeControlObservation = nil
        }

        private func applySubtitleSelection(to item: AVPlayerItem) {
            let track = playback.clientSubtitleTrack
            // Use the async API so the asset's track metadata is guaranteed loaded.
            Task { @MainActor in
                do {
                    guard let group = try await item.asset.loadMediaSelectionGroup(for: .legible) else {
                        return
                    }
                    if track == nil {
                        item.select(nil, in: group)
                        return
                    }
                    let option = group.options.first { opt in
                        if let lang = track?.language, let locale = opt.locale {
                            return locale.languageCode == lang || locale.identifier.hasPrefix(lang)
                        }
                        return false
                    } ?? group.options.first { opt in
                        guard let title = track?.title, !title.isEmpty else { return false }
                        return opt.displayName.localizedCaseInsensitiveContains(title)
                    }
                    if let option {
                        item.select(option, in: group)
                    }
                } catch {
                    logger.error("failed to load legible media selection group: \(error)")
                }
            }
        }

        func playerViewControllerDidEndDismissalTransition(_ playerViewController: AVPlayerViewController) {
            // Report the final position to the server so it can save progress
            // and close the session. stopPlayback() calls closePlayer() internally
            // when there is no stopUrl (trailers), so close() is not needed here.
            let currentSeconds: Int = {
                guard let p = playerViewController.player else { return 0 }
                let t = CMTimeGetSeconds(p.currentTime())
                return t.isFinite ? max(0, Int(t)) : 0
            }()
            Task { await stopPlayback(currentSeconds) }
        }
    }
}
#else
// MARK: - iOS / iPadOS custom player
//
// A SwiftUI player that ports the feature set of the web Player.svelte — the
// cleanest Xuva player — to AVPlayer:
//   • scrubbable seek bar with drag-to-seek and chapter markers
//   • Skip Intro button + Credits marker (server chapter detection)
//   • in-player subtitle + audio track menus (AVMediaSelectionGroup)
//   • ±10 / +30 skip with on-screen toast, double-tap-to-skip zones
//   • resume toast, loading spinner, auto-hiding chrome
//   • progress heartbeat + final playback-state write on exit
struct XuvaVideoPlayer: View {
    @EnvironmentObject private var store: XuvaClientStore
    let url: URL
    let authToken: String?
    let playback: PlaybackStartResponse
    let close: () -> Void

    @State private var player: AVPlayer
    @State private var timeObserver: Any?
    @State private var statusObservation: NSKeyValueObservation?
    @State private var observers: [NSObjectProtocol] = []

    @State private var currentSeconds: Double = 0
    @State private var durationSeconds: Double = 0
    @State private var isPlaying = true
    @State private var loading = true
    @State private var loadError: String?

    @State private var controlsVisible = true
    @State private var hideTask: Task<Void, Never>?

    @State private var seekToast: String?
    @State private var seekToastTask: Task<Void, Never>?
    @State private var resumeToast: String?

    @State private var isScrubbing = false
    @State private var scrubSeconds: Double = 0

    @State private var chapters: ChaptersResponse?
    @State private var skipIntroDismissed = false

    @State private var audioGroup: AVMediaSelectionGroup?
    @State private var legibleGroup: AVMediaSelectionGroup?
    @State private var showAudioMenu = false
    @State private var showSubMenu = false
    @State private var menuRefresh = 0

    init(url: URL, authToken: String?, playback: PlaybackStartResponse, close: @escaping () -> Void) {
        self.url = url
        self.authToken = authToken
        self.playback = playback
        self.close = close
        _player = State(initialValue: AVPlayer(playerItem: makePlayerItem(url: url, authToken: authToken)))
    }

    private var displaySeconds: Double { isScrubbing ? scrubSeconds : currentSeconds }

    private var showSkipIntro: Bool {
        guard let intro = chapters?.intro, !skipIntroDismissed else { return false }
        return currentSeconds >= intro.start && currentSeconds <= intro.end
    }

    private var showCredits: Bool {
        guard let credits = chapters?.credits else { return false }
        return currentSeconds >= credits.start
    }

    var body: some View {
        ZStack {
            VideoPlayer(player: player)
                .ignoresSafeArea()
                .onAppear {
                    logger.debug("VideoPlayer.onAppear")
                    addTimeObserver()
                    addObservers()
                    if let s = playback.clientStartPositionSeconds, s > 0 {
                        showResumeToast(s)
                    }
                    player.play()
                    isPlaying = true
                    scheduleHide()
                }
                .onDisappear {
                    player.pause()
                    removeTimeObserver()
                    removeObservers()
                }

            if let err = loadError {
                ErrorOverlay(message: err, url: url, close: { Task { await stopAndClose() } })
            } else {
                tapLayer
                if loading {
                    ProgressView()
                        .progressViewStyle(.circular)
                        .tint(.white)
                        .scaleEffect(1.5)
                }
                overlays
                if controlsVisible { customChrome }
            }
        }
        .background(.black)
        .animation(.easeInOut(duration: 0.2), value: controlsVisible)
        .task {
            // Progress heartbeat loop
            while !Task.isCancelled {
                let intervalMs = max(playback.heartbeatIntervalMs ?? 10_000, 2_000)
                try? await Task.sleep(nanoseconds: UInt64(intervalMs) * 1_000_000)
                await sendHeartbeat()
            }
        }
        .task { await loadChaptersAndTracks() }
    }

    // MARK: Tap / gesture layer (single tap toggles chrome; double-tap skips)

    private var tapLayer: some View {
        HStack(spacing: 0) {
            tapZone(skip: -10)
            tapZone(skip: 30)
        }
        .ignoresSafeArea()
    }

    private func tapZone(skip seconds: Double) -> some View {
        Color.clear
            .contentShape(Rectangle())
            .onTapGesture(count: 2) { skip(by: seconds) }
            .onTapGesture(count: 1) { toggleControls() }
    }

    // MARK: Overlays (toasts, skip-intro, credits)

    private var overlays: some View {
        ZStack {
            if let toast = seekToast {
                Text(toast)
                    .font(.system(size: 28, weight: .bold, design: .rounded))
                    .foregroundStyle(.white)
                    .padding(.horizontal, 22)
                    .padding(.vertical, 12)
                    .background(.black.opacity(0.55), in: Capsule())
                    .transition(.opacity)
            }

            VStack {
                Spacer()
                HStack {
                    Spacer()
                    if showSkipIntro {
                        Button {
                            skipIntro()
                        } label: {
                            Text("Skip Intro")
                                .font(.system(size: 15, weight: .semibold))
                                .padding(.horizontal, 20)
                                .padding(.vertical, 10)
                        }
                        .buttonStyle(XuvaSecondaryButtonStyle())
                        .transition(.opacity)
                    } else if showCredits {
                        Text("Credits")
                            .font(.system(size: 14, weight: .medium))
                            .foregroundStyle(.white.opacity(0.85))
                            .padding(.horizontal, 18)
                            .padding(.vertical, 9)
                            .background(.black.opacity(0.5), in: RoundedRectangle(cornerRadius: 8, style: .continuous))
                            .overlay(RoundedRectangle(cornerRadius: 8, style: .continuous).stroke(.white.opacity(0.3)))
                    }
                }
                .padding(.trailing, 28)
                .padding(.bottom, controlsVisible ? 150 : 36)
            }

            if let resumeToast {
                VStack {
                    Spacer()
                    Text(resumeToast)
                        .font(.system(size: 14, weight: .medium))
                        .foregroundStyle(.white.opacity(0.9))
                        .padding(.horizontal, 18)
                        .padding(.vertical, 9)
                        .background(.black.opacity(0.55), in: Capsule())
                        .padding(.bottom, controlsVisible ? 150 : 36)
                }
                .transition(.opacity)
            }
        }
        .animation(.easeInOut(duration: 0.2), value: seekToast)
        .animation(.easeInOut(duration: 0.2), value: resumeToast)
        .animation(.easeInOut(duration: 0.2), value: showSkipIntro)
    }

    // MARK: Chrome (top bar + seek bar + controls)

    private var customChrome: some View {
        ZStack {
            LinearGradient(
                colors: [.black.opacity(0.55), .clear, .clear, .black.opacity(0.82)],
                startPoint: .top,
                endPoint: .bottom
            )
            .ignoresSafeArea()
            .allowsHitTesting(false)

            VStack {
                // Top bar
                HStack(spacing: 12) {
                    Button(action: { Task { await stopAndClose() } }) {
                        Image(systemName: "chevron.left")
                    }
                    .buttonStyle(XuvaIconButtonStyle())
                    RouteBadge(decision: playback.decision ?? playback.route?.decision)
                    Spacer()
                    trackButtons
                }
                .padding(.horizontal, 24)
                .padding(.top, 20)

                Spacer()

                // Seek bar + transport
                VStack(spacing: 14) {
                    SeekBar(
                        current: displaySeconds,
                        duration: durationSeconds,
                        chapters: chapters,
                        isScrubbing: $isScrubbing,
                        onScrub: { secs in
                            scrubSeconds = secs
                            keepVisible()
                        },
                        onCommit: { secs in
                            seek(to: secs)
                            scheduleHide()
                        }
                    )
                    HStack {
                        Text(formatTime(displaySeconds))
                        Spacer()
                        Text(durationSeconds > 0 ? "-\(formatTime(max(0, durationSeconds - displaySeconds)))" : "--:--")
                    }
                    .font(.caption.monospacedDigit().weight(.semibold))
                    .foregroundStyle(.white.opacity(0.72))

                    HStack(spacing: 26) {
                        Button(action: { skip(by: -10) }) { Image(systemName: "gobackward.10") }
                            .buttonStyle(XuvaIconButtonStyle())
                        Button(action: togglePlay) {
                            Image(systemName: isPlaying ? "pause.fill" : "play.fill")
                        }
                        .buttonStyle(XuvaIconButtonStyle())
                        Button(action: { skip(by: 30) }) { Image(systemName: "goforward.30") }
                            .buttonStyle(XuvaIconButtonStyle())
                    }
                    .padding(.top, 2)
                }
                .padding(.horizontal, 28)
                .padding(.bottom, 28)
            }
        }
        .overlay(alignment: .topTrailing) {
            if showSubMenu {
                trackMenu(title: "Subtitles", group: legibleGroup, allowsOff: true)
            } else if showAudioMenu {
                trackMenu(title: "Audio", group: audioGroup, allowsOff: false)
            }
        }
    }

    private var trackButtons: some View {
        HStack(spacing: 12) {
            if let legibleGroup, !legibleGroup.options.isEmpty {
                Button(action: {
                    showSubMenu.toggle()
                    showAudioMenu = false
                    keepVisible()
                }) {
                    Image(systemName: "captions.bubble")
                }
                .buttonStyle(XuvaIconButtonStyle())
            }
            if let audioGroup, audioGroup.options.count > 1 {
                Button(action: {
                    showAudioMenu.toggle()
                    showSubMenu = false
                    keepVisible()
                }) {
                    Image(systemName: "waveform")
                }
                .buttonStyle(XuvaIconButtonStyle())
            }
        }
    }

    private func trackMenu(title: String, group: AVMediaSelectionGroup?, allowsOff: Bool) -> some View {
        let item: AVPlayerItem? = player.currentItem
        let selected: AVMediaSelectionOption? = {
            guard let group, let item else { return nil }
            return item.currentMediaSelection.selectedMediaOption(in: group)
        }()
        _ = menuRefresh // re-render when selection changes
        return VStack(alignment: .leading, spacing: 0) {
            Text(title)
                .font(.system(size: 13, weight: .semibold))
                .foregroundStyle(.white.opacity(0.6))
                .padding(.horizontal, 16)
                .padding(.vertical, 10)
            if allowsOff {
                trackRow(label: "Off", isSelected: selected == nil) {
                    if let group { player.currentItem?.select(nil, in: group) }
                    closeMenus()
                }
            }
            ForEach(Array((group?.options ?? []).enumerated()), id: \.offset) { _, option in
                trackRow(label: option.displayName, isSelected: option == selected) {
                    if let group { player.currentItem?.select(option, in: group) }
                    closeMenus()
                }
            }
        }
        .frame(width: 280)
        .padding(.vertical, 6)
        .background(.black.opacity(0.85), in: RoundedRectangle(cornerRadius: 14, style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: 14, style: .continuous).stroke(.white.opacity(0.12)))
        .padding(.top, 72)
        .padding(.trailing, 24)
    }

    private func trackRow(label: String, isSelected: Bool, action: @escaping () -> Void) -> some View {
        Button(action: action) {
            HStack {
                Text(label)
                    .font(.system(size: 15))
                    .foregroundStyle(.white)
                Spacer()
                if isSelected {
                    Image(systemName: "checkmark")
                        .font(.system(size: 13, weight: .bold))
                        .foregroundStyle(Color.white)
                }
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 11)
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
    }

    private func closeMenus() {
        showSubMenu = false
        showAudioMenu = false
        menuRefresh += 1
        scheduleHide()
    }

    // MARK: Controls visibility

    private func toggleControls() {
        if controlsVisible {
            controlsVisible = false
            hideTask?.cancel()
        } else {
            controlsVisible = true
            scheduleHide()
        }
    }

    private func keepVisible() {
        controlsVisible = true
        hideTask?.cancel()
    }

    private func scheduleHide() {
        hideTask?.cancel()
        guard isPlaying else { return }
        hideTask = Task {
            try? await Task.sleep(nanoseconds: 3_000_000_000)
            guard !Task.isCancelled else { return }
            await MainActor.run {
                if !showSubMenu && !showAudioMenu { controlsVisible = false }
            }
        }
    }

    // MARK: Transport

    private func togglePlay() {
        if isPlaying {
            player.pause()
            keepVisible()
        } else {
            player.play()
            scheduleHide()
        }
        isPlaying.toggle()
    }

    private func skip(by seconds: Double) {
        let next = max(0, min(durationSeconds > 0 ? durationSeconds : .greatestFiniteMagnitude, currentSeconds + seconds))
        seek(to: next)
        showSeekToast(seconds)
        keepVisible()
        scheduleHide()
    }

    private func seek(to seconds: Double) {
        // Cancel any in-flight seek before issuing a new one so rapid taps
        // don't queue competing seeks that fight for the same playhead.
        player.currentItem?.cancelPendingSeeks()
        player.seek(to: CMTime(seconds: seconds, preferredTimescale: 600))
        currentSeconds = seconds
        if let intro = chapters?.intro, seconds < intro.start { skipIntroDismissed = false }
    }

    private func skipIntro() {
        guard let intro = chapters?.intro else { return }
        seek(to: intro.end)
        skipIntroDismissed = true
    }

    private func showSeekToast(_ seconds: Double) {
        let n = Int(seconds)
        seekToast = n > 0 ? "+\(n)s" : "\(n)s"
        seekToastTask?.cancel()
        seekToastTask = Task {
            try? await Task.sleep(nanoseconds: 800_000_000)
            guard !Task.isCancelled else { return }
            await MainActor.run { seekToast = nil }
        }
    }

    private func showResumeToast(_ seconds: Int) {
        resumeToast = "Resuming from \(formatTime(Double(seconds)))"
        Task {
            try? await Task.sleep(nanoseconds: 3_000_000_000)
            await MainActor.run { resumeToast = nil }
        }
    }

    // MARK: Observers

    private func addObservers() {
        let center = NotificationCenter.default
        let failed = center.addObserver(forName: .AVPlayerItemFailedToPlayToEndTime, object: nil, queue: .main) { note in
            if let err = note.userInfo?[AVPlayerItemFailedToPlayToEndTimeErrorKey] as? NSError {
                logger.error("AVPlayer FAILED: \(err) code=\(err.code)")
                loadError = "Playback failed: \(err.localizedDescription) (code \(err.code))"
            } else {
                loadError = "Playback failed to reach end of stream."
            }
        }
        let ended = center.addObserver(forName: .AVPlayerItemDidPlayToEndTime, object: player.currentItem, queue: .main) { _ in
            isPlaying = false
            keepVisible()
            Task { await writeFinalState(completed: true) }
        }
        observers = [failed, ended]

        // Resolve resume seek + track groups once the item is ready.
        statusObservation = player.currentItem?.observe(\.status, options: [.new]) { item, _ in
            guard item.status == .readyToPlay else { return }
            loading = false
            if let s = playback.clientStartPositionSeconds, s > 0 {
                let target = CMTime(seconds: Double(s), preferredTimescale: 600)
                item.seek(to: target, toleranceBefore: .positiveInfinity, toleranceAfter: .positiveInfinity)
            }
            Task { await loadSelectionGroups(item: item) }
            statusObservation?.invalidate()
            statusObservation = nil
        }
    }

    private func removeObservers() {
        for ob in observers { NotificationCenter.default.removeObserver(ob) }
        observers.removeAll()
        statusObservation?.invalidate()
        statusObservation = nil
        hideTask?.cancel()
        seekToastTask?.cancel()
    }

    private func addTimeObserver() {
        guard timeObserver == nil else { return }
        timeObserver = player.addPeriodicTimeObserver(forInterval: CMTime(seconds: 0.5, preferredTimescale: 600), queue: .main) { time in
            if !isScrubbing {
                currentSeconds = max(0, CMTimeGetSeconds(time))
            }
            if let duration = player.currentItem?.duration {
                let value = CMTimeGetSeconds(duration)
                if value.isFinite { durationSeconds = value }
            }
            if loading, player.timeControlStatus == .playing { loading = false }
        }
    }

    private func removeTimeObserver() {
        if let timeObserver {
            player.removeTimeObserver(timeObserver)
            self.timeObserver = nil
        }
    }

    // MARK: Track groups

    @MainActor
    private func loadSelectionGroups(item: AVPlayerItem) async {
        if let audio = try? await item.asset.loadMediaSelectionGroup(for: .audible) {
            audioGroup = audio
        }
        if let legible = try? await item.asset.loadMediaSelectionGroup(for: .legible) {
            legibleGroup = legible
            applyInitialSubtitle(group: legible, item: item)
        }
    }

    private func applyInitialSubtitle(group: AVMediaSelectionGroup, item: AVPlayerItem) {
        let track = playback.clientSubtitleTrack
        guard let track else {
            item.select(nil, in: group)
            return
        }
        let option = group.options.first { opt in
            if let lang = track.language, let locale = opt.locale {
                return locale.languageCode == lang || locale.identifier.hasPrefix(lang)
            }
            return false
        } ?? group.options.first { opt in
            guard let title = track.title, !title.isEmpty else { return false }
            return opt.displayName.localizedCaseInsensitiveContains(title)
        }
        if let option { item.select(option, in: group) }
    }

    // MARK: Chapters

    private func loadChaptersAndTracks() async {
        guard let api = store.api, let msid = playback.mediaSourceId, !msid.isEmpty else { return }
        if let ch = try? await api.chapters(mediaSourceId: msid) {
            chapters = ch
        }
    }

    // MARK: Heartbeat / stop

    private func stopAndClose() async {
        let seconds = CMTimeGetSeconds(player.currentTime())
        await writeFinalState(completed: false)
        if seconds.isFinite {
            await store.stopPlayback(positionSeconds: max(0, Int(seconds)), completed: false)
        } else {
            close()
        }
    }

    private func writeFinalState(completed: Bool) async {
        guard let api = store.api, let msid = playback.mediaSourceId, !msid.isEmpty else { return }
        let pos = CMTimeGetSeconds(player.currentTime())
        guard pos.isFinite else { return }
        let dur = durationSeconds.isFinite ? durationSeconds : 0
        try? await api.setPlaybackState(
            mediaSourceId: msid,
            update: PlaybackStateUpdate(progressSeconds: max(0, pos), durationSeconds: dur, watched: completed ? true : nil)
        )
    }

    private func sendHeartbeat() async {
        guard let heartbeatUrl = playback.heartbeatUrl else { return }
        let seconds = CMTimeGetSeconds(player.currentTime())
        guard seconds.isFinite else { return }
        try? await store.api?.heartbeat(path: heartbeatUrl, positionSeconds: max(0, Int(seconds)), isPaused: !isPlaying)
    }
}

// MARK: - Scrub bar

private struct SeekBar: View {
    let current: Double
    let duration: Double
    let chapters: ChaptersResponse?
    @Binding var isScrubbing: Bool
    let onScrub: (Double) -> Void
    let onCommit: (Double) -> Void

    private var progress: CGFloat {
        guard duration > 0 else { return 0 }
        return CGFloat(min(max(current / duration, 0), 1))
    }

    var body: some View {
        GeometryReader { proxy in
            let width = proxy.size.width
            ZStack(alignment: .leading) {
                Capsule().fill(.white.opacity(0.20))
                Capsule().fill(.white)
                    .frame(width: width * progress)
                // Chapter markers
                if let credits = chapters?.credits, duration > 0 {
                    marker(at: credits.start, width: width)
                }
                if let intro = chapters?.intro, duration > 0, intro.start > 0 {
                    marker(at: intro.start, width: width)
                }
                Circle()
                    .fill(.white)
                    .frame(width: isScrubbing ? 18 : 13, height: isScrubbing ? 18 : 13)
                    .shadow(radius: 2)
                    .offset(x: width * progress - (isScrubbing ? 9 : 6.5))
            }
            .frame(height: 6)
            .frame(maxHeight: .infinity, alignment: .center)
            .contentShape(Rectangle())
            .gesture(
                DragGesture(minimumDistance: 0)
                    .onChanged { value in
                        isScrubbing = true
                        let ratio = max(0, min(1, value.location.x / width))
                        onScrub(ratio * duration)
                    }
                    .onEnded { value in
                        let ratio = max(0, min(1, value.location.x / width))
                        isScrubbing = false
                        onCommit(ratio * duration)
                    }
            )
        }
        .frame(height: 22)
    }

    private func marker(at seconds: Double, width: CGFloat) -> some View {
        RoundedRectangle(cornerRadius: 1)
            .fill(.white.opacity(0.55))
            .frame(width: 3, height: 12)
            .offset(x: width * CGFloat(min(max(seconds / duration, 0), 1)) - 1.5)
    }
}
#endif

/// Build an AVPlayerItem with an optional X-Auth-Token forwarded on the
/// initial playlist/segment request. AVAssetResourceLoader sub-requests for
/// HLS sub-segments don't always inherit this, so we rely on signed query
/// tokens for fan-out — the header here is principal auth.
private func makePlayerItem(url: URL, authToken: String?) -> AVPlayerItem {
    let token = authToken?.trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
    let item: AVPlayerItem
    if token.isEmpty {
        item = AVPlayerItem(url: url)
    } else {
        let asset = AVURLAsset(url: url, options: [
            "AVURLAssetHTTPHeaderFieldsKey": ["X-Auth-Token": token]
        ])
        item = AVPlayerItem(asset: asset)
    }
    // 30 s forward buffer: seeks into already-buffered regions are instant
    // and there's headroom to recover from an aggressive forward scrub without
    // immediately exhausting the download pipeline and stalling.
    item.preferredForwardBufferDuration = 30
    return item
}

/// Shared time formatter for player chrome.
private func formatTime(_ seconds: Double) -> String {
    guard seconds.isFinite else { return "--:--" }
    let value = max(0, Int(seconds.rounded()))
    let hours = value / 3600
    let minutes = (value % 3600) / 60
    let secs = value % 60
    if hours > 0 { return "\(hours):\(String(format: "%02d", minutes)):\(String(format: "%02d", secs))" }
    return "\(minutes):\(String(format: "%02d", secs))"
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
