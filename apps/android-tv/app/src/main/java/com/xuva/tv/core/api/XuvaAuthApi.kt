package com.xuva.tv.core.api

import com.xuva.tv.core.model.HomePayload
import com.xuva.tv.core.storage.SecureStore
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
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
    suspend fun getHome(profile: String = "android-tv", limit: Int = 24): HomePayload =
        withContext(Dispatchers.IO) {
            val baseUrl = requireBaseUrl()
            val request = authed("$baseUrl/api/client/home?clientProfile=$profile&limit=$limit").build()
            client.newCall(request).execute().use { response ->
                val text = response.body?.string().orEmpty()
                check(response.isSuccessful) { "getHome failed: ${response.code} $text" }
                XuvaJson.decodeFromString(HomePayload.serializer(), text)
            }
        }

    private fun authed(url: String): Request.Builder {
        val token = secureStore.sessionToken
            ?: error("not paired — sessionToken is null")
        return Request.Builder().url(url).get().header("X-Auth-Token", token)
    }

    private fun requireBaseUrl(): String =
        secureStore.baseUrl ?: error("not paired — baseUrl is null")

    companion object {
        val defaultClient: OkHttpClient = OkHttpClient.Builder()
            .connectTimeout(5, TimeUnit.SECONDS)
            .readTimeout(15, TimeUnit.SECONDS)
            .build()
    }
}
