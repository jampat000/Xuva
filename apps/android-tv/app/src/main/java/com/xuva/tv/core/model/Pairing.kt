package com.xuva.tv.core.model

import kotlinx.serialization.Serializable

@Serializable
data class CreateRequest(
    val deviceName: String,
    val clientProfile: String,
    val deviceId: String,
)

@Serializable
data class AuthGrant(
    val method: String,
    val sessionToken: String,
    val expiresAt: String? = null,
)

@Serializable
data class PairingItem(
    val id: String,
    val code: String? = null,
    val deviceName: String,
    val clientProfile: String,
    val deviceId: String? = null,
    val auth: AuthGrant? = null,
    val status: String,
    val approvedBy: String? = null,
    val expiresAt: String? = null,
    val createdAt: String? = null,
    val updatedAt: String? = null,
)
