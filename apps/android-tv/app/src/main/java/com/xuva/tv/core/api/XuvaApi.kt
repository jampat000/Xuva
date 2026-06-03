package com.xuva.tv.core.api

import com.xuva.tv.core.model.CreateRequest
import com.xuva.tv.core.model.PairingItem
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.serialization.encodeToString
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import java.util.concurrent.TimeUnit

class XuvaApi(
    private val client: OkHttpClient = defaultClient,
) {
    /**
     * Create a pairing request for this device. Server returns the request
     * with a status of "pending" and a 4-letter code that the user reads
     * into the web admin to approve. status transitions to "approved" with
     * an auth grant populated once approved server-side.
     */
    suspend fun createPairingRequest(baseUrl: String, body: CreateRequest): PairingItem =
        withContext(Dispatchers.IO) {
            val payload = XuvaJson.encodeToString(body)
            val request = Request.Builder()
                .url("$baseUrl/api/pairing/requests")
                .post(payload.toRequestBody(JSON))
                .build()
            client.newCall(request).execute().use { response ->
                val text = response.body?.string().orEmpty()
                check(response.isSuccessful) { "createPairingRequest failed: ${response.code} $text" }
                XuvaJson.decodeFromString(PairingItem.serializer(), text)
            }
        }

    /**
     * Poll the status of an existing pairing request. Returns the updated
     * item; when status == "approved" the auth grant contains a session
     * token to use as `X-Auth-Token` on every subsequent request.
     */
    suspend fun getPairingStatus(baseUrl: String, id: String): PairingItem =
        withContext(Dispatchers.IO) {
            val request = Request.Builder()
                .url("$baseUrl/api/pairing/requests/$id")
                .get()
                .build()
            client.newCall(request).execute().use { response ->
                val text = response.body?.string().orEmpty()
                check(response.isSuccessful) { "getPairingStatus failed: ${response.code} $text" }
                XuvaJson.decodeFromString(PairingItem.serializer(), text)
            }
        }

    /**
     * Withdraw a pending pairing request. Safe to call after approval — the
     * server reports a 409 in that case which we treat as success.
     */
    suspend fun cancelPairingRequest(baseUrl: String, id: String, deviceId: String) =
        withContext(Dispatchers.IO) {
            val request = Request.Builder()
                .url("$baseUrl/api/pairing/requests/$id?deviceId=$deviceId")
                .delete()
                .build()
            client.newCall(request).execute().use { /* drain + close */ }
        }

    /**
     * Quick liveness probe used when the user types a server URL manually,
     * to fail fast on typos before kicking off the pairing flow.
     */
    suspend fun probe(baseUrl: String): Boolean = withContext(Dispatchers.IO) {
        val request = Request.Builder()
            .url("$baseUrl/api/health")
            .get()
            .build()
        runCatching {
            client.newCall(request).execute().use { it.isSuccessful }
        }.getOrDefault(false)
    }

    companion object {
        private val JSON = "application/json; charset=utf-8".toMediaType()

        val defaultClient: OkHttpClient = OkHttpClient.Builder()
            .connectTimeout(5, TimeUnit.SECONDS)
            .readTimeout(10, TimeUnit.SECONDS)
            .build()
    }
}
