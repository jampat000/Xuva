import Foundation

public enum XuvaAPIError: LocalizedError {
    case invalidURL
    case badStatus(Int, String)
    case missingStreamURL
    case fileNotProbed

    public var errorDescription: String? {
        switch self {
        case .invalidURL:
            return "The Xuva address is not valid."
        case let .badStatus(code, body):
            return body.isEmpty ? "Connection returned HTTP \(code)." : "Connection returned HTTP \(code): \(body)"
        case .missingStreamURL:
            return "Xuva could not prepare a playable stream URL."
        case .fileNotProbed:
            return "This file has not been analysed yet. Open Settings → Activity on your Xuva server to run the Probe job, then try again."
        }
    }
}

public final class XuvaAPI: @unchecked Sendable {
    public let baseURL: URL
    public var authToken: String?
    private let session: URLSession
    private let decoder = JSONDecoder()
    private let encoder = JSONEncoder()

    public init(baseURL: URL, authToken: String? = nil, session: URLSession = .shared) {
        self.baseURL = baseURL
        self.authToken = authToken
        self.session = session
    }

    /// Percent-encode an ID for safe inclusion in a URL path. Server IDs
    /// can be derived from file paths or arbitrary identifiers and may
    /// contain `#`, `?`, `%`, `+`, or space — all of which would otherwise
    /// break the URL or be mis-parsed by the server.
    private func enc(_ id: String) -> String {
        id.addingPercentEncoding(withAllowedCharacters: .urlPathAllowed) ?? id
    }

    public func bootstrap() async throws -> BootstrapResponse {
        try await send("GET", path: "/api/client/bootstrap?clientProfile=\(XuvaClientProfile.current)")
    }

    public func createPairing(deviceName: String, deviceId: String) async throws -> PairingResponse {
        let request = PairingCreateRequest(deviceName: deviceName, clientProfile: XuvaClientProfile.current, deviceId: deviceId)
        return try await send("POST", path: "/api/pairing/requests", body: request)
    }

    public func pairingStatus(id: String) async throws -> PairingResponse {
        try await send("GET", path: "/api/pairing/requests/\(enc(id))")
    }

    public func cancelPairing(id: String, deviceId: String) async throws {
        struct Body: Codable { let deviceId: String }
        let _: EmptyResponse = try await send("DELETE", path: "/api/pairing/requests/\(enc(id))", body: Body(deviceId: deviceId))
    }

    public func home() async throws -> ClientHomeResponse {
        try await send("GET", path: "/api/client/home?clientProfile=\(XuvaClientProfile.current)")
    }

    public func detail(kind: String, id: String) async throws -> DetailResponse {
        let normalized = kind.lowercased().contains("series") || kind.lowercased().contains("show") ? "series" : "movies"
        return try await send("GET", path: "/api/client/\(normalized)/\(enc(id))")
    }

    public func metadata(kind: String, id: String) async throws -> MetadataRecordsResponse {
        let normalized = kind.lowercased().contains("series") || kind.lowercased().contains("show") ? "series" : "movie"
        return try await send("GET", path: "/api/metadata/\(normalized)/\(enc(id))")
    }

    public func similar(kind: String, id: String) async throws -> SimilarResponse {
        let segment = kind.lowercased().contains("series") || kind.lowercased().contains("show") ? "series" : "movies"
        return try await send("GET", path: "/api/client/\(segment)/\(enc(id))/similar")
    }


    public func libraryMovies() async throws -> LibraryMoviesResponse {
        try await send("GET", path: "/api/movies")
    }

    public func librarySeries() async throws -> LibrarySeriesResponse {
        try await send("GET", path: "/api/series")
    }

    /// Permanently deletes the media source file from the server (admin only).
    public func deleteMediaSource(id: String) async throws {
        let _: EmptyResponse = try await send("DELETE", path: "/api/media-sources/\(enc(id))")
    }

    public func startPlayback(
        mediaSourceId: String,
        positionSeconds: Int = 0,
        audioTrackIndex: Int? = nil,
        subtitleTrackIndex: Int? = nil,
        subtitleTrackActive: Bool? = nil
    ) async throws -> PlaybackStartResponse {
        let body = PlaybackStartRequest(
            mediaSourceId: mediaSourceId,
            clientProfile: XuvaClientProfile.current,
            positionSeconds: positionSeconds,
            audioTrackIndex: audioTrackIndex,
            subtitleTrackIndex: subtitleTrackIndex,
            subtitleTrackActive: subtitleTrackActive,
            supportsAdaptive: true,
            preferAdaptive: true
        )
        return try await send("POST", path: "/api/client/playback/start", body: body)
    }

    public func requestStreamToken(mediaSourceId: String, sessionId: String, deviceId: String) async throws -> StreamTokenResponse {
        let body = StreamTokenRequest(sessionId: sessionId, deviceId: deviceId)
        return try await send("POST", path: "/api/media-sources/\(enc(mediaSourceId))/stream-token", body: body)
    }

    public func heartbeat(path: String, positionSeconds: Int, isPaused: Bool) async throws {
        let body = PlaybackHeartbeat(positionSeconds: positionSeconds, isPaused: isPaused, completed: false)
        let _: EmptyResponse = try await send("PATCH", path: path, body: body)
    }

    public func stop(path: String, positionSeconds: Int, completed: Bool) async throws {
        let body = PlaybackHeartbeat(positionSeconds: positionSeconds, isPaused: true, completed: completed)
        let _: EmptyResponse = try await send("POST", path: path, body: body)
    }

    // MARK: – QR pair token

    public func claimQRToken(token: String, deviceName: String, clientProfile: String, deviceId: String) async throws -> QRClaimResponse {
        let body = QRClaimBody(deviceName: deviceName, clientProfile: clientProfile, deviceId: deviceId)
        return try await send("POST", path: "/api/pairing/qr/\(enc(token))/claim", body: body)
    }

    // MARK: – Watchlist

    public func watchlistList() async throws -> WatchlistListResponse {
        try await send("GET", path: "/api/client/watchlist")
    }

    public func watchlistAdd(_ req: WatchlistAddRequest) async throws -> WatchlistServerItem {
        try await send("POST", path: "/api/client/watchlist", body: req)
    }

    public func watchlistRemove(mediaId: String, kind: String) async throws {
        guard var comps = URLComponents(string: "/api/client/watchlist/\(mediaId)") else { return }
        comps.queryItems = [URLQueryItem(name: "kind", value: kind)]
        let path = comps.url?.absoluteString ?? "/api/client/watchlist/\(mediaId)?kind=\(kind)"
        let _: EmptyResponse = try await send("DELETE", path: path)
    }

    // MARK: – Player metadata (tracks, chapters, state)

    /// Audio + subtitle tracks for a media source. Used to populate the
    /// in-player track menus (matches the web player's track switcher).
    public func mediaSourceTracks(mediaSourceId: String) async throws -> MediaSourceTracksResponse {
        try await send("GET", path: "/api/media-sources/\(enc(mediaSourceId))/tracks")
    }

    /// Detected intro / credits markers. Drives the Skip Intro button and the
    /// Credits marker overlay. Returns an empty response when none are stored.
    public func chapters(mediaSourceId: String) async throws -> ChaptersResponse {
        try await send("GET", path: "/api/media-sources/\(enc(mediaSourceId))/chapters")
    }

    /// Persist final playback progress / watched state. Called when playback
    /// ends so Continue Watching reflects the right position.
    public func setPlaybackState(mediaSourceId: String, update: PlaybackStateUpdate) async throws {
        let _: EmptyResponse = try await send("PUT", path: "/api/playback/state/\(enc(mediaSourceId))", body: update)
    }

    public func resolvedURL(_ pathOrURL: String) -> URL? {
        if let url = URL(string: pathOrURL), url.scheme != nil { return url }
        guard pathOrURL.hasPrefix("/") else { return nil }
        return URL(string: pathOrURL, relativeTo: baseURL)?.absoluteURL
    }

    private func send<Response: Decodable>(_ method: String, path: String) async throws -> Response {
        try await send(method, path: path, body: Optional<EmptyResponse>.none)
    }

    private func send<RequestBody: Encodable, Response: Decodable>(_ method: String, path: String, body: RequestBody?) async throws -> Response {
        guard let url = resolvedURL(path) else { throw XuvaAPIError.invalidURL }
        var request = URLRequest(url: url)
        request.httpMethod = method
        request.setValue("application/json", forHTTPHeaderField: "Accept")
        request.setValue(XuvaClientProfile.current, forHTTPHeaderField: "X-Xuva-Client-Profile")
        if let authToken, !authToken.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            request.setValue(authToken, forHTTPHeaderField: "X-Auth-Token")
        }
        if let body {
            request.httpBody = try encoder.encode(body)
            request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        }

        let (data, response) = try await session.data(for: request)
        let httpResponse = response as? HTTPURLResponse
        if let rotatedToken = httpResponse?.value(forHTTPHeaderField: "X-Auth-Token"), !rotatedToken.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            authToken = rotatedToken
        }
        let statusCode = httpResponse?.statusCode ?? 0
        guard (200..<300).contains(statusCode) else {
            throw XuvaAPIError.badStatus(statusCode, String(data: data, encoding: .utf8) ?? "")
        }
        if Response.self == EmptyResponse.self {
            return EmptyResponse() as! Response
        }
        return try decoder.decode(Response.self, from: data)
    }
}

public struct EmptyResponse: Codable {
    public init() {}
}

public struct QRClaimBody: Codable {
    public let deviceName: String
    public let clientProfile: String
    public let deviceId: String
}

public struct QRClaimResponse: Codable {
    public let deviceId: String
    public let authToken: String?
    public let expiresAt: String?
}

public struct WatchlistServerItem: Codable {
    public let userId: String
    public let mediaId: String
    public let kind: String
    public let title: String
    public let year: Int?
    public let posterUrl: String?
    public let backdropUrl: String?
    public let genres: [String]?
    public let addedAt: String
}

public struct WatchlistListResponse: Codable {
    public let items: [WatchlistServerItem]
}

public struct WatchlistAddRequest: Codable {
    public let mediaId: String
    public let kind: String
    public let title: String
    public let year: Int?
    public let posterUrl: String?
    public let backdropUrl: String?
    public let genres: [String]?

    public init(mediaId: String, kind: String, title: String, year: Int?, posterUrl: String?, backdropUrl: String?, genres: [String]?) {
        self.mediaId = mediaId
        self.kind = kind
        self.title = title
        self.year = year
        self.posterUrl = posterUrl
        self.backdropUrl = backdropUrl
        self.genres = genres
    }
}
