import Foundation
import SwiftUI
import Security
import os
#if os(tvOS)
import TVServices
#endif

@MainActor
public final class XuvaClientStore: ObservableObject {
    private let logger = Logger(subsystem: "com.xuva.client", category: "store")

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
    /// Full library lists loaded when Movies/TV tabs are first activated.
    /// Kept populated so subsequent tab switches are instant; refreshed when
    /// the home data is refreshed.
    @Published public var moviesLibrary: [HomeItem]?
    @Published public var seriesLibrary: [HomeItem]?

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
    private static let keychainAuthTokenAccount = "xuva.apple.authToken"

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
        // showBusy: false — the pairing screen already shows its own inline
        // progress (the "Servers on your network" header spinner + the
        // "Generating pairing code…" line in the pairing card). Letting the
        // big full-screen overlay run here covers those over with a black
        // pill, which felt redundant and noisy.
        await run(showBusy: false) {
            guard let url = URL(string: normalizedServerURL()) else { throw XuvaAPIError.invalidURL }
            let nextAPI = XuvaAPI(baseURL: url, authToken: storedAuthToken())
            bootstrap = try await nextAPI.bootstrap()
            api = nextAPI
            connectionState = .connected
            screen = .pair
        }
    }

    public func startPairing() async {
        // Same reasoning as connect() — the pairing card has its own progress
        // copy; suppress the full-screen overlay here.
        await run(showBusy: false) {
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
        if let snapshot = home, let currentAPI = api {
            scheduleTopShelfRefresh(home: snapshot, api: currentAPI)
        }
        // Silently refresh any library pages already cached so Movies/TV tabs
        // stay in sync after returning from player or a foreground transition.
        Task { [weak self] in
            guard let self else { return }
            if self.moviesLibrary != nil { await self.loadMoviesLibrary() }
            if self.seriesLibrary != nil { await self.loadSeriesLibrary() }
        }
        if UserDefaults.standard.bool(forKey: "xuva.dev.autoOpenFirstItem") {
            if let first = home?.rows?.flatMap({ $0.items ?? [] }).first {
                try? await Task.sleep(nanoseconds: 600_000_000)
                await open(item: first)
            }
        }
    }

    public func loadMoviesLibrary() async {
        await run(showBusy: false) {
            guard let api else { return }
            let resp = try await api.libraryMovies()
            moviesLibrary = (resp.movies ?? []).map { $0.toHomeItem() }
        }
    }

    public func loadSeriesLibrary() async {
        await run(showBusy: false) {
            guard let api else { return }
            let resp = try await api.librarySeries()
            seriesLibrary = (resp.series ?? []).map { $0.toHomeItem() }
        }
    }

    public func open(item: HomeItem) async {
        let detailId = item.resolvedDetailId
        let detailKind = item.resolvedDetailKind
        logger.debug("open item detailId=\(detailId) kind=\(detailKind) title=\(item.title ?? "-") progress=\(item.progress ?? 0)")
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
            logger.debug("detail loaded versions=\(self.selectedDetail?.versions?.count ?? 0)")
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
        // Launch both requests in parallel but treat similar as a soft failure:
        // if the endpoint is missing or returns an error (e.g. old server binary,
        // title has no genres), cast/metadata must still display.
        async let metadataTask = api.metadata(kind: itemKind, id: itemId)
        async let similarTask  = api.similar(kind: itemKind, id: itemId)
        do {
            let metadata = try await metadataTask
            // Don't overwrite if user navigated away
            guard selectedDetail?.item?.id == itemId else { return }
            selectedDetail?.enrichedMetadata = metadata.best
            // Similar failure is non-fatal — apply if available, skip if not.
            let similar = try? await similarTask
            selectedDetail?.relatedTitles = similar?.items
            logger.debug("enriched cast=\(metadata.best?.cast?.count ?? 0) directors=\(metadata.best?.directors?.count ?? 0) studios=\(metadata.best?.studios?.count ?? 0) similar=\(similar?.items?.count ?? 0)")
        } catch {
            logger.debug("enrich failed: \(error)")
        }
    }

    public func play(version: MediaVersion? = nil, audioTrack: MediaTrack? = nil, subtitleTrack: MediaTrack? = nil) async {
        logger.debug("play() called, hasAPI=\(self.api != nil) pendingResume=\(self.pendingResumeFraction ?? 0)")
        await run {
            guard let api else { throw XuvaAPIError.invalidURL }
            // Pick the right mediaSource: explicit param > resume's source > first version.
            let mediaSourceId = version?.mediaSourceId
                ?? pendingResumeMediaSourceId
                ?? selectedDetail?.versions?.first?.mediaSourceId
            logger.debug("play() mediaSourceId=\(mediaSourceId ?? "<none>")")
            guard let mediaSourceId, !mediaSourceId.isEmpty else { throw XuvaAPIError.missingStreamURL }
            // Compute resume position from progress fraction × that source's duration.
            var positionSeconds = 0
            if let fraction = pendingResumeFraction, fraction > 0,
               let duration = selectedDetail?.versions?.first(where: { $0.mediaSourceId == mediaSourceId })?.source?.durationSeconds, duration > 0 {
                positionSeconds = Int((fraction * duration).rounded())
                logger.debug("resume positionSeconds=\(positionSeconds) (\(Int(fraction * 100))% of \(Int(duration))s)")
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
            // Hard-block: file has not been probed yet. Surface a clear message
            // instead of falling into the poll loop for 90 seconds.
            if response.route?.status == "deferred" || response.route?.route == "deferred" {
                throw XuvaAPIError.fileNotProbed
            }
            let routeType = response.route?.route ?? ""
            logger.debug("startPlayback OK routeType=\(routeType) status=\(response.route?.status ?? "<none>") session=\(response.sessionId ?? "<none>")")
            if routeType == "direct" || routeType == "" {
                // Direct play: sign the stream URL so AVPlayer can fetch it without a
                // persistent auth header (token in query string survives redirects).
                if let sessionId = response.sessionId, !sessionId.isEmpty,
                   let deviceId = response.deviceId, !deviceId.isEmpty {
                    let signed = try await api.requestStreamToken(mediaSourceId: mediaSourceId, sessionId: sessionId, deviceId: deviceId)
                    logger.debug("streamToken OK")
                    guard let signedUrl = signed.streamUrl, !signedUrl.isEmpty else {
                        throw XuvaAPIError.missingStreamURL
                    }
                    response.route?.url = signedUrl
                } else {
                    logger.warning("missing sessionId/deviceId — skipping streamToken")
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
            logger.debug("screen=.player")
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
        // Refresh home silently so the Continue Watching row updates immediately
        // when the user returns to the home screen after stopping playback.
        Task { [weak self] in await self?.loadHome() }
    }

    public func backToHome() {
        screen = .home
        // Keep selectedDetail cached so re-entering the same title is instant.
    }

    /// Permanently deletes the media source file from the server, then navigates
    /// back to the home screen so the stale detail view is not left showing.
    public func deleteMediaSource(id: String) async {
        await run {
            guard let api else { throw XuvaAPIError.invalidURL }
            try await api.deleteMediaSource(id: id)
            // Navigate away; the item no longer exists.
            screen = .home
            selectedDetail = nil
        }
    }

    public func setSection(_ section: String) {
        activeSection = section
        heroIndex = 0
        Task {
            switch section {
            case "Movies": await loadMoviesLibrary()
            case "TV":     await loadSeriesLibrary()
            default: break
            }
        }
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
        deleteAuthTokenFromKeychain()
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
            logger.error("ERR \(error)")
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
        let scheme = isLocalAddress(trimmed) ? "http" : "https"
        return "\(scheme)://\(trimmed)"
    }

    private func isLocalAddress(_ address: String) -> Bool {
        let host = address.components(separatedBy: ":").first ?? address
        if host == "localhost" || host.hasSuffix(".local") { return true }
        if host.hasPrefix("10.") || host.hasPrefix("192.168.") { return true }
        for i in 16...31 where host.hasPrefix("172.\(i).") { return true }
        return false
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
            saveAuthTokenToKeychain(token)
        }
        connectionState = .paired
        UserDefaults.standard.set(true, forKey: Self.pairedDeviceKey)
        UserDefaults.standard.set(serverText, forKey: Self.pairedServerURLKey)
    }

    private func storedAuthToken() -> String? {
        loadAuthTokenFromKeychain()
    }

    private func persistCurrentAuthToken() {
        guard let token = api?.authToken?.trimmingCharacters(in: .whitespacesAndNewlines), !token.isEmpty else { return }
        saveAuthTokenToKeychain(token)
    }

    private func saveAuthTokenToKeychain(_ token: String) {
        guard let data = token.data(using: .utf8) else { return }
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrAccount as String: Self.keychainAuthTokenAccount
        ]
        let updateStatus = SecItemUpdate(query as CFDictionary, [kSecValueData as String: data] as CFDictionary)
        if updateStatus == errSecItemNotFound {
            var addQuery = query
            addQuery[kSecValueData as String] = data
            // Accessible after first unlock so background refresh / launch can read
            // the token, but never while the device is locked and never synced to
            // iCloud or migrated to a new device.
            addQuery[kSecAttrAccessible as String] = kSecAttrAccessibleAfterFirstUnlockThisDeviceOnly
            let addStatus = SecItemAdd(addQuery as CFDictionary, nil)
            if addStatus != errSecSuccess {
                logger.error("keychain auth-token add failed: OSStatus \(addStatus)")
            }
        } else if updateStatus != errSecSuccess {
            logger.error("keychain auth-token update failed: OSStatus \(updateStatus)")
        }
    }

    private func loadAuthTokenFromKeychain() -> String? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrAccount as String: Self.keychainAuthTokenAccount,
            kSecReturnData as String: true,
            kSecMatchLimit as String: kSecMatchLimitOne
        ]
        var result: AnyObject?
        guard SecItemCopyMatching(query as CFDictionary, &result) == errSecSuccess,
              let data = result as? Data,
              let token = String(data: data, encoding: .utf8),
              !token.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else { return nil }
        return token
    }

    private func deleteAuthTokenFromKeychain() {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrAccount as String: Self.keychainAuthTokenAccount
        ]
        let status = SecItemDelete(query as CFDictionary)
        if status != errSecSuccess && status != errSecItemNotFound {
            logger.error("keychain auth-token delete failed: OSStatus \(status)")
        }
    }

    // MARK: - Deep links

    public func openDeepLink(kind: String, id: String) async {
        guard connectionState == .paired, api != nil else { return }
        await run {
            guard let api else { throw XuvaAPIError.invalidURL }
            selectedDetail = try await api.detail(kind: kind, id: id)
            persistCurrentAuthToken()
            screen = .detail
        }
    }

    // MARK: - Top Shelf

    private static let topShelfAppGroup = "group.com.xuva.tvos"
    private static let topShelfPayloadKey = "xuva.topShelf.payload"

    func scheduleTopShelfRefresh(home: ClientHomeResponse, api: XuvaAPI) {
        Task.detached { [weak self] in
            await self?.refreshTopShelf(home: home, api: api)
        }
    }

    private func refreshTopShelf(home: ClientHomeResponse, api: XuvaAPI) async {
        guard let container = FileManager.default.containerURL(
            forSecurityApplicationGroupIdentifier: Self.topShelfAppGroup) else { return }
        let imagesDir = container.appendingPathComponent("topshelf-images")
        try? FileManager.default.createDirectory(at: imagesDir, withIntermediateDirectories: true)

        let (items, sectionTitle) = topShelfCandidates(from: home)
        guard !items.isEmpty else { return }

        var entries: [TopShelfEntryData] = []
        for item in items.prefix(10) {
            let filename = await downloadTopShelfImage(item: item, api: api, into: imagesDir)
            entries.append(TopShelfEntryData(
                id: item.id,
                title: item.title ?? "",
                detailKind: item.resolvedDetailKind,
                detailId: item.resolvedDetailId,
                imageFilename: filename,
                progress: item.progress
            ))
        }

        let payload = TopShelfPayloadData(items: entries, sectionTitle: sectionTitle)
        if let data = try? JSONEncoder().encode(payload) {
            UserDefaults(suiteName: Self.topShelfAppGroup)?.set(data, forKey: Self.topShelfPayloadKey)
        }
        #if os(tvOS)
        TVTopShelfContentProvider.topShelfContentDidChange()
        #endif
    }

    private func topShelfCandidates(from home: ClientHomeResponse) -> ([HomeItem], String) {
        let rows = home.rows ?? []
        let cwRow = rows.first {
            $0.id.lowercased() == "continue" || ($0.kind ?? "").lowercased().contains("continue")
        }
        if let items = cwRow?.items, !items.isEmpty {
            return (items, "Continue Watching")
        }
        let newItems = rows.flatMap { $0.items ?? [] }.filter { ($0.progress ?? 0) == 0 }
        if !newItems.isEmpty { return (Array(newItems.prefix(10)), "Recently Added") }
        return (Array(rows.flatMap { $0.items ?? [] }.prefix(10)), "From Your Library")
    }

    private func downloadTopShelfImage(item: HomeItem, api: XuvaAPI, into dir: URL) async -> String? {
        let urlStr = item.backdropUrl ?? item.posterUrl ?? item.imageUrl
        guard let urlStr, let url = api.resolvedURL(urlStr) else { return nil }
        let filename = "\(item.id).jpg"
        let dest = dir.appendingPathComponent(filename)
        if FileManager.default.fileExists(atPath: dest.path) { return filename }
        var req = URLRequest(url: url)
        if let token = api.authToken { req.setValue(token, forHTTPHeaderField: "X-Auth-Token") }
        guard let (data, _) = try? await URLSession.shared.data(for: req) else { return nil }
        try? data.write(to: dest)
        return FileManager.default.fileExists(atPath: dest.path) ? filename : nil
    }
}

private struct TopShelfEntryData: Codable {
    let id: String; let title: String; let detailKind: String
    let detailId: String; let imageFilename: String?; let progress: Double?
}
private struct TopShelfPayloadData: Codable {
    let items: [TopShelfEntryData]; let sectionTitle: String
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
