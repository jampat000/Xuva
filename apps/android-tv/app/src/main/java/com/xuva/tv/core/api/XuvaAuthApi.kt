package com.xuva.tv.core.api

import com.xuva.tv.core.model.HomePayload
import com.xuva.tv.core.model.MovieDetail
import com.xuva.tv.core.model.PlaybackRoute
import com.xuva.tv.core.model.SeriesDetail
import com.xuva.tv.core.storage.SecureStore
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.serialization.KSerializer
import okhttp3.OkHttpClient
import okhttp3.Request
import java.util.concurrent.TimeUnit

/**
 * Authenticated API client. Reads `baseUrl` + `sessionToken` from
 * [SecureStore] for every call, attaches `X-Auth-Token`, returns the
 * decoded payload. Throws if unpaired.
 */
class XuvaAuthApi(
    private val secureStore: SecureStore,
    private val client: OkHttpClient = defaultClient,
) {
    /** Resolved base URL for the paired server. Used by the player to
     *  prefix the relative stream URL returned by /api/playback/route. */
    val baseUrl: String get() = requireBaseUrl()

    /** Resolved session token. Used by the player's HTTP data source so
     *  every HLS segment fetch carries X-Auth-Token. */
    val sessionToken: String get() = requireSessionToken()

    suspend fun getHome(profile: String = "android-tv", limit: Int = 24): HomePayload =
        get("/api/client/home?clientProfile=$profile&limit=$limit", HomePayload.serializer())

    suspend fun getMovieDetail(id: String): MovieDetail =
        get("/api/movies/${encode(id)}", MovieDetail.serializer())

    suspend fun getSeriesDetail(id: String): SeriesDetail =
        get("/api/series/${encode(id)}", SeriesDetail.serializer())

    suspend fun getPlaybackRoute(mediaSourceId: String, profile: String = "android-tv"): PlaybackRoute =
        get(
            "/api/playback/route?mediaSourceId=${encode(mediaSourceId)}&clientProfile=$profile",
            PlaybackRoute.serializer(),
        )

    private suspend fun <T> get(path: String, serializer: KSerializer<T>): T =
        withContext(Dispatchers.IO) {
            val request = authed("$baseUrl$path").build()
            client.newCall(request).execute().use { response ->
                val text = response.body?.string().orEmpty()
                check(response.isSuccessful) { "GET $path failed: ${response.code} $text" }
                XuvaJson.decodeFromString(serializer, text)
            }
        }

    private fun authed(url: String): Request.Builder =
        Request.Builder().url(url).get().header("X-Auth-Token", requireSessionToken())

    private fun requireBaseUrl(): String =
        secureStore.baseUrl ?: error("not paired — baseUrl is null")

    private fun requireSessionToken(): String =
        secureStore.sessionToken ?: error("not paired — sessionToken is null")

    private fun encode(value: String): String =
        java.net.URLEncoder.encode(value, "UTF-8")

    companion object {
        val defaultClient: OkHttpClient = OkHttpClient.Builder()
            .connectTimeout(5, TimeUnit.SECONDS)
            .readTimeout(15, TimeUnit.SECONDS)
            .build()
    }
}
