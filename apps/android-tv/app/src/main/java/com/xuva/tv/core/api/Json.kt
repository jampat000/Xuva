package com.xuva.tv.core.api

import kotlinx.serialization.json.Json

// Shared kotlinx.serialization Json instance. Lenient about unknown server-side
// fields so a server roll-forward doesn't crash the app — the contract is
// "we'll add fields, you'll ignore the ones you don't know about."
val XuvaJson: Json = Json {
    ignoreUnknownKeys = true
    explicitNulls = false
    encodeDefaults = true
}
