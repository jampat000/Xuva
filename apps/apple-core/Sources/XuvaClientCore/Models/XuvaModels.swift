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
    public var voteAverage: Double?
    public var runtime: String?
    public var runtimeMinutes: Int?
    public var progress: Double?
    public var posterUrl: String?
    public var backdropUrl: String?
    public var imageUrl: String?
    public var thumbnailUrl: String?
    public var logoUrl: String?
    public var bannerUrl: String?
    public var mediaSourceId: String?
    public var route: String?
    public var versionCount: Int?
    public var genres: [String]?
    public var overview: String?
    public var director: String?

    public var rating: Double? { voteAverage }
    public var routeLabel: String? { route }
}

public struct DetailResponse: Codable, Equatable {
    public var defaultMediaSourceId: String?
    public var item: DetailItem?
    public var versions: [MediaVersion]?
}

public struct DetailItem: Codable, Equatable {
    public var id: String?
    public var kind: String?
    public var title: String?
    public var subtitle: String?
    public var tagline: String?
    public var overview: String?
    public var year: Int?
    public var runtimeMinutes: Int?
    public var voteAverage: Double?
    public var contentRating: String?
    public var genres: [String]?
    public var posterUrl: String?
    public var backdropUrl: String?
    public var thumbnailUrl: String?
    public var logoUrl: String?
    public var bannerUrl: String?
    public var director: String?
    public var trailerUrl: String?
    public var videoKey: String?
    public var versionCount: Int?
}

public struct MediaVersion: Codable, Identifiable, Equatable {
    public var id: String?
    public var mediaSourceId: String?
    public var qualityLabel: String?
    public var path: String?
    public var name: String?
    public var audioTracks: [MediaTrack]?
    public var subtitleTracks: [MediaTrack]?
    public var sidecars: [SubtitleSidecar]?
    public var decision: PlaybackDecision?
    public var source: MediaSource?

    public var stableID: String { mediaSourceId ?? id ?? name ?? path ?? UUID().uuidString }

    public var displayResolution: String? {
        if let w = source?.width, let h = source?.height, w > 0, h > 0 { return "\(w)×\(h)" }
        return decision?.selected?["resolution"]
    }

    public var displayVideoCodec: String? {
        if let codec = source?.videoCodec, !codec.isEmpty { return codec.uppercased() }
        return decision?.selected?["videoCodec"]?.uppercased()
    }

    public var displayAudioSummary: String? {
        if let track = audioTracks?.first {
            let lang = track.language?.uppercased() ?? ""
            let codec = (track.codec ?? "").uppercased()
            let channels = track.channels.map { "\($0)ch" } ?? ""
            return [lang, codec, channels].filter { !$0.isEmpty }.joined(separator: " · ")
        }
        return nil
    }

    public var displayBitrate: String? {
        guard let bitrate = source?.bitrate, bitrate > 0 else { return nil }
        let mbps = Double(bitrate) / 1_000_000
        return String(format: "%.1f Mbps", mbps)
    }

    public var displayDuration: String? {
        guard let seconds = source?.durationSeconds, seconds > 0 else { return nil }
        let total = Int(seconds)
        let hours = total / 3600
        let minutes = (total % 3600) / 60
        if hours > 0 { return "\(hours)h \(minutes)m" }
        return "\(minutes)m"
    }

    public var displaySize: String? {
        guard let bytes = source?.sizeBytes, bytes > 0 else { return nil }
        let gb = Double(bytes) / (1024 * 1024 * 1024)
        return String(format: "%.1f GB", gb)
    }
}

public struct MediaSource: Codable, Equatable {
    public var id: String?
    public var libraryId: String?
    public var kind: String?
    public var path: String?
    public var relPath: String?
    public var name: String?
    public var `extension`: String?
    public var sizeBytes: Int64?
    public var container: String?
    public var durationSeconds: Double?
    public var bitrate: Int?
    public var videoCodec: String?
    public var videoProfile: String?
    public var videoLevel: String?
    public var videoBitDepth: Int?
    public var videoFrameRate: Double?
    public var pixelFormat: String?
    public var width: Int?
    public var height: Int?
    public var audioStreams: Int?
    public var probed: Bool?
}

public struct SubtitleSidecar: Codable, Equatable {
    public var path: String?
    public var relPath: String?
    public var language: String?
    public var format: String?
    public var forced: Bool?
    public var hearingImpaired: Bool?
    public var requiresVideoBurn: Bool?
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
    public var mediaSourceId: String?
    public var defaultSubtitlesEnabled: Bool?
}

public struct PlaybackRoute: Codable, Equatable {
    public var url: String?
    public var manifestUrl: String?
    public var protocolValue: String?
    public var route: String?
    public var status: String?
    public var decision: PlaybackDecision?

    enum CodingKeys: String, CodingKey {
        case url, manifestUrl, decision, route, status
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
    public var reason: String?
    public var serverImpact: String?
    public var estimatedCpuCost: String?
    public var estimatedGpuCost: String?
    public var estimatedNetworkBitrate: Int?
    public var selected: [String: String]?
}

public struct PlaybackHeartbeat: Codable {
    public var positionSeconds: Int
    public var isPaused: Bool?
    public var completed: Bool?
}

public extension DetailResponse {
    var displayTitle: String { item?.title ?? "Unknown Title" }
    var displayOverview: String { item?.overview ?? "" }
    var displayPosterURL: String? { item?.posterUrl }
    var displayBackdropURL: String? { item?.backdropUrl ?? item?.thumbnailUrl }
    var displayLogoURL: String? { item?.logoUrl }
    var displayGenres: [String] { item?.genres ?? [] }
    var displayYear: Int? { item?.year }
    var displayRating: Double? { item?.voteAverage }
    var displayTrailerPath: String? { item?.trailerUrl }
    var displayVideoKey: String? { item?.videoKey }
    var displayDirectors: [String] {
        let d = (item?.director ?? "").trimmingCharacters(in: .whitespacesAndNewlines)
        return d.isEmpty ? [] : [d]
    }
    var displayContentRating: String? { item?.contentRating }
    var displayTagline: String? { item?.tagline }
    var displayRuntime: String? {
        guard let minutes = item?.runtimeMinutes, minutes > 0 else { return nil }
        let h = minutes / 60
        let m = minutes % 60
        if h > 0 { return "\(h)h \(m)m" }
        return "\(m)m"
    }
    var kind: String? { item?.kind }

    var audioTracks: [MediaTrack] {
        versions?.first?.audioTracks ?? []
    }
    var subtitleTracks: [MediaTrack] {
        let embedded = versions?.first?.subtitleTracks ?? []
        let sidecarTracks: [MediaTrack] = (versions?.first?.sidecars ?? []).enumerated().map { idx, sidecar in
            MediaTrack(
                id: sidecar.relPath ?? sidecar.path,
                index: 100 + idx,
                kind: "subtitle",
                language: sidecar.language,
                title: sidecar.language?.uppercased() ?? "External",
                codec: sidecar.format,
                channels: nil,
                default: false,
                forced: sidecar.forced ?? false,
                external: true
            )
        }
        return embedded + sidecarTracks
    }
}

public extension PlaybackDecision {
    var badgeLabel: String {
        let modeText = (mode ?? "").lowercased()
        let video = (videoAction ?? "").lowercased()
        let audio = (audioAction ?? "").lowercased()
        let container = (containerAction ?? "").lowercased()
        if modeText.contains("direct") { return "Direct Play" }
        if modeText.contains("adaptive") || modeText.contains("hls") { return "Adaptive" }
        if video == "transcode" || video == "encode" { return "Transcoding" }
        if container == "remux" || container == "mux" { return "Remux" }
        if audio == "transcode" || audio == "encode" { return "Audio Tx" }
        if modeText.contains("deferred") || modeText.contains("probe") { return "Pending" }
        return mode?.capitalized ?? "Route"
    }
}
