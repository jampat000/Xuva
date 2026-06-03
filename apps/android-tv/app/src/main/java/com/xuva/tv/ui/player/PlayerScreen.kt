package com.xuva.tv.ui.player

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import androidx.compose.ui.viewinterop.AndroidView
import androidx.lifecycle.viewmodel.compose.viewModel
import androidx.media3.common.MediaItem
import androidx.media3.datasource.DefaultHttpDataSource
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.exoplayer.hls.HlsMediaSource
import androidx.media3.exoplayer.source.ProgressiveMediaSource
import androidx.media3.ui.PlayerView
import androidx.tv.material3.Button
import androidx.tv.material3.ExperimentalTvMaterial3Api
import androidx.tv.material3.MaterialTheme
import androidx.tv.material3.Text
import com.xuva.tv.XuvaApp

/**
 * Fullscreen ExoPlayer host. Fetches /api/playback/route → builds an
 * HlsMediaSource or ProgressiveMediaSource with a DefaultHttpDataSource
 * factory pre-configured with `X-Auth-Token: <session>` so every segment
 * fetch carries the auth header → plays.
 *
 * Slice 4 lands the core watch experience. Resume position (via
 * /api/playback/state) and progress beacons (/api/playback/progress)
 * are the obvious next polish; they slot into [DisposableEffect].
 */
@OptIn(ExperimentalTvMaterial3Api::class, androidx.media3.common.util.UnstableApi::class)
@Composable
fun PlayerScreen(
    app: XuvaApp,
    mediaSourceId: String,
    onExit: () -> Unit,
) {
    val viewModel: PlayerViewModel = viewModel(
        key = "player:$mediaSourceId",
        factory = PlayerViewModel.Factory(app.authApi, mediaSourceId),
    )
    val state by viewModel.state.collectAsState()

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black),
    ) {
        when (val ui = state) {
            PlayerUiState.Loading -> Centered("Preparing stream…")
            is PlayerUiState.Error -> ErrorBox(message = ui.message, onExit = onExit)
            is PlayerUiState.Unplayable -> ErrorBox(message = "Can't play this title (${ui.reason})", onExit = onExit)
            is PlayerUiState.Ready -> ExoHost(
                url = ui.absoluteUrl,
                sessionToken = ui.sessionToken,
                isHls = ui.isHls,
                onExit = onExit,
            )
        }
    }
}

@OptIn(ExperimentalTvMaterial3Api::class, androidx.media3.common.util.UnstableApi::class)
@Composable
private fun ExoHost(url: String, sessionToken: String, isHls: Boolean, onExit: () -> Unit) {
    val context = LocalContext.current

    // Player lives across recompositions; release in DisposableEffect so a
    // back-press teardown doesn't leak the codec/audio session.
    val player = remember(url) {
        val httpFactory = DefaultHttpDataSource.Factory()
            .setUserAgent("Xuva-TV/0.1.0 (Android)")
            .setDefaultRequestProperties(mapOf("X-Auth-Token" to sessionToken))
        val source = if (isHls) {
            HlsMediaSource.Factory(httpFactory).createMediaSource(MediaItem.fromUri(url))
        } else {
            ProgressiveMediaSource.Factory(httpFactory).createMediaSource(MediaItem.fromUri(url))
        }
        ExoPlayer.Builder(context).build().apply {
            setMediaSource(source)
            playWhenReady = true
            prepare()
        }
    }

    DisposableEffect(player) {
        onDispose { player.release() }
    }

    Box(modifier = Modifier.fillMaxSize()) {
        AndroidView(
            modifier = Modifier.fillMaxSize(),
            factory = { ctx ->
                PlayerView(ctx).apply {
                    this.player = player
                    useController = true
                    // TV-leanback affordances: hide the controller after 4s of
                    // inactivity (matches Apple TV scrub-bar behaviour).
                    controllerShowTimeoutMs = 4000
                }
            },
            update = { view -> view.player = player },
        )
    }
}

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun ErrorBox(message: String, onExit: () -> Unit) {
    Box(modifier = Modifier.fillMaxSize().padding(48.dp), contentAlignment = Alignment.Center) {
        Column(horizontalAlignment = Alignment.CenterHorizontally) {
            Text("Playback error", style = MaterialTheme.typography.headlineSmall)
            Spacer(Modifier.height(8.dp))
            Text(message, style = MaterialTheme.typography.bodyMedium, color = MaterialTheme.colorScheme.error)
            Spacer(Modifier.height(24.dp))
            Button(onClick = onExit) { Text("Back") }
        }
    }
}

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun Centered(text: String) {
    Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
        Text(text, style = MaterialTheme.typography.headlineSmall, color = MaterialTheme.colorScheme.onBackground)
    }
}
