package com.xuva.tv.core.model

import kotlinx.serialization.Serializable

@Serializable
data class HomePayload(
    val profile: String? = null,
    val heroes: List<HomeItem> = emptyList(),
    val rows: List<HomeRow> = emptyList(),
)

@Serializable
data class HomeRow(
    val id: String,
    val title: String,
    val eyebrow: String? = null,
    val items: List<HomeItem> = emptyList(),
)

@Serializable
data class HomeItem(
    val id: String,
    val kind: String,
    val title: String,
    val subtitle: String? = null,
    val year: Int = 0,
    val posterUrl: String? = null,
    val backdropUrl: String? = null,
    val logoUrl: String? = null,
    val thumbnailUrl: String? = null,
    val voteAverage: Double? = null,
    val genres: List<String> = emptyList(),
)
