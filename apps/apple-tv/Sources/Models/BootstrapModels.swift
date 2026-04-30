import Foundation

struct ClientBootstrap: Decodable {
    let server: ServerIdentity
    let auth: AuthInfo
    let client: ClientInfo
    let features: FeatureFlags
    let endpoints: [String: String]
}

struct ServerIdentity: Decodable {
    let product: String
    let name: String
    let baseUrl: String
    let httpAddr: String
    let lanAddresses: [String]
    let startedAt: String
}

struct AuthInfo: Decodable {
    let required: Bool
    let methods: [String]
}

struct ClientInfo: Decodable {
    let requestedProfile: String
    let profile: DeviceProfile
}

struct DeviceProfile: Decodable, Identifiable {
    let id: String
    let name: String
    let containers: [String]
    let videoCodecs: [String]
    let audioCodecs: [String]
    let subtitleCodecs: [String]
    let supportsHdr: Bool
    let supportsToneMapping: Bool
    let supportsHls: Bool
}

struct FeatureFlags: Decodable {
    let directPlay: Bool
    let hlsAdaptive: Bool
    let resume: Bool
    let watchedState: Bool
    let trackSelection: Bool
    let remoteDiagnostics: Bool
    let vendorRelay: Bool
}

struct MediaPoster: Identifiable, Hashable {
    let id: String
    let title: String
    let subtitle: String
    let route: String
    let imageName: String?

    static let samples: [MediaPoster] = [
        MediaPoster(id: "sample-1", title: "Local Feature", subtitle: "4K HDR - Direct Play", route: "Direct Play", imageName: nil),
        MediaPoster(id: "sample-2", title: "Cinema Archive", subtitle: "1080p - Adaptive Stream", route: "Adaptive Stream", imageName: nil),
        MediaPoster(id: "sample-3", title: "Family Library", subtitle: "HD - Ready", route: "Ready", imageName: nil),
    ]
}

struct PairingRequestPayload: Encodable {
    let deviceName: String
    let clientProfile: String
}

struct PairingRequest: Decodable, Identifiable {
    let id: String
    let code: String?
    let deviceName: String
    let clientProfile: String
    let deviceId: String?
    let status: String
    let expiresAt: String
    let createdAt: String
    let updatedAt: String

    var isPending: Bool { status == "pending" }
    var isApproved: Bool { status == "approved" && !(deviceId ?? "").isEmpty }
    var isClosed: Bool { status == "approved" || status == "denied" || status == "expired" }
}

struct TVHomeResponse: Decodable {
    let profile: String
    let hero: TVHomeItem
    let rows: [TVHomeRow]
    let actions: [String: String]
}

struct TVHomeRow: Decodable, Identifiable {
    let id: String
    let title: String
    let items: [TVHomeItem]
}

struct TVHomeItem: Decodable, Identifiable, Hashable {
    let id: String
    let kind: String
    let title: String
    let subtitle: String
    let posterUrl: String?
    let backdropUrl: String?
    let mediaSourceId: String?
    let route: String

    func posterModel() -> MediaPoster {
        MediaPoster(id: id, title: title, subtitle: subtitle, route: route, imageName: nil)
    }
}
