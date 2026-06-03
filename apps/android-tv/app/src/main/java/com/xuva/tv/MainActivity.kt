package com.xuva.tv

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.tv.material3.ExperimentalTvMaterial3Api
import androidx.tv.material3.MaterialTheme
import androidx.tv.material3.OutlinedButton
import androidx.tv.material3.Surface
import androidx.tv.material3.Text
import com.xuva.tv.ui.browse.BrowseScreen
import com.xuva.tv.ui.pairing.PairingFlow
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
 * Top-level navigation gate. Three states:
 *   - Unpaired → PairingFlow
 *   - Paired   → BrowseScreen (home rows from /api/client/home)
 *   - Detail   → DetailPlaceholder (kind+id, "play" comes in slice 4)
 */
@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun XuvaRoot(app: XuvaApp) {
    var paired by remember { mutableStateOf(app.secureStore.isPaired) }
    var openItem by remember { mutableStateOf<Pair<String, String>?>(null) }

    when {
        !paired -> PairingFlow(app = app, onPaired = { paired = true })
        openItem != null -> {
            val (kind, id) = openItem!!
            DetailPlaceholder(
                kind = kind,
                id = id,
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

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun DetailPlaceholder(kind: String, id: String, onBack: () -> Unit) {
    Surface(modifier = Modifier.fillMaxSize()) {
        Box(modifier = Modifier.fillMaxSize().padding(48.dp), contentAlignment = Alignment.Center) {
            Column(horizontalAlignment = Alignment.CenterHorizontally) {
                Text("$kind / $id", style = MaterialTheme.typography.displaySmall)
                Spacer(Modifier.height(12.dp))
                Text(
                    text = "Detail screen + ExoPlayer arrive in slice 4.",
                    style = MaterialTheme.typography.bodyLarge,
                    color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.6f),
                )
                Spacer(Modifier.height(40.dp))
                OutlinedButton(onClick = onBack) { Text("Back") }
            }
        }
    }
}
