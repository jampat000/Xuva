package com.xuva.tv

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.BackHandler
import androidx.activity.compose.setContent
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import com.xuva.tv.ui.browse.BrowseScreen
import com.xuva.tv.ui.detail.DetailScreen
import com.xuva.tv.ui.pairing.PairingFlow
import com.xuva.tv.ui.player.PlayerScreen
import com.xuva.tv.ui.theme.XuvaTVTheme

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        val app = applicationContext as XuvaApp
        setContent {
            XuvaTVTheme {
                XuvaRoot(app)
            }
        }
    }
}

/**
 * Top-level navigation gate — four states (sealed inputs, manual stack):
 *   - Unpaired      → PairingFlow
 *   - Paired/Browse → BrowseScreen (home rows from /api/client/home)
 *   - Detail        → DetailScreen for a (kind, id)
 *   - Playing       → PlayerScreen for a mediaSourceId
 *
 * Navigation is just `remember { mutableStateOf(...) }` — a real Navigation
 * Compose graph is overkill for a 4-state machine. The dependency is wired
 * for slice 5 when episode-up-next and a back-stack of detail screens land.
 */
@Composable
private fun XuvaRoot(app: XuvaApp) {
    var paired by remember { mutableStateOf(app.secureStore.isPaired) }
    var openItem by remember { mutableStateOf<Pair<String, String>?>(null) }
    var playingMediaSourceId by remember { mutableStateOf<String?>(null) }

    // Hardware Back: drop one screen off the navigation stack at a time
    // (Playing → Detail, Detail → Browse) instead of nuking the activity.
    BackHandler(enabled = playingMediaSourceId != null) { playingMediaSourceId = null }
    BackHandler(enabled = playingMediaSourceId == null && openItem != null) { openItem = null }

    when {
        !paired -> PairingFlow(app = app, onPaired = { paired = true })
        playingMediaSourceId != null -> PlayerScreen(
            app = app,
            mediaSourceId = playingMediaSourceId!!,
            onExit = { playingMediaSourceId = null },
        )
        openItem != null -> {
            val (kind, id) = openItem!!
            DetailScreen(
                app = app,
                kind = kind,
                id = id,
                onPlay = { mediaSourceId -> playingMediaSourceId = mediaSourceId },
                onBack = { openItem = null },
            )
        }
        else -> BrowseScreen(
            app = app,
            onOpenItem = { kind, id -> openItem = kind to id },
            onUnpair = {
                app.pairingRepository.unpair()
                paired = false
            },
        )
    }
}
