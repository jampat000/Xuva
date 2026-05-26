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
    public var eyebrow: String?
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
    /// 0.0–1.0 watched fraction. Server sends as `progressPercent` for
    /// Continue Watching rows; we decode either spelling so older payloads
    /// keep working.
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
    /// For Continue Watching rows the `id` is the mediaSource id; `parentId`
    /// + `parentKind` point at the canonical movie/series so detail navigation
    /// uses the right entity.
    public var parentId: String?
    public var parentKind: String?

    public var rating: Double? { voteAverage }
    public var routeLabel: String? { route }
    public var resolvedDetailId: String { parentId ?? id }
    public var resolvedDetailKind: String { parentKind ?? kind ?? "movie" }

    enum CodingKeys: String, CodingKey {
        case id, kind, title, subtitle, year, voteAverage, runtime, runtimeMinutes
        case progress
        case progressPercent
        case posterUrl, backdropUrl, imageUrl, thumbnailUrl, logoUrl, bannerUrl
        case mediaSourceId, route, versionCount, genres, overview, director
        case parentId, parentKind
    }

    public init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        id = try c.decode(String.self, forKey: .id)
        kind = try c.decodeIfPresent(String.self, forKey: .kind)
        title = try c.decodeIfPresent(String.self, forKey: .title)
        subtitle = try c.decodeIfPresent(String.self, forKey: .subtitle)
        year = try c.decodeIfPresent(Int.self, forKey: .year)
        voteAverage = try c.decodeIfPresent(Double.self, forKey: .voteAverage)
        runtime = try c.decodeIfPresent(String.self, forKey: .runtime)
        runtimeMinutes = try c.decodeIfPresent(Int.self, forKey: .runtimeMinutes)
        progress = try c.decodeIfPresent(Double.self, forKey: .progress)
            ?? c.decodeIfPresent(Double.self, forKey: .progressPercent)
        posterUrl = try c.decodeIfPresent(String.self, forKey: .posterUrl)
        backdropUrl = try c.decodeIfPresent(String.self, forKey: .backdropUrl)
        imageUrl = try c.decodeIfPresent(String.self, forKey: .imageUrl)
        thumbnailUrl = try c.decodeIfPresent(String.self, forKey: .thumbnailUrl)
        logoUrl = try c.decodeIfPresent(String.self, forKey: .logoUrl)
        bannerUrl = try c.decodeIfPresent(String.self, forKey: .bannerUrl)
        mediaSourceId = try c.decodeIfPresent(String.self, forKey: .mediaSourceId)
        route = try c.decodeIfPresent(String.self, forKey: .route)
        versionCount = try c.decodeIfPresent(Int.self, forKey: .versionCount)
        genres = try c.decodeIfPresent([String].self, forKey: .genres)
        overview = try c.decodeIfPresent(String.self, forKey: .overview)
        director = try c.decodeIfPresent(String.self, forKey: .director)
        parentId = try c.decodeIfPresent(String.self, forKey: .parentId)
        parentKind = try c.decodeIfPresent(String.self, forKey: .parentKind)
    }

    public func encode(to encoder: Encoder) throws {
        var c = encoder.container(keyedBy: CodingKeys.self)
        try c.encode(id, forKey: .id)
        try c.encodeIfPresent(kind, forKey: .kind)
        try c.encodeIfPresent(title, forKey: .title)
        try c.encodeIfPresent(subtitle, forKey: .subtitle)
        try c.encodeIfPresent(year, forKey: .year)
        try c.encodeIfPresent(voteAverage, forKey: .voteAverage)
        try c.encodeIfPresent(runtime, forKey: .runtime)
        try c.encodeIfPresent(runtimeMinutes, forKey: .runtimeMinutes)
        try c.encodeIfPresent(progress, forKey: .progress)
        try c.encodeIfPresent(posterUrl, forKey: .posterUrl)
        try c.encodeIfPresent(backdropUrl, forKey: .backdropUrl)
        try c.encodeIfPresent(imageUrl, forKey: .imageUrl)
        try c.encodeIfPresent(thumbnailUrl, forKey: .thumbnailUrl)
        try c.encodeIfPresent(logoUrl, forKey: .logoUrl)
        try c.encodeIfPresent(bannerUrl, forKey: .bannerUrl)
        try c.encodeIfPresent(mediaSourceId, forKey: .mediaSourceId)
        try c.encodeIfPresent(route, forKey: .route)
        try c.encodeIfPresent(versionCount, forKey: .versionCount)
        try c.encodeIfPresent(genres, forKey: .genres)
        try c.encodeIfPresent(overview, forKey: .overview)
        try c.encodeIfPresent(director, forKey: .director)
        try c.encodeIfPresent(parentId, forKey: .parentId)
        try c.encodeIfPresent(parentKind, forKey: .parentKind)
    }

    public init(id: String, kind: String? = nil, title: String? = nil, subtitle: String? = nil, year: Int? = nil, genres: [String]? = nil, overview: String? = nil) {
        self.id = id
        self.kind = kind
        self.title = title
        self.subtitle = subtitle
        self.year = year
        self.genres = genres
        self.overview = overview
    }
}

public struct DetailResponse: Codable, Equatable {
    public var defaultMediaSourceId: String?
    public var item: DetailItem?
    public var versions: [MediaVersion]?
    /// Present for series detail responses — array of seasons each with their
    /// own episode list. Movies have versions[]; series have seasons[].
    public var seasons: [SeasonItem]?
    /// Populated client-side after a follow-up call to `/api/metadata/{kind}/{id}`.
    /// The /api/client/* detail endpoint omits cast / writers / studios; we merge them in
    /// once the metadata records resolve so the UI sees a single object.
    public var enrichedMetadata: MetadataRecord?
    /// Populated client-side from `GET /api/client/{kind}/{id}/similar`.
    public var relatedTitles: [SimilarItem]?

    public var isSeries: Bool { (seasons?.isEmpty == false) || item?.kind?.lowercased().contains("series") == true }
}

public struct SeasonItem: Codable, Equatable, Identifiable {
    public var seasonNumber: Int?
    public var name: String?
    public var airDate: String?
    public var overview: String?
    public var posterUrl: String?
    public var backdropUrl: String?
    public var episodes: [EpisodeItem]?

    public var id: Int { seasonNumber ?? 0 }
    public var displayTitle: String { name ?? (seasonNumber.map { "Season \($0)" } ?? "Season") }
}

public struct EpisodeItem: Codable, Equatable, Identifiable {
    public var id: String
    public var seasonNumber: Int?
    public var episodeNumber: Int?
    public var title: String?
    public var overview: String?
    public var airDate: String?
    public var runtimeMinutes: Int?
    public var thumbnailUrl: String?
    public var versionCount: Int?
    public var versions: [MediaVersion]?
    /// 0.0–1.0 watched fraction. Populated by the server when the episode
    /// has playback history. Used to pre-seek on resume.
    public var progress: Double?
    /// Absolute resume position in seconds (alternative to fraction).
    /// Takes precedence over progress × duration when non-nil and > 0.
    public var positionSeconds: Int?

    public var displayTitle: String {
        if let n = episodeNumber, let t = title { return "E\(n) · \(t)" }
        if let n = episodeNumber { return "Episode \(n)" }
        return title ?? "Episode"
    }
    public var displayRuntime: String? {
        guard let m = runtimeMinutes, m > 0 else { return nil }
        let h = m / 60
        let rem = m % 60
        if h > 0 { return "\(h)h \(rem)m" }
        return "\(rem)m"
    }
    public var defaultMediaSourceId: String? { versions?.first?.mediaSourceId }
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

public struct MetadataRecord: Codable, Equatable {
    public var provider: String?
    public var externalId: String?
    public var title: String?
    public var year: Int?
    public var overview: String?
    public var tagline: String?
    public var posterUrl: String?
    public var backdropUrl: String?
    public var logoUrl: String?
    public var thumbnailUrl: String?
    public var videoKey: String?
    public var trailerPath: String?
    public var runtimeMinutes: Int?
    public var genres: [String]?
    public var contentRating: String?
    public var voteAverage: Double?
    public var cast: [MetadataCredit]?
    public var directors: [String]?
    public var writers: [String]?
    public var studios: [String]?
    public var productionCompanies: [String]?
    public var networks: [String]?
}

public struct MetadataCredit: Codable, Equatable, Identifiable {
    public var name: String?
    public var role: String?
    public var character: String?
    public var profileUrl: String?
    public var sortOrder: Int?

    public var id: String { name ?? UUID().uuidString }
}

public struct MetadataRecordsResponse: Codable, Equatable {
    public var best: MetadataRecord?
    public var records: [MetadataRecord]?
    public var providers: [String]?
}

/// A lightweight title stub used in the "More like this" similar-titles row.
public struct SimilarItem: Codable, Equatable, Identifiable {
    public var itemId: String?
    public var kind: String?
    public var title: String?
    public var year: Int?
    public var posterUrl: String?

    public var id: String { itemId ?? title ?? UUID().uuidString }

    private enum CodingKeys: String, CodingKey {
        case itemId = "id"
        case kind, title, year, posterUrl
    }
}

public struct SimilarResponse: Codable, Equatable {
    public var items: [SimilarItem]?
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
    /// Request HLS segment-based playback regardless of network conditions.
    /// Gives instant seeking on Siri Remote scrubbing; server still exempts
    /// Dolby Vision pass-through files where transcoding would strip DV metadata.
    public var preferAdaptive: Bool?
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
    public var deviceId: String?
    public var defaultSubtitlesEnabled: Bool?
    /// Filled in by the client (not the server) — the requested resume
    /// position. AVPlayer seeks to this once the player item is ready.
    public var clientStartPositionSeconds: Int?
    /// Filled in by the client — the subtitle track the user picked in the
    /// detail screen. XuvaVideoPlayer selects the matching AVMediaSelectionOption
    /// when the player item is ready.
    public var clientSubtitleTrack: MediaTrack?
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

public struct StreamTokenRequest: Codable {
    public var sessionId: String
    public var deviceId: String
}

public struct StreamTokenResponse: Codable, Equatable {
    public var token: String?
    public var expiresAt: String?
    public var streamUrl: String?
    public var subtitleBaseUrl: String?
    public var query: String?
}

public extension DetailResponse {
    var displayTitle: String { enrichedMetadata?.title ?? item?.title ?? "Unknown Title" }
    var displayOverview: String { enrichedMetadata?.overview ?? item?.overview ?? "" }
    var displayPosterURL: String? { enrichedMetadata?.posterUrl ?? item?.posterUrl }
    var displayBackdropURL: String? { enrichedMetadata?.backdropUrl ?? item?.backdropUrl ?? item?.thumbnailUrl }
    var displayLogoURL: String? { enrichedMetadata?.logoUrl ?? item?.logoUrl }
    var displayGenres: [String] { enrichedMetadata?.genres ?? item?.genres ?? [] }
    var displayYear: Int? { enrichedMetadata?.year ?? item?.year }
    var displayRating: Double? { enrichedMetadata?.voteAverage ?? item?.voteAverage }
    var displayTrailerPath: String? { item?.trailerUrl }
    var displayVideoKey: String? { enrichedMetadata?.videoKey ?? item?.videoKey }
    var displayDirectors: [String] {
        if let directors = enrichedMetadata?.directors, !directors.isEmpty { return directors }
        let d = (item?.director ?? "").trimmingCharacters(in: .whitespacesAndNewlines)
        return d.isEmpty ? [] : [d]
    }
    var displayWriters: [String] { enrichedMetadata?.writers ?? [] }
    var displayCast: [MetadataCredit] {
        let cast = enrichedMetadata?.cast ?? []
        return cast.sorted { ($0.sortOrder ?? Int.max) < ($1.sortOrder ?? Int.max) }
    }
    var displayStudios: [String] {
        var seen = Set<String>()
        var all: [String] = []
        all.append(contentsOf: enrichedMetadata?.studios ?? [])
        all.append(contentsOf: enrichedMetadata?.productionCompanies ?? [])
        all.append(contentsOf: enrichedMetadata?.networks ?? [])
        return all
            .filter { !$0.isEmpty }
            .filter { seen.insert($0).inserted }
    }
    var displayContentRating: String? { enrichedMetadata?.contentRating ?? item?.contentRating }
    var displayTagline: String? { enrichedMetadata?.tagline ?? item?.tagline }
    var displayRuntime: String? {
        let minutes = enrichedMetadata?.runtimeMinutes ?? item?.runtimeMinutes ?? 0
        guard minutes > 0 else { return nil }
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
