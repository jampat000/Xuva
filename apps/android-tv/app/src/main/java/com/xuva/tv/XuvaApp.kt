package com.xuva.tv

import android.app.Application
import com.xuva.tv.core.api.XuvaApi
import com.xuva.tv.core.api.XuvaAuthApi
import com.xuva.tv.core.discovery.ServerDiscovery
import com.xuva.tv.core.pairing.PairingRepository
import com.xuva.tv.core.storage.SecureStore

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
    val authApi: XuvaAuthApi by lazy { XuvaAuthApi(secureStore) }
}
