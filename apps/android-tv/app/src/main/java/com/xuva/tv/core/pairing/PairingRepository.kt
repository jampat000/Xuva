package com.xuva.tv.core.pairing

import com.xuva.tv.core.api.XuvaApi
import com.xuva.tv.core.model.CreateRequest
import com.xuva.tv.core.model.PairingItem
import com.xuva.tv.core.storage.SecureStore
import kotlinx.coroutines.delay

/**
 * Orchestrates the pairing flow against the server.
 *
 * Lifecycle of a pairing:
 *   1. `start(baseUrl, deviceName)` — POST /api/pairing/requests, returns the
 *      PairingItem with `code` and `status="pending"`.
 *   2. `await(item)` — poll GET /api/pairing/requests/{id} every 2s until
 *      status transitions away from "pending". On "approved" the returned item
 *      includes an AuthGrant; we persist the session token + base URL via
 *      SecureStore so the next app launch lands paired.
 *   3. `cancel(item)` — DELETE the pending request when the user backs out.
 *
 * All HTTP errors propagate as exceptions for the ViewModel to render.
 */
class PairingRepository(
    private val api: XuvaApi,
    private val secureStore: SecureStore,
) {
    suspend fun start(baseUrl: String, deviceName: String, profile: String = "android-tv"): PairingItem {
        return api.createPairingRequest(
            baseUrl = baseUrl,
            body = CreateRequest(
                deviceName = deviceName,
                clientProfile = profile,
                deviceId = secureStore.deviceId,
            ),
        )
    }

    /**
     * Polls until the pairing reaches a terminal state. Returns the final
     * PairingItem (status is one of approved/denied/expired/cancelled). On
     * "approved" the auth grant is also persisted to SecureStore — caller
     * just needs to read [SecureStore.isPaired].
     */
    suspend fun await(baseUrl: String, id: String, pollIntervalMs: Long = 2_000): PairingItem {
        while (true) {
            val item = api.getPairingStatus(baseUrl, id)
            when (item.status) {
                "approved" -> {
                    val token = item.auth?.sessionToken
                    if (!token.isNullOrBlank()) {
                        secureStore.baseUrl = baseUrl
                        secureStore.sessionToken = token
                    }
                    return item
                }
                "denied", "expired", "cancelled" -> return item
                else -> delay(pollIntervalMs)
            }
        }
    }

    suspend fun cancel(baseUrl: String, id: String) {
        runCatching { api.cancelPairingRequest(baseUrl, id, secureStore.deviceId) }
    }

    fun unpair() {
        secureStore.clearPairing()
    }
}
