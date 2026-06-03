package com.xuva.tv

import android.app.Application
import com.xuva.tv.data.api.XuvaApi
import com.xuva.tv.data.discovery.ServerDiscovery
import com.xuva.tv.data.pairing.PairingRepository
import com.xuva.tv.data.storage.SecureStore

/**
 * Manual DI container. Tiny scope (single screen at this point) — Hilt/Koin
 * are overkill until the graph grows. Wired by AndroidManifest's
 * `android:name=".XuvaApp"`.
 */
class XuvaApp : Application() {

    val secureStore: SecureStore by lazy { SecureStore(applicationContext) }
    val api: XuvaApi by lazy { XuvaApi() }
    val pairingRepository: PairingRepository by lazy { PairingRepository(api, secureStore) }
    val serverDiscovery: ServerDiscovery by lazy { ServerDiscovery(applicationContext) }
}
