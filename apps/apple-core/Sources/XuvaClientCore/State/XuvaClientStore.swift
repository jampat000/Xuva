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
        print("[XUVA] open item id=\(item.id) kind=\(item.kind ?? "-") title=\(item.title ?? "-")")
        await run {
            guard let api else { throw XuvaAPIError.invalidURL }
            selectedDetail = try await api.detail(kind: item.kind ?? "movie", id: item.id)
            print("[XUVA] detail loaded versions=\(selectedDetail?.versions?.count ?? 0)")
            persistCurrentAuthToken()
            screen = .detail
        }
        // Enrich with cast/writers/studios in the background — the detail screen
        // re-renders once selectedDetail.enrichedMetadata appears.
        Task { [weak self] in
            await self?.enrichSelectedDetail(itemKind: item.kind ?? "movie", itemId: item.id)
        }
    }

    private func enrichSelectedDetail(itemKind: String, itemId: String) async {
        guard let api else { return }
        do {
            let metadata = try await api.metadata(kind: itemKind, id: itemId)
            // Don't overwrite if user navigated away
            guard selectedDetail?.item?.id == itemId else { return }
            selectedDetail?.enrichedMetadata = metadata.best
            print("[XUVA] enriched cast=\(metadata.best?.cast?.count ?? 0) directors=\(metadata.best?.directors?.count ?? 0) studios=\(metadata.best?.studios?.count ?? 0)")
        } catch {
            print("[XUVA] enrich failed: \(error)")
        }
    }

    public func play(version: MediaVersion? = nil, audioTrack: MediaTrack? = nil, subtitleTrack: MediaTrack? = nil) async {
        print("[XUVA] play() called, hasAPI=\(api != nil)")
        await run {
            guard let api else { throw XuvaAPIError.invalidURL }
            let mediaSourceId = version?.mediaSourceId ?? selectedDetail?.versions?.first?.mediaSourceId
            print("[XUVA] play() mediaSourceId=\(mediaSourceId ?? "<none>")")
            guard let mediaSourceId, !mediaSourceId.isEmpty else { throw XuvaAPIError.missingStreamURL }
            var response = try await api.startPlayback(
                mediaSourceId: mediaSourceId,
                audioTrackIndex: audioTrack?.index,
                subtitleTrackIndex: subtitleTrack?.index,
                subtitleTrackActive: subtitleTrack != nil
            )
            print("[XUVA] startPlayback OK session=\(response.sessionId ?? "<none>") deviceId=\(response.deviceId ?? "<none>") routeUrl=\(response.route?.url ?? "<none>")")
            if let sessionId = response.sessionId, !sessionId.isEmpty,
               let deviceId = response.deviceId, !deviceId.isEmpty {
                let signed = try await api.requestStreamToken(mediaSourceId: mediaSourceId, sessionId: sessionId, deviceId: deviceId)
                print("[XUVA] streamToken OK signedUrl=\(signed.streamUrl ?? "<none>")")
                if let signedUrl = signed.streamUrl, !signedUrl.isEmpty {
                    response.route?.url = signedUrl
                }
            } else {
                print("[XUVA] WARN missing sessionId/deviceId — skipping streamToken")
            }
            playback = response
            persistCurrentAuthToken()
            screen = .player
            print("[XUVA] screen=.player, final url=\(response.route?.url ?? "<none>")")
        }
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
