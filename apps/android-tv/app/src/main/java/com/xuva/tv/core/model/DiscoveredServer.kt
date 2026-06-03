package com.xuva.tv.core.model

data class DiscoveredServer(
    val name: String,
    val host: String,
    val port: Int,
) {
    val baseUrl: String get() = "http://$host:$port"
}
