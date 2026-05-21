import Foundation

public enum XuvaAPIError: LocalizedError {
    case invalidURL
    case badStatus(Int, String)
    case missingStreamURL

    public var errorDescription: String? {
        switch self {
        case .invalidURL:
            return "The Xuva server URL is not valid."
        case let .badStatus(code, body):
            return body.isEmpty ? "Server returned HTTP \(code)." : "Server returned HTTP \(code): \(body)"
        case .missingStreamURL:
            return "The server did not return a playable stream URL."
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

    public func bootstrap() async throws -> BootstrapResponse {
        try await send("GET", path: "/api/client/bootstrap?clientProfile=\(XuvaClientProfile.current)")
    }

    public func createPairing(deviceName: String, deviceId: String) async throws -> PairingResponse {
        let request = PairingCreateRequest(deviceName: deviceName, clientProfile: XuvaClientProfile.current, deviceId: deviceId)
        return try await send("POST", path: "/api/pairing/requests", body: request)
    }

    public func pairingStatus(id: String) async throws -> PairingResponse {
        try await send("GET", path: "/api/pairing/requests/\(id)")
    }

    public func home() async throws -> ClientHomeResponse {
        try await send("GET", path: "/api/client/home?clientProfile=\(XuvaClientProfile.current)")
    }

    public func detail(kind: String, id: String) async throws -> DetailResponse {
        let normalized = kind.lowercased().contains("series") || kind.lowercased().contains("show") ? "series" : "movies"
        return try await send("GET", path: "/api/client/\(normalized)/\(id)")
    }

    public func startPlayback(mediaSourceId: String, positionSeconds: Int = 0) async throws -> PlaybackStartResponse {
        let body = PlaybackStartRequest(
            mediaSourceId: mediaSourceId,
            clientProfile: XuvaClientProfile.current,
            positionSeconds: positionSeconds,
            audioTrackIndex: nil,
            subtitleTrackIndex: nil,
            subtitleTrackActive: nil,
            supportsAdaptive: true
        )
        return try await send("POST", path: "/api/client/playback/start", body: body)
    }

    public func heartbeat(path: String, positionSeconds: Int, isPaused: Bool) async throws {
        let body = PlaybackHeartbeat(positionSeconds: positionSeconds, isPaused: isPaused, completed: false)
        let _: EmptyResponse = try await send("PATCH", path: path, body: body)
    }

    public func stop(path: String, positionSeconds: Int, completed: Bool) async throws {
        let body = PlaybackHeartbeat(positionSeconds: positionSeconds, isPaused: true, completed: completed)
        let _: EmptyResponse = try await send("POST", path: path, body: body)
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
