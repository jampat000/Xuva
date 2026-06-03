package com.xuva.tv.ui.browse

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel
import androidx.tv.material3.Button
import androidx.tv.material3.ExperimentalTvMaterial3Api
import androidx.tv.material3.MaterialTheme
import androidx.tv.material3.Surface
import androidx.tv.material3.Text
import coil.compose.AsyncImage
import com.xuva.tv.XuvaApp
import com.xuva.tv.core.model.HomeItem
import com.xuva.tv.core.model.HomeRow

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
fun BrowseScreen(
    app: XuvaApp,
    onOpenItem: (kind: String, id: String) -> Unit,
    onUnpair: () -> Unit,
) {
    val viewModel: BrowseViewModel = viewModel(
        factory = BrowseViewModel.Factory(api = app.authApi)
    )
    val state by viewModel.state.collectAsState()

    Surface(modifier = Modifier.fillMaxSize()) {
        when (val ui = state) {
            BrowseUiState.Loading -> CenteredText("Loading…")
            is BrowseUiState.Error -> {
                Column(
                    modifier = Modifier.fillMaxSize().padding(48.dp),
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.Center,
                ) {
                    Text("Can't reach your server", style = MaterialTheme.typography.headlineSmall)
                    Spacer(Modifier.height(8.dp))
                    Text(
                        text = ui.message,
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.error,
                    )
                    Spacer(Modifier.height(24.dp))
                    Button(onClick = viewModel::refresh) { Text("Retry") }
                    Spacer(Modifier.height(16.dp))
                    Button(onClick = onUnpair) { Text("Unpair") }
                }
            }
            is BrowseUiState.Ready -> Browse(payload = ui.payload, onOpenItem = onOpenItem)
        }
    }
}

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun Browse(
    payload: com.xuva.tv.core.model.HomePayload,
    onOpenItem: (kind: String, id: String) -> Unit,
) {
    LazyColumn(
        modifier = Modifier.fillMaxSize(),
        contentPadding = PaddingValues(vertical = 48.dp),
        verticalArrangement = Arrangement.spacedBy(40.dp),
    ) {
        items(
            items = payload.rows.filter { it.items.isNotEmpty() },
            key = { row -> row.id },
        ) { row -> Row(row = row, onOpenItem = onOpenItem) }
    }
}

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun Row(row: HomeRow, onOpenItem: (kind: String, id: String) -> Unit) {
    Column(modifier = Modifier.fillMaxWidth()) {
        row.eyebrow?.let { eyebrow ->
            Text(
                text = eyebrow.uppercase(),
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.6f),
                modifier = Modifier.padding(start = 48.dp),
            )
        }
        Text(
            text = row.title,
            style = MaterialTheme.typography.headlineSmall,
            modifier = Modifier.padding(start = 48.dp),
        )
        Spacer(Modifier.height(12.dp))
        LazyRow(
            contentPadding = PaddingValues(horizontal = 48.dp),
            horizontalArrangement = Arrangement.spacedBy(16.dp),
        ) {
            items(items = row.items, key = { "${it.kind}:${it.id}" }) { item ->
                Card(item = item, onClick = { onOpenItem(item.kind, item.id) })
            }
        }
    }
}

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun Card(item: HomeItem, onClick: () -> Unit) {
    Button(
        onClick = onClick,
        modifier = Modifier.width(180.dp),
    ) {
        Column {
            Box(
                modifier = Modifier
                    .width(160.dp)
                    .aspectRatio(2f / 3f),
            ) {
                if (!item.posterUrl.isNullOrBlank()) {
                    AsyncImage(
                        model = item.posterUrl,
                        contentDescription = item.title,
                        contentScale = ContentScale.Crop,
                        modifier = Modifier.fillMaxSize(),
                    )
                } else {
                    Surface(modifier = Modifier.fillMaxSize()) {
                        Box(contentAlignment = Alignment.Center) {
                            Text(item.title.take(2).uppercase(), style = MaterialTheme.typography.headlineMedium)
                        }
                    }
                }
            }
            Spacer(Modifier.height(8.dp))
            Text(
                text = item.title,
                style = MaterialTheme.typography.bodyMedium,
                maxLines = 1,
            )
            item.subtitle?.let {
                Text(
                    text = it,
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.6f),
                    maxLines = 1,
                )
            }
        }
    }
}

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun CenteredText(text: String) {
    Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
        Text(text, style = MaterialTheme.typography.headlineSmall)
    }
}
