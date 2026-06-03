package com.xuva.tv.ui.pairing

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel
import androidx.tv.material3.Button
import androidx.tv.material3.ExperimentalTvMaterial3Api
import androidx.tv.material3.MaterialTheme
import androidx.tv.material3.OutlinedButton
import androidx.tv.material3.Surface
import androidx.tv.material3.Text
import com.xuva.tv.XuvaApp
import com.xuva.tv.core.model.DiscoveredServer

/**
 * Top-level pairing flow composable. Picks the screen to render based on
 * `PairingUiState`. The host activity passes the [XuvaApp] container in so
 * the ViewModel factory can wire the singletons without DI machinery.
 */
@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
fun PairingFlow(app: XuvaApp, onPaired: () -> Unit) {
    val viewModel: PairingViewModel = viewModel(
        factory = PairingViewModel.Factory(
            repository = app.pairingRepository,
            discovery = app.serverDiscovery,
            secureStore = app.secureStore,
        )
    )
    val state by viewModel.state.collectAsState()

    Surface(modifier = Modifier.fillMaxSize()) {
        Box(
            modifier = Modifier.fillMaxSize().padding(48.dp),
            contentAlignment = Alignment.Center,
        ) {
            when (val ui = state) {
                is PairingUiState.ServerSelection -> ServerSelection(
                    state = ui,
                    onManualUrlChange = viewModel::setManualUrl,
                    onPair = viewModel::pair,
                    onPairManual = viewModel::pairManualEntry,
                )
                is PairingUiState.WaitingForApproval -> WaitingForApproval(
                    item = ui.item,
                    onCancel = viewModel::cancel,
                )
                is PairingUiState.Rejected -> Rejected(
                    reason = ui.reason,
                    onReset = viewModel::resetAfterRejection,
                )
                is PairingUiState.Paired -> {
                    onPaired()
                }
            }
        }
    }
}

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun ServerSelection(
    state: PairingUiState.ServerSelection,
    onManualUrlChange: (String) -> Unit,
    onPair: (DiscoveredServer) -> Unit,
    onPairManual: () -> Unit,
) {
    Column(
        modifier = Modifier.fillMaxWidth(0.7f),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text("Pair your Xuva server", style = MaterialTheme.typography.headlineMedium)
        Spacer(Modifier.height(12.dp))
        Text(
            text = "Choose a discovered server, or enter the URL manually.",
            style = MaterialTheme.typography.bodyLarge,
            textAlign = TextAlign.Center,
            color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.7f),
        )
        Spacer(Modifier.height(32.dp))

        if (state.discovered.isEmpty()) {
            Text(
                text = "Searching the local network…",
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.5f),
            )
        } else {
            Text(
                text = "Discovered servers",
                style = MaterialTheme.typography.labelLarge,
                color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.7f),
            )
            Spacer(Modifier.height(8.dp))
            LazyColumn(
                modifier = Modifier.fillMaxWidth(),
                contentPadding = PaddingValues(vertical = 4.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                items(state.discovered, key = { "${it.host}:${it.port}" }) { server ->
                    Button(
                        onClick = { onPair(server) },
                        modifier = Modifier.fillMaxWidth(),
                    ) {
                        Column(Modifier.padding(8.dp)) {
                            Text(server.name, style = MaterialTheme.typography.titleMedium)
                            Text(server.baseUrl, style = MaterialTheme.typography.bodySmall)
                        }
                    }
                }
            }
        }

        Spacer(Modifier.height(32.dp))
        Text(
            text = "Or enter a server URL",
            style = MaterialTheme.typography.labelLarge,
            color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.7f),
        )
        Spacer(Modifier.height(8.dp))
        BasicTextField(
            value = state.manualUrl,
            onValueChange = onManualUrlChange,
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 24.dp),
            singleLine = true,
            cursorBrush = SolidColor(MaterialTheme.colorScheme.primary),
        )
        Spacer(Modifier.height(12.dp))
        OutlinedButton(onClick = onPairManual) {
            Text("Pair with this server")
        }
        state.error?.let { err ->
            Spacer(Modifier.height(12.dp))
            Text(err, color = MaterialTheme.colorScheme.error, style = MaterialTheme.typography.bodyMedium)
        }
    }
}

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun WaitingForApproval(item: com.xuva.tv.core.model.PairingItem, onCancel: () -> Unit) {
    Column(horizontalAlignment = Alignment.CenterHorizontally) {
        Text("Approve this device on the server", style = MaterialTheme.typography.headlineSmall)
        Spacer(Modifier.height(16.dp))
        Text(
            text = "Open the Xuva web admin and approve this device.",
            style = MaterialTheme.typography.bodyLarge,
            textAlign = TextAlign.Center,
            color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.7f),
        )
        Spacer(Modifier.height(40.dp))
        item.code?.let { code ->
            Text(
                text = code,
                style = MaterialTheme.typography.displayLarge,
                color = MaterialTheme.colorScheme.primary,
            )
            Spacer(Modifier.height(8.dp))
            Text(
                text = "Code valid for a few minutes.",
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.5f),
            )
        } ?: Text(
            text = "Waiting for the server to issue a code…",
            style = MaterialTheme.typography.bodyMedium,
        )
        Spacer(Modifier.height(40.dp))
        OutlinedButton(onClick = onCancel) {
            Text("Cancel and choose a different server")
        }
    }
}

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun Rejected(reason: String, onReset: () -> Unit) {
    Column(horizontalAlignment = Alignment.CenterHorizontally) {
        Text("Pairing didn't go through", style = MaterialTheme.typography.headlineSmall)
        Spacer(Modifier.height(16.dp))
        Text(
            text = "Reason: $reason",
            style = MaterialTheme.typography.bodyLarge,
            color = MaterialTheme.colorScheme.error,
        )
        Spacer(Modifier.height(32.dp))
        Button(onClick = onReset) {
            Text("Try again")
        }
    }
}
