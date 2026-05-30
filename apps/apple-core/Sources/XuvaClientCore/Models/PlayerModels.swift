import Foundation

// MARK: - Track listing
//
// Mirrors the server's `GET /api/media-sources/{id}/tracks` payload
// (catalog.MediaSourceTracks → probe.Track). MediaTrack already decodes the
// probe.Track JSON shape (index/codec/language/title/channels/forced/default),
// so we reuse it here rather than introducing a parallel type.

public struct MediaSourceTracksResponse: Codable, Equatable {
    public var audioTracks: [MediaTrack]?
    public var subtitleTracks: [MediaTrack]?

    public init(audioTracks: [MediaTrack]? = nil, subtitleTracks: [MediaTrack]? = nil) {
        self.audioTracks = audioTracks
        self.subtitleTracks = subtitleTracks
    }
}

// MARK: - Chapters (intro / credits markers)
//
// Mirrors the server's `GET /api/media-sources/{id}/chapters` payload
// (chapters.Chapters → chapters.Segment). Drives the Skip Intro button and
// the Credits marker, matching the web player's behaviour.

public struct ChapterSegment: Codable, Equatable {
    public var start: Double
    public var end: Double

    public init(start: Double, end: Double) {
        self.start = start
        self.end = end
    }
}

public struct ChaptersResponse: Codable, Equatable {
    public var mediaSourceId: String?
    public var intro: ChapterSegment?
    public var credits: ChapterSegment?
    public var analyzedAt: String?

    public init(mediaSourceId: String? = nil, intro: ChapterSegment? = nil, credits: ChapterSegment? = nil, analyzedAt: String? = nil) {
        self.mediaSourceId = mediaSourceId
        self.intro = intro
        self.credits = credits
        self.analyzedAt = analyzedAt
    }
}

// MARK: - Playback state write
//
// Mirrors the server's `PUT /api/playback/state/{id}` body (playstate.Update).
// Written when playback ends so Continue Watching / watched state reflects the
// final position — the web player does the same on `ended` and on unload.

public struct PlaybackStateUpdate: Codable {
    public var progressSeconds: Double
    public var durationSeconds: Double
    public var watched: Bool?

    public init(progressSeconds: Double, durationSeconds: Double, watched: Bool? = nil) {
        self.progressSeconds = progressSeconds
        self.durationSeconds = durationSeconds
        self.watched = watched
    }
}
