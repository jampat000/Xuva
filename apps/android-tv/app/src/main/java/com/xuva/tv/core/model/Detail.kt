package com.xuva.tv.core.model

import kotlinx.serialization.Serializable

// Mirror of `catalog.MovieVersion` — a single playable file backing a movie.
// The mediaSourceId is what we pass to /api/playback/route to get a stream URL.
@Serializable
data class MediaVersion(
    val mediaSourceId: String,
    val path: String? = null,
    val relPath: String? = null,
    val edition: String? = null,
    val qualityLabel: String? = null,
    val sizeBytes: Long = 0,
    val modifiedAt: String? = null,
)

@Serializable
data class MetadataLite(
    val title: String? = null,
    val overview: String? = null,
    val year: Int = 0,
    val genres: List<String> = emptyList(),
    val posterUrl: String? = null,
    val backdropUrl: String? = null,
    val logoUrl: String? = null,
)

@Serializable
data class MovieDetail(
    val id: String,
    val title: String,
    val year: Int = 0,
    val versions: List<MediaVersion> = emptyList(),
    val metadata: MetadataLite? = null,
) {
    fun playableMediaSourceId(): String? = versions.firstOrNull()?.mediaSourceId
}

@Serializable
data class EpisodeBrief(
    val id: String,
    val seasonNumber: Int = 0,
    val episodeNumber: Int = 0,
    val title: String? = null,
    val versionCount: Int = 0,
    val versions: List<MediaVersion> = emptyList(),
    val metadata: MetadataLite? = null,
) {
    fun playableMediaSourceId(): String? = versions.firstOrNull()?.mediaSourceId
}

@Serializable
data class SeasonDetail(
    val id: String,
    val seasonNumber: Int = 0,
    val episodes: List<EpisodeBrief> = emptyList(),
    val metadata: MetadataLite? = null,
)

@Serializable
data class SeriesDetail(
    val id: String,
    val title: String,
    val seasonCount: Int = 0,
    val episodeCount: Int = 0,
    val seasons: List<SeasonDetail> = emptyList(),
    val metadata: MetadataLite? = null,
)

// The /api/playback/route response. Different `route` values populate
// different URL fields — direct/remux use `url`, HLS variants use
// `manifestUrl`. Lenient field set covers all server branches.
@Serializable
data class PlaybackRoute(
    val route: String,
    val status: String? = null,
    val url: String? = null,
    val manifestUrl: String? = null,
    val protocol: String? = null,
) {
    /** Absolute path the player should open (relative to baseUrl). */
    fun streamPath(): String? = manifestUrl ?: url

    val isHls: Boolean get() =
        protocol.equals("hls", ignoreCase = true) || manifestUrl != null

    val isPlayable: Boolean get() =
        status == "ready" && !streamPath().isNullOrBlank()
}
