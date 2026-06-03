package com.xuva.tv.ui.detail

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
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
import androidx.tv.material3.OutlinedButton
import androidx.tv.material3.Surface
import androidx.tv.material3.Text
import coil.compose.AsyncImage
import com.xuva.tv.XuvaApp
import com.xuva.tv.core.model.EpisodeBrief
import com.xuva.tv.core.model.MovieDetail
import com.xuva.tv.core.model.SeasonDetail
import com.xuva.tv.core.model.SeriesDetail

/**
 * Detail screen for a movie or series. Hero backdrop + metadata + Play button
 * for movies; series additionally renders a season-then-episode list with a
 * per-episode Play button. Tapping Play hands a `mediaSourceId` back to the
 * activity, which swaps to the player composable.
 */
@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
fun DetailScreen(
    app: XuvaApp,
    kind: String,
    id: String,
    onPlay: (mediaSourceId: String) -> Unit,
    onBack: () -> Unit,
) {
    val viewModel: DetailViewModel = viewModel(
        key = "$kind:$id",
        factory = DetailViewModel.Factory(app.authApi, kind, id),
    )
    val state by viewModel.state.collectAsState()

    Surface(modifier = Modifier.fillMaxSize()) {
        when (val ui = state) {
            DetailUiState.Loading -> Centered("Loading…")
            is DetailUiState.Error -> ErrorState(message = ui.message, onRetry = viewModel::refresh, onBack = onBack)
            is DetailUiState.Movie -> MovieDetailContent(detail = ui.detail, onPlay = onPlay, onBack = onBack)
            is DetailUiState.Series -> SeriesDetailContent(detail = ui.detail, onPlay = onPlay, onBack = onBack)
        }
    }
}

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun MovieDetailContent(detail: MovieDetail, onPlay: (String) -> Unit, onBack: () -> Unit) {
    val backdrop = detail.metadata?.backdropUrl
    val overview = detail.metadata?.overview.orEmpty()
    val playable = detail.playableMediaSourceId()

    Box(modifier = Modifier.fillMaxSize()) {
        if (!backdrop.isNullOrBlank()) {
            AsyncImage(
                model = backdrop,
                contentDescription = null,
                contentScale = ContentScale.Crop,
                modifier = Modifier.fillMaxSize(),
            )
        }
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(start = 80.dp, top = 80.dp, end = 80.dp, bottom = 60.dp),
            verticalArrangement = Arrangement.Bottom,
        ) {
            Text(detail.title, style = MaterialTheme.typography.displayMedium)
            if (detail.year > 0) {
                Text("${detail.year}", style = MaterialTheme.typography.titleMedium)
            }
            if (overview.isNotBlank()) {
                Spacer(Modifier.height(16.dp))
                Text(
                    text = overview,
                    style = MaterialTheme.typography.bodyLarge,
                    modifier = Modifier.fillMaxWidth(0.7f),
                    maxLines = 5,
                )
            }
            Spacer(Modifier.height(24.dp))
            Row(horizontalArrangement = Arrangement.spacedBy(16.dp)) {
                Button(
                    onClick = { playable?.let(onPlay) },
                    enabled = playable != null,
                ) { Text(if (playable != null) "Play" else "No playable version") }
                OutlinedButton(onClick = onBack) { Text("Back") }
            }
        }
    }
}

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun SeriesDetailContent(detail: SeriesDetail, onPlay: (String) -> Unit, onBack: () -> Unit) {
    val backdrop = detail.metadata?.backdropUrl
    val overview = detail.metadata?.overview.orEmpty()
    LazyColumn(
        modifier = Modifier.fillMaxSize(),
        contentPadding = PaddingValues(start = 80.dp, top = 60.dp, end = 80.dp, bottom = 80.dp),
        verticalArrangement = Arrangement.spacedBy(28.dp),
    ) {
        item {
            Column {
                if (!backdrop.isNullOrBlank()) {
                    AsyncImage(
                        model = backdrop,
                        contentDescription = null,
                        contentScale = ContentScale.Crop,
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(280.dp),
                    )
                    Spacer(Modifier.height(16.dp))
                }
                Text(detail.title, style = MaterialTheme.typography.displayMedium)
                if (overview.isNotBlank()) {
                    Spacer(Modifier.height(8.dp))
                    Text(
                        text = overview,
                        style = MaterialTheme.typography.bodyLarge,
                        modifier = Modifier.fillMaxWidth(0.7f),
                        maxLines = 4,
                    )
                }
                Spacer(Modifier.height(8.dp))
                OutlinedButton(onClick = onBack) { Text("Back") }
            }
        }
        items(detail.seasons, key = { it.id }) { season ->
            SeasonBlock(season = season, onPlay = onPlay)
        }
    }
}

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun SeasonBlock(season: SeasonDetail, onPlay: (String) -> Unit) {
    Column {
        Text(
            text = "Season ${season.seasonNumber}",
            style = MaterialTheme.typography.headlineSmall,
        )
        Spacer(Modifier.height(12.dp))
        Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
            season.episodes.forEach { episode -> EpisodeRow(episode = episode, onPlay = onPlay) }
        }
    }
}

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun EpisodeRow(episode: EpisodeBrief, onPlay: (String) -> Unit) {
    val playable = episode.playableMediaSourceId()
    val title = episode.title?.takeIf { it.isNotBlank() } ?: "Episode ${episode.episodeNumber}"
    Row(
        modifier = Modifier.fillMaxWidth(),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Column(modifier = Modifier.weight(1f)) {
            Text(
                text = "${episode.episodeNumber}. $title",
                style = MaterialTheme.typography.titleMedium,
            )
            episode.metadata?.overview?.takeIf { it.isNotBlank() }?.let { synopsis ->
                Text(
                    text = synopsis,
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.7f),
                    maxLines = 2,
                )
            }
        }
        Spacer(Modifier.width(16.dp))
        Button(
            onClick = { playable?.let(onPlay) },
            enabled = playable != null,
        ) {
            Text(if (playable != null) "Play" else "No file")
        }
    }
}

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun ErrorState(message: String, onRetry: () -> Unit, onBack: () -> Unit) {
    Box(modifier = Modifier.fillMaxSize().padding(48.dp), contentAlignment = Alignment.Center) {
        Column(horizontalAlignment = Alignment.CenterHorizontally) {
            Text("Couldn't load", style = MaterialTheme.typography.headlineSmall)
            Spacer(Modifier.height(8.dp))
            Text(message, style = MaterialTheme.typography.bodyMedium, color = MaterialTheme.colorScheme.error)
            Spacer(Modifier.height(24.dp))
            Row(horizontalArrangement = Arrangement.spacedBy(16.dp)) {
                Button(onClick = onRetry) { Text("Retry") }
                OutlinedButton(onClick = onBack) { Text("Back") }
            }
        }
    }
}

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun Centered(text: String) {
    Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
        Text(text, style = MaterialTheme.typography.headlineSmall)
    }
}
