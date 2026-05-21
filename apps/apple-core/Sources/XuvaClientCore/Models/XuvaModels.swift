import Foundation

public enum XuvaClientProfile {
    public static var current: String {
        #if os(tvOS)
        return "apple-tv"
        #else
        return "ios"
        #endif
    }
}

public struct BootstrapResponse: Codable, Equatable {
    public var server: ServerIdentity?
    public var auth: AuthInfo?
    public var features: FeatureFlags?
    public var endpoints: EndpointMap?
}

public struct ServerIdentity: Codable, Equatable {
    public var name: String?
    public var baseUrl: String?
    public var httpAddr: String?
}

public struct AuthInfo: Codable, Equatable {
    public var required: Bool?
    public var bootstrapAllowed: Bool?
}

public struct FeatureFlags: Codable, Equatable {
    public var directPlay: Bool?
    public var hlsAdaptive: Bool?
    public var resume: Bool?
    public var watchedState: Bool?
    public var vendorRelay: Bool?
}

public struct EndpointMap: Codable, Equatable {
    public var pairingCreate: String?
    public var pairingStatus: String?
    public var clientHome: String?
    public var playbackDecision: String?
    public var playbackRoute: String?
}

public struct PairingCreateRequest: Codable {
    public var deviceName: String
    public var clientProfile: String
    public var deviceId: String
}

public struct PairingResponse: Codable, Equatable {
    public var id: String?
    public var requestId: String?
    public var code: String?
    public var status: String?
    public var deviceId: String?
    public var auth: PairingAuthGrant?
    public var expiresAt: String?
    public var stableID: String { id ?? requestId ?? "" }
}

public struct PairingAuthGrant: Codable, Equatable {
    public var method: String?
    public var sessionToken: String?
    public var expiresAt: String?
}

public struct ClientHomeResponse: Codable, Equatable {
    public var hero: HomeItem?
    public var heroes: [HomeItem]?
    public var rows: [HomeRow]?
}

public struct HomeRow: Codable, Identifiable, Equatable {
    public var id: String
    public var title: String?
    public var subtitle: String?
    public var kind: String?
    public var items: [HomeItem]?
}

public struct HomeItem: Codable, Identifiable, Equatable {
    public var id: String
    public var kind: String?
    public var title: String?
    public var subtitle: String?
    public var year: Int?
    public var rating: Double?
    public var runtime: String?
    public var progress: Double?
    public var posterUrl: String?
    public var backdropUrl: String?
    public var imageUrl: String?
    public var logoUrl: String?
    public var mediaSourceId: String?
    public var routeLabel: String?
    public var genres: [String]?
    public var overview: String?
}

public struct DetailResponse: Codable, Equatable {
    public var id: String?
    public var kind: String?
    public var title: String?
    public var subtitle: String?
    public var overview: String?
    public var tagline: String?
    public var year: Int?
    public var runtime: String?
    public var runtimeMinutes: Int?
    public var rating: Double?
    public var contentRating: String?
    public var posterUrl: String?
    public var backdropUrl: String?
    public var logoUrl: String?
    public var genres: [String]?
    public var versions: [MediaVersion]?
    public var audioTracks: [MediaTrack]?
    public var subtitleTracks: [MediaTrack]?
    public var playbackDecision: PlaybackDecision?
    public var metadata: MetadataEnvelope?
}

public struct MetadataEnvelope: Codable, Equatable {
    public var title: String?
    public var overview: String?
    public var tagline: String?
    public var posterUrl: String?
    public var backdropUrl: String?
    public var logoUrl: String?
    public var genres: [String]?
    public var year: Int?
    public var voteAverage: Double?
    public var runtime: String?
    public var runtimeMinutes: Int?
    public var contentRating: String?
    public var videoKey: String?
    public var trailerPath: String?
    public var cast: [MetadataCredit]?
    public var directors: [String]?
    public var writers: [String]?
    public var studios: [String]?
    public var productionCompanies: [String]?
    public var networks: [String]?
    public var collection: MetadataCollection?
}

public struct MetadataCredit: Codable, Identifiable, Equatable {
    public var id: String?
    public var name: String?
    public var character: String?
    public var profileUrl: String?
    public var stableID: String { id ?? name ?? UUID().uuidString }
}

public struct MetadataCollection: Codable, Equatable {
    public var id: String?
    public var name: String?
    public var posterUrl: String?
    public var backdropUrl: String?
    public var logoUrl: String?
}

public struct MediaVersion: Codable, Identifiable, Equatable {
    public var id: String?
    public var mediaSourceId: String?
    public var name: String?
    public var qualityLabel: String?
    public var resolution: String?
    public var videoCodec: String?
    public var audioSummary: String?
    public var bitrate: Int?
    public var sizeBytes: Int64?
    public var decision: PlaybackDecision?
    public var stableID: String { mediaSourceId ?? id ?? name ?? UUID().uuidString }
}

public struct MediaTrack: Codable, Identifiable, Equatable {
    public var id: String?
    public var index: Int?
    public var kind: String?
    public var language: String?
    public var title: String?
    public var codec: String?
    public var channels: Int?
    public var `default`: Bool?
    public var forced: Bool?
    public var external: Bool?
    public var stableID: String { id ?? "\(kind ?? "track")-\(index ?? 0)-\(language ?? "und")" }
}

public struct PlaybackStartRequest: Codable {
    public var mediaSourceId: String
    public var clientProfile: String
    public var positionSeconds: Int?
    public var audioTrackIndex: Int?
    public var subtitleTrackIndex: Int?
    public var subtitleTrackActive: Bool?
    public var supportsAdaptive: Bool?
}

public struct PlaybackStartResponse: Codable, Equatable {
    public var sessionId: String?
    public var heartbeatUrl: String?
    public var stopUrl: String?
    public var playbackStateUrl: String?
    public var heartbeatIntervalMs: Int?
    public var decision: PlaybackDecision?
    public var route: PlaybackRoute?
}

public struct PlaybackRoute: Codable, Equatable {
    public var url: String?
    public var manifestUrl: String?
    public var protocolValue: String?
    public var decision: PlaybackDecision?

    enum CodingKeys: String, CodingKey {
        case url, manifestUrl, decision
        case protocolValue = "protocol"
    }

    public var streamURL: String? { manifestUrl ?? url }
}

public struct PlaybackDecision: Codable, Equatable {
    public var mode: String?
    public var videoAction: String?
    public var audioAction: String?
    public var containerAction: String?
    public var subtitleAction: String?
    public var reasonCode: String?
    public var reasonText: String?
    public var serverImpact: String?
}

public struct PlaybackHeartbeat: Codable {
    public var positionSeconds: Int
    public var isPaused: Bool?
    public var completed: Bool?
}

public extension DetailResponse {
    var displayTitle: String { metadata?.title ?? title ?? "Unknown Title" }
    var displayOverview: String { metadata?.overview ?? overview ?? "" }
    var displayPosterURL: String? { metadata?.posterUrl ?? posterUrl }
    var displayBackdropURL: String? { metadata?.backdropUrl ?? backdropUrl }
    var displayLogoURL: String? { metadata?.logoUrl ?? logoUrl }
    var displayGenres: [String] { metadata?.genres ?? genres ?? [] }
    var displayYear: Int? { metadata?.year ?? year }
    var displayRating: Double? { metadata?.voteAverage ?? rating }
    var displayTrailerPath: String? { metadata?.trailerPath }
    var displayVideoKey: String? { metadata?.videoKey }
    var displayCast: [MetadataCredit] { metadata?.cast ?? [] }
    var displayDirectors: [String] { metadata?.directors ?? [] }
    var displayWriters: [String] { metadata?.writers ?? [] }
    var displayStudios: [String] {
        var values: [String] = []
        values.append(contentsOf: metadata?.studios ?? [])
        values.append(contentsOf: metadata?.productionCompanies ?? [])
        values.append(contentsOf: metadata?.networks ?? [])
        var seen = Set<String>()
        return values.filter { seen.insert($0).inserted }.prefix(6).map { $0 }
    }
    var displayRuntime: String? {
        if let runtime = metadata?.runtime ?? runtime { return runtime }
        guard let minutes = metadata?.runtimeMinutes ?? runtimeMinutes else { return nil }
        return "\(minutes / 60)h \(minutes % 60)m"
    }
}

public extension PlaybackDecision {
    var badgeLabel: String {
        let modeText = (mode ?? "").lowercased()
        let video = (videoAction ?? "").lowercased()
        let audio = (audioAction ?? "").lowercased()
        let container = (containerAction ?? "").lowercased()
        if modeText == "adaptive" || modeText == "hls" { return "Adaptive" }
        if video == "transcode" || video == "encode" { return "Transcoding" }
        if container == "remux" || container == "mux" { return "Remux" }
        if modeText == "direct" || modeText == "direct_play" || modeText == "directplay" { return "Direct Play" }
        if audio == "transcode" || audio == "encode" { return "Audio Tx" }
        return "Route"
    }
}
