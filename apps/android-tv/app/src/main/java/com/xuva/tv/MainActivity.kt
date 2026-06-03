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
 * Top-level navigation gate. Two states until slice 3 introduces browse:
 * either we route through pairing, or we show a paired-placeholder that
 * links to unpair so the user can re-test the flow.
 */
@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun XuvaRoot(app: XuvaApp) {
    var paired by remember { mutableStateOf(app.secureStore.isPaired) }
    if (!paired) {
        PairingFlow(app = app, onPaired = { paired = true })
    } else {
        PairedPlaceholder(
            baseUrl = app.secureStore.baseUrl.orEmpty(),
            onUnpair = {
                app.pairingRepository.unpair()
                paired = false
            },
        )
    }
}

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun PairedPlaceholder(baseUrl: String, onUnpair: () -> Unit) {
    Surface(modifier = Modifier.fillMaxSize()) {
        Box(modifier = Modifier.fillMaxSize().padding(48.dp), contentAlignment = Alignment.Center) {
            Column(horizontalAlignment = Alignment.CenterHorizontally) {
                Text("Paired", style = MaterialTheme.typography.displayMedium)
                Spacer(Modifier.height(12.dp))
                Text("Server: $baseUrl", style = MaterialTheme.typography.bodyLarge)
                Spacer(Modifier.height(8.dp))
                Text(
                    text = "Browse + playback arrive in the next slices.",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.6f),
                )
                Spacer(Modifier.height(40.dp))
                OutlinedButton(onClick = onUnpair) { Text("Unpair this device") }
            }
        }
    }
}
