package com.xuva.tv.core.storage

import android.content.Context
import android.content.SharedPreferences
import androidx.security.crypto.EncryptedSharedPreferences
import androidx.security.crypto.MasterKey
import java.util.UUID

/**
 * Persisted, encrypted-at-rest storage for the small bits of state we MUST
 * carry across launches:
 *
 *  - `deviceId`   — UUID generated once on first launch and reused forever;
 *                   used to identify this physical device to the server
 *                   (lets us cancel our own pairing request, lets the server
 *                   recognise us across pairings).
 *  - `baseUrl`    — the URL of the server we paired with, used to construct
 *                   every subsequent request.
 *  - `sessionToken` — the X-Auth-Token from the approved pairing grant.
 *
 * EncryptedSharedPreferences is backed by the Android Keystore, so the
 * blob on disk is unreadable without the device unlocked.
 */
class SecureStore(context: Context) {
    private val prefs: SharedPreferences = run {
        val masterKey = MasterKey.Builder(context)
            .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
            .build()
        EncryptedSharedPreferences.create(
            context,
            "xuva-secure",
            masterKey,
            EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
            EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM,
        )
    }

    /** Stable per-install UUID used as deviceId in pairing requests. */
    val deviceId: String
        get() = prefs.getString(KEY_DEVICE_ID, null) ?: UUID.randomUUID().toString().also {
            prefs.edit().putString(KEY_DEVICE_ID, it).apply()
        }

    var baseUrl: String?
        get() = prefs.getString(KEY_BASE_URL, null)
        set(value) {
            prefs.edit().apply {
                if (value == null) remove(KEY_BASE_URL) else putString(KEY_BASE_URL, value)
            }.apply()
        }

    var sessionToken: String?
        get() = prefs.getString(KEY_SESSION_TOKEN, null)
        set(value) {
            prefs.edit().apply {
                if (value == null) remove(KEY_SESSION_TOKEN) else putString(KEY_SESSION_TOKEN, value)
            }.apply()
        }

    val isPaired: Boolean get() = baseUrl != null && sessionToken != null

    fun clearPairing() {
        prefs.edit()
            .remove(KEY_BASE_URL)
            .remove(KEY_SESSION_TOKEN)
            .apply()
    }

    companion object {
        private const val KEY_DEVICE_ID = "device_id"
        private const val KEY_BASE_URL = "base_url"
        private const val KEY_SESSION_TOKEN = "session_token"
    }
}
