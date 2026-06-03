package com.xuva.tv.core.discovery

import android.content.Context
import android.net.nsd.NsdManager
import android.net.nsd.NsdServiceInfo
import android.os.Build
import android.util.Log
import com.xuva.tv.core.model.DiscoveredServer
import kotlinx.coroutines.channels.awaitClose
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.callbackFlow

/**
 * Wraps Android's NsdManager into a Flow of `DiscoveredServer`. Emits one item
 * per resolved Xuva server on the LAN; collection stops discovery and releases
 * NsdManager listeners. Same wire contract as the Apple TV client:
 *
 *   service type:  _xuva._tcp
 *   domain:        local.
 *   port:          HTTP port (usually 8097)
 *
 * Multicast packets are interface-specific — on some Android TVs the multicast
 * lock must be acquired (CHANGE_WIFI_MULTICAST_STATE permission is declared in
 * the manifest). NsdManager handles that internally on recent platforms.
 */
class ServerDiscovery(private val context: Context) {

    fun discover(): Flow<DiscoveredServer> = callbackFlow {
        val nsdManager = context.getSystemService(Context.NSD_SERVICE) as NsdManager
        val seen = mutableSetOf<String>()

        val resolveListener = object : NsdManager.ResolveListener {
            override fun onResolveFailed(serviceInfo: NsdServiceInfo?, errorCode: Int) {
                Log.w(TAG, "resolve failed: $errorCode for ${serviceInfo?.serviceName}")
            }

            override fun onServiceResolved(serviceInfo: NsdServiceInfo?) {
                val info = serviceInfo ?: return
                val host = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.UPSIDE_DOWN_CAKE) {
                    info.hostAddresses?.firstOrNull()?.hostAddress
                } else {
                    @Suppress("DEPRECATION")
                    info.host?.hostAddress
                } ?: return
                val key = "$host:${info.port}"
                if (!seen.add(key)) return
                val name = info.serviceName?.takeIf { it.isNotBlank() } ?: "Xuva on $host"
                trySend(DiscoveredServer(name = name, host = host, port = info.port))
            }
        }

        val discoveryListener = object : NsdManager.DiscoveryListener {
            override fun onDiscoveryStarted(serviceType: String?) {
                Log.d(TAG, "discovery started: $serviceType")
            }

            override fun onServiceFound(serviceInfo: NsdServiceInfo?) {
                val info = serviceInfo ?: return
                // resolveService is one-shot; for each found service spin up a
                // resolve. NsdManager queues these internally so we don't
                // need our own backpressure.
                nsdManager.resolveService(info, resolveListener)
            }

            override fun onServiceLost(serviceInfo: NsdServiceInfo?) {
                // No-op: we keep the entry around until discovery restarts.
            }

            override fun onDiscoveryStopped(serviceType: String?) {
                Log.d(TAG, "discovery stopped: $serviceType")
            }

            override fun onStartDiscoveryFailed(serviceType: String?, errorCode: Int) {
                close(IllegalStateException("startDiscovery failed: $errorCode"))
            }

            override fun onStopDiscoveryFailed(serviceType: String?, errorCode: Int) {
                Log.w(TAG, "stopDiscovery failed: $errorCode")
            }
        }

        nsdManager.discoverServices(SERVICE_TYPE, NsdManager.PROTOCOL_DNS_SD, discoveryListener)

        awaitClose {
            runCatching { nsdManager.stopServiceDiscovery(discoveryListener) }
        }
    }

    companion object {
        private const val TAG = "ServerDiscovery"
        const val SERVICE_TYPE = "_xuva._tcp."
    }
}
