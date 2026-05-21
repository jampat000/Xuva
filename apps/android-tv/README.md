# Xuva Android TV

Native Android TV client.

Android TV follows after the Apple TV alpha proves the server-client playback contract. The Android app should reuse the same pairing, catalog, capability, playback decision, session heartbeat, and HLS/direct-play contracts rather than creating a separate product path.

Responsibilities:

- Browse libraries.
- Play media (ExoPlayer for direct + HLS, equivalent of AVPlayerViewController).
- Report client playback capabilities via clientProfile=android-tv.
- Support local pairing (POST /api/pairing/requests; poll GET /api/pairing/requests/{id}; DELETE the id when the user cancels).
- Support manual server URL entry.
- Expose playback quality and subtitle controls.

## Server discovery (mDNS / Bonjour)

The Xuva server advertises itself on the LAN via mDNS. Implement discovery first so users don't have to type IP addresses.

**Wire format** (already in use by the Apple TV client):

| Field | Value |
| --- | --- |
| Service type | `_xuva._tcp` |
| Domain | `local.` |
| Port | server HTTP port (usually 8097) |
| TXT records | `app=xuva`, `serverName=<display name>`, `hostName=<host>`, `web=<http://host:port>`, `api=/api/client/bootstrap` |

**Android implementation:**

```kotlin
val nsd = getSystemService(Context.NSD_SERVICE) as NsdManager
val listener = object : NsdManager.DiscoveryListener {
    override fun onServiceFound(service: NsdServiceInfo) {
        nsd.resolveService(service, resolveListener)
    }
    override fun onStartDiscoveryFailed(serviceType: String, errorCode: Int) {}
    override fun onStopDiscoveryFailed(serviceType: String, errorCode: Int) {}
    override fun onDiscoveryStarted(serviceType: String) {}
    override fun onDiscoveryStopped(serviceType: String) {}
    override fun onServiceLost(service: NsdServiceInfo) {}
}
nsd.discoverServices("_xuva._tcp.", NsdManager.PROTOCOL_DNS_SD, listener)
```

The resolveListener gets a fully resolved NsdServiceInfo with host + port + TXT records. Build the base URL as `http://${host}:${port}` and present a "Servers on your network" list. Tapping one runs the pairing flow as documented above.

If discovery yields nothing within ~4 seconds, fall back to a manual URL field (see the Apple TV's `PairingScreen.swift` `discoveryTimedOut` behavior).

Manifest needs `<uses-permission android:name="android.permission.INTERNET" />` and `<uses-permission android:name="android.permission.CHANGE_WIFI_MULTICAST_STATE" />` for multicast on some devices.
