import Foundation
import SwiftUI

@MainActor
public final class XuvaClientStore: ObservableObject {
    @Published public var serverText = "http://127.0.0.1:8097" {
        didSet {
            UserDefaults.standard.set(serverText, forKey: Self.serverURLKey)
        }
    }
    @Published public var bootstrap: BootstrapResponse?
    @Published public var pairing: PairingResponse?
    @Published public var home: ClientHomeResponse?
    @Published public var selectedDetail: DetailResponse?
    @Published public var playback: PlaybackStartResponse?
    @Published public var isBusy = false
    @Published public var errorMessage: String?
    @Published public var screen: XuvaScreen = .connect
    @Published public var connectionState: XuvaConnectionState = .idle
    @Published public var activeSection: String = "Home"
    @Published public var heroIndex: Int = 0
    /// When the user opens a Continue Watching tile, we hold onto the
    /// progress fraction here so the next play() can resume at the right
    /// position. Cleared after use.
    @Published public var pendingResumeFraction: Double?
    /// For Continue Watching tiles the home item's id is the mediaSource id;
    /// the detail endpoint needs the parent movie/series id. We open detail
    /// for the parent but remember which version to actually start.
    @Published public var pendingResumeMediaSourceId: String?

    public private(set) var api: XuvaAPI?
    public let deviceId: String
    public var usesDefaultServerURL: Bool {
        serverText == Self.defaultServerURL
    }

    private static let defaultServerURL = "http://127.0.0.1:8097"
    private static let deviceIDKey = "xuva.apple.deviceId"
    private static let serverURLKey = "xuva.apple.serverURL"
    private static let pairedServerURLKey = "xuva.apple.pairedServerURL"
    private static let pairedDeviceKey = "xuva.apple.pairedDevice"
    private static let authTokenKey = "xuva.apple.authToken"

    public init() {
        if let existingServer = UserDefaults.standard.string(forKey: Self.serverURLKey), !existingServer.isEmpty {
            serverText = existingServer
        }

        if let existing = UserDefaults.standard.string(forKey: Self.deviceIDKey) {
            deviceId = existing
        } else {
            let generated = UUID().uuidString
            UserDefaults.standard.set(generated, forKey: Self.deviceIDKey)
            deviceId = generated
        }

        if UserDefaults.standard.bool(forKey: Self.pairedDeviceKey) {
            connectionState = .paired
            // Skip the connect screen entirely — go straight to home and load
            // data in the background. The spinner overlay covers the empty state
            // until resumeSessionIfPossible() completes.
            screen = .home
        }
    }

    public func connect() async {
        await run {
            guard let url = URL(string: normalizedServerURL()) else { throw XuvaAPIError.invalidURL }
            let nextAPI = XuvaAPI(baseURL: url, authToken: storedAuthToken())
            bootstrap = try await nextAPI.bootstrap()
            api = nextAPI
            connectionState = .connected
            screen = .pair
        }
    }

    public func startPairing() async {
        await run {
            guard let api else { throw XuvaAPIError.invalidURL }
            pairing = try await api.createPairing(deviceName: deviceName(), deviceId: deviceId)
            connectionState = .pairing
            screen = .pair
        }
    }

    /// Select a server discovered via Bonjour and immediately request a
    /// pairing code. All the admin has to do is approve from the web UI.
    public func selectDiscoveredServer(_ server: DiscoveredServer) async {
        serverText = server.baseURL.absoluteString
        await connect()
        guard errorMessage == nil else { return }
        await startPairing()
    }

    public func pollPairingOnce() async {
        guard let api, let id = pairing?.stableID, !id.isEmpty else { return }
        await run(showBusy: false) {
            pairing = try await api.pairingStatus(id: id)
            if pairing?.status?.lowercased() == "approved" {
                markPaired(with: pairing?.auth?.sessionToken)
                await loadHome()
            }
        }
    }

    public func loadHome() async {
        await run {
            guard let api else { throw XuvaAPIError.invalidURL }
            home = try await api.home()
            persistCurrentAuthToken()
            connectionState = .paired
            screen = .home
        }
        if UserDefaults.standard.bool(forKey: "xuva.dev.autoOpenFirstItem") {
            if let first = home?.rows?.flatMap({ $0.items ?? [] }).first {
                try? await Task.sleep(nanoseconds: 600_000_000)
                await open(item: first)
            }
        }
    }

    public func open(item: HomeItem) async {
        let detailId = item.resolvedDetailId
        let detailKind = item.resolvedDetailKind
        print("[XUVA] open item detailId=\(detailId) kind=\(detailKind) title=\(item.title ?? "-") progress=\(item.progress ?? 0)")
        // If this is a Continue Watching tile, capture the resume context so
        // the user's next Play press picks the right media source + position.
        if let progress = item.progress, progress > 0, let msid = item.mediaSourceId, !msid.isEmpty {
            pendingResumeFraction = progress
            pendingResumeMediaSourceId = msid
        } else {
            pendingResumeFraction = nil
            pendingResumeMediaSourceId = nil
        }
        await run {
            guard let api else { throw XuvaAPIError.invalidURL }
            selectedDetail = try await api.detail(kind: detailKind, id: detailId)
            print("[XUVA] detail loaded versions=\(selectedDetail?.versions?.count ?? 0)")
            persistCurrentAuthToken()
            screen = .detail
        }
        // Enrich with cast/writers/studios in the background — the detail screen
        // re-renders once selectedDetail.enrichedMetadata appears.
        Task { [weak self] in
            await self?.enrichSelectedDetail(itemKind: detailKind, itemId: detailId)
        }
    }

    private func enrichSelectedDetail(itemKind: String, itemId: String) async {
        guard let api else { return }
        do {
            // Fetch metadata and similar titles in parallel.
            async let metadataTask = api.metadata(kind: itemKind, id: itemId)
            async let similarTask = api.similar(kind: itemKind, id: itemId)
            let (metadata, similar) = try await (metadataTask, similarTask)
            // Don't overwrite if user navigated away
            guard selectedDetail?.item?.id == itemId else { return }
            selectedDetail?.enrichedMetadata = metadata.best
            selectedDetail?.relatedTitles = similar.items
            print("[XUVA] enriched cast=\(metadata.best?.cast?.count ?? 0) directors=\(metadata.best?.directors?.count ?? 0) studios=\(metadata.best?.studios?.count ?? 0) similar=\(similar.items?.count ?? 0)")
        } catch {
            print("[XUVA] enrich failed: \(error)")
        }
    }

    public func play(version: MediaVersion? = nil, audioTrack: MediaTrack? = nil, subtitleTrack: MediaTrack? = nil) async {
        print("[XUVA] play() called, hasAPI=\(api != nil) pendingResume=\(pendingResumeFraction ?? 0)")
        await run {
            guard let api else { throw XuvaAPIError.invalidURL }
            // Pick the right mediaSource: explicit param > resume's source > first version.
            let mediaSourceId = version?.mediaSourceId
                ?? pendingResumeMediaSourceId
                ?? selectedDetail?.versions?.first?.mediaSourceId
            print("[XUVA] play() mediaSourceId=\(mediaSourceId ?? "<none>")")
            guard let mediaSourceId, !mediaSourceId.isEmpty else { throw XuvaAPIError.missingStreamURL }
            // Compute resume position from progress fraction × that source's duration.
            var positionSeconds = 0
            if let fraction = pendingResumeFraction, fraction > 0,
               let duration = selectedDetail?.versions?.first(where: { $0.mediaSourceId == mediaSourceId })?.source?.durationSeconds, duration > 0 {
                positionSeconds = Int((fraction * duration).rounded())
                print("[XUVA] resume positionSeconds=\(positionSeconds) (\(Int(fraction * 100))% of \(Int(duration))s)")
            }
            var response = try await api.startPlayback(
                mediaSourceId: mediaSourceId,
                positionSeconds: positionSeconds,
                audioTrackIndex: audioTrack?.index,
                subtitleTrackIndex: subtitleTrack?.index,
                subtitleTrackActive: subtitleTrack != nil
            )
            // One-shot — clear so a fresh detail open doesn't carry over.
            pendingResumeFraction = nil
            pendingResumeMediaSourceId = nil
            if positionSeconds > 0 {
                response.clientStartPositionSeconds = positionSeconds
            }
            response.clientSubtitleTrack = subtitleTrack
            let routeType = response.route?.route ?? ""
            print("[XUVA] startPlayback OK routeType=\(routeType) status=\(response.route?.status ?? "<none>") session=\(response.sessionId ?? "<none>") deviceId=\(response.deviceId ?? "<none>")")
            if routeType == "direct" || routeType == "" {
                // Direct play: sign the stream URL so AVPlayer can fetch it without a
                // persistent auth header (token in query string survives redirects).
                if let sessionId = response.sessionId, !sessionId.isEmpty,
                   let deviceId = response.deviceId, !deviceId.isEmpty {
                    let signed = try await api.requestStreamToken(mediaSourceId: mediaSourceId, sessionId: sessionId, deviceId: deviceId)
                    print("[XUVA] streamToken OK signedUrl=\(signed.streamUrl ?? "<none>")")
                    guard let signedUrl = signed.streamUrl, !signedUrl.isEmpty else {
                        throw XuvaAPIError.missingStreamURL
                    }
                    response.route?.url = signedUrl
                } else {
                    print("[XUVA] WARN missing sessionId/deviceId — skipping streamToken")
                }
            } else if routeType == "adaptive" {
                // Adaptive HLS: the manifestUrl is used directly.
                // AVURLAsset passes X-Auth-Token on every sub-request, so no extra token needed.
                guard response.route?.manifestUrl != nil else { throw XuvaAPIError.missingStreamURL }
            } else {
                // Remux / transcode: the server starts a background job and returns the work file
                // URL once complete. Poll until ready (up to 90 s for a typical remux).
                // AVURLAsset's X-Auth-Token header handles auth for /api/work/{id}/file.
                if response.route?.url == nil {
                    let inProgress: Set<String> = ["queued", "running"]
                    for _ in 0..<30 {
                        guard let status = response.route?.status, inProgress.contains(status) else { break }
                        try await Task.sleep(nanoseconds: 3_000_000_000)
                        let next = try await api.startPlayback(
                            mediaSourceId: mediaSourceId,
                            positionSeconds: positionSeconds,
                            audioTrackIndex: audioTrack?.index,
                            subtitleTrackIndex: subtitleTrack?.index,
                            subtitleTrackActive: subtitleTrack != nil
                        )
                        response = next
                        if next.route?.url != nil { break }
                    }
                    guard response.route?.url != nil else { throw XuvaAPIError.missingStreamURL }
                }
                // Work file URL auth is via X-Auth-Token header in AVURLAsset — no stream token.
            }
            playback = response
            persistCurrentAuthToken()
            screen = .player
            print("[XUVA] screen=.player, final url=\(response.route?.url ?? response.route?.manifestUrl ?? "<none>")")
        }
    }

    public func playEpisode(_ episode: EpisodeItem) async {
        guard let mediaSourceId = episode.defaultMediaSourceId, !mediaSourceId.isEmpty else {
            errorMessage = "This episode has no playable source yet."
            return
        }
        let version = episode.versions?.first
        // Thread resume position through to play() the same way Continue
        // Watching tiles do via pendingResumeFraction/pendingResumeMediaSourceId.
        if let seconds = episode.positionSeconds, seconds > 0 {
            // Server sent an absolute position — convert to fraction using the
            // source duration so play() can compute positionSeconds correctly.
            if let duration = version?.source?.durationSeconds, duration > 0 {
                pendingResumeFraction = min(1.0, Double(seconds) / duration)
            } else {
                pendingResumeFraction = nil
            }
        } else if let fraction = episode.progress, fraction > 0 {
            pendingResumeFraction = fraction
        } else {
            pendingResumeFraction = nil
        }
        pendingResumeMediaSourceId = pendingResumeFraction != nil ? mediaSourceId : nil
        await play(version: version)
    }

    public func closePlayer() {
        playback = nil
        screen = .detail
    }

    public func stopPlayback(positionSeconds: Int = 0, completed: Bool = false) async {
        guard let api, let stopURL = playback?.stopUrl else {
            closePlayer()
            return
        }
        do {
            try await api.stop(path: stopURL, positionSeconds: positionSeconds, completed: completed)
            persistCurrentAuthToken()
        } catch {
            errorMessage = error.localizedDescription
        }
        closePlayer()
    }

    public func backToHome() {
        screen = .home
        // Keep selectedDetail cached so re-entering the same title is instant.
    }

    public func setSection(_ section: String) {
        activeSection = section
        heroIndex = 0
    }

    public func clearError() {
        errorMessage = nil
    }

    public func resetConnection() {
        // Best-effort: tell the server to drop the still-pending pairing we
        // created. Fire-and-forget — if the server's offline, the row will
        // auto-expire after 10 minutes anyway.
        if let api,
           let pairingId = pairing?.stableID, !pairingId.isEmpty,
           pairing?.status?.lowercased() != "approved" {
            let dev = deviceId
            Task { try? await api.cancelPairing(id: pairingId, deviceId: dev) }
        }
        bootstrap = nil
        pairing = nil
        home = nil
        selectedDetail = nil
        playback = nil
        api = nil
        errorMessage = nil
        connectionState = .idle
        screen = .connect
        UserDefaults.standard.removeObject(forKey: Self.pairedDeviceKey)
        UserDefaults.standard.removeObject(forKey: Self.pairedServerURLKey)
        UserDefaults.standard.removeObject(forKey: Self.authTokenKey)
    }

    public func reconnectIfPossible() async {
        guard !isBusy, api == nil, !serverText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else { return }
        await connect()
    }

    public func autoConnectIfPossible() async {
        let trimmed = serverText.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !isBusy, api == nil, !trimmed.isEmpty else { return }
        guard !UserDefaults.standard.bool(forKey: Self.pairedDeviceKey) else { return }
        await connect()
        guard errorMessage == nil else { return }
        await startPairing()
    }

    public func resumeSessionIfPossible() async {
        guard !isBusy, api == nil, UserDefaults.standard.bool(forKey: Self.pairedDeviceKey) else { return }
        await run {
            guard let url = URL(string: normalizedServerURL()) else { throw XuvaAPIError.invalidURL }
            let nextAPI = XuvaAPI(baseURL: url, authToken: storedAuthToken())
            bootstrap = try await nextAPI.bootstrap()
            api = nextAPI
            home = try await nextAPI.home()
            persistCurrentAuthToken()
            connectionState = .paired
            screen = .home
        }
        // Only drop back to the connect screen on definitive auth failures
        // (401/403 → connectionState == .needsAuthCredential). A timeout or
        // "server unreachable" error just means the server is temporarily
        // offline — the device is still paired and should stay on the home
        // screen so the user can retry without going through pairing again.
        if connectionState == .needsAuthCredential {
            screen = .connect
        }
        if UserDefaults.standard.bool(forKey: "xuva.dev.autoOpenFirstItem") {
            if let first = home?.rows?.flatMap({ $0.items ?? [] }).first {
                try? await Task.sleep(nanoseconds: 800_000_000)
                await open(item: first)
            }
        }
    }

    private func run(showBusy: Bool = true, _ operation: () async throws -> Void) async {
        if showBusy { isBusy = true }
        errorMessage = nil
        do {
            try await operation()
        } catch {
            print("[XUVA] ERR \(error)")
            errorMessage = error.localizedDescription
            if case XuvaAPIError.badStatus(401, _) = error {
                connectionState = .needsAuthCredential
            } else if case XuvaAPIError.badStatus(403, _) = error {
                connectionState = .needsAuthCredential
            }
        }
        if showBusy { isBusy = false }
    }

    private func normalizedServerURL() -> String {
        let trimmed = serverText.trimmingCharacters(in: .whitespacesAndNewlines)
        if trimmed.hasPrefix("http://") || trimmed.hasPrefix("https://") { return trimmed }
        return "http://\(trimmed)"
    }

    private func deviceName() -> String {
        #if os(tvOS)
        return "Xuva Apple TV"
        #else
        return "Xuva iOS"
        #endif
    }

    /// Claim a QR pair token that was generated in the web admin.
    /// The claim endpoint auto-approves the device and returns an authToken.
    public func claimQRToken(_ token: String) async {
        await run {
            guard let api else { throw XuvaAPIError.invalidURL }
            let response = try await api.claimQRToken(
                token: token,
                deviceName: deviceName(),
                clientProfile: clientProfile(),
                deviceId: deviceId
            )
            markPaired(with: response.authToken)
            await loadHome()
        }
    }

    private func clientProfile() -> String {
        #if os(tvOS)
        return "apple-tv"
        #else
        return "apple-ios"
        #endif
    }

    private func markPaired(with token: String?) {
        if let token = token?.trimmingCharacters(in: .whitespacesAndNewlines), !token.isEmpty {
            api?.authToken = token
            UserDefaults.standard.set(token, forKey: Self.authTokenKey)
        }
        connectionState = .paired
        UserDefaults.standard.set(true, forKey: Self.pairedDeviceKey)
        UserDefaults.standard.set(serverText, forKey: Self.pairedServerURLKey)
    }

    private func storedAuthToken() -> String? {
        let token = UserDefaults.standard.string(forKey: Self.authTokenKey)?.trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
        return token.isEmpty ? nil : token
    }

    private func persistCurrentAuthToken() {
        guard let token = api?.authToken?.trimmingCharacters(in: .whitespacesAndNewlines), !token.isEmpty else { return }
        UserDefaults.standard.set(token, forKey: Self.authTokenKey)
    }
}

public enum XuvaScreen {
    case connect
    case pair
    case home
    case detail
    case player
}

public enum XuvaConnectionState: Equatable {
    case idle
    case connected
    case pairing
    case paired
    case needsAuthCredential
}
