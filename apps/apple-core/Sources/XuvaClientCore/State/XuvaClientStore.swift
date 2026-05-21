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
    }

    public func open(item: HomeItem) async {
        await run {
            guard let api else { throw XuvaAPIError.invalidURL }
            selectedDetail = try await api.detail(kind: item.kind ?? "movie", id: item.id)
            persistCurrentAuthToken()
            screen = .detail
        }
    }

    public func play(version: MediaVersion? = nil) async {
        await run {
            guard let api else { throw XuvaAPIError.invalidURL }
            let mediaSourceId = version?.mediaSourceId ?? selectedDetail?.versions?.first?.mediaSourceId
            guard let mediaSourceId, !mediaSourceId.isEmpty else { throw XuvaAPIError.missingStreamURL }
            playback = try await api.startPlayback(mediaSourceId: mediaSourceId)
            persistCurrentAuthToken()
            screen = .player
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
        selectedDetail = nil
        screen = .home
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

    private func run(showBusy: Bool = true, _ operation: () async throws -> Void) async {
        if showBusy { isBusy = true }
        errorMessage = nil
        do {
            try await operation()
        } catch {
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
