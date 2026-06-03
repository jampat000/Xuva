package com.xuva.tv.ui.player

import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.viewModelScope
import com.xuva.tv.core.api.XuvaAuthApi
import com.xuva.tv.core.model.PlaybackRoute
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

sealed interface PlayerUiState {
    data object Loading : PlayerUiState
    /** Server returned a playable route. `absoluteUrl` already prefixed
     *  with the paired server's base URL. */
    data class Ready(
        val route: PlaybackRoute,
        val absoluteUrl: String,
        val sessionToken: String,
        val isHls: Boolean,
    ) : PlayerUiState
    data class Error(val message: String) : PlayerUiState
    /** Server returned a non-ready status — e.g. blocked by parental
     *  policy, deferred for transcode, or no playable mode. */
    data class Unplayable(val reason: String) : PlayerUiState
}

class PlayerViewModel(
    private val api: XuvaAuthApi,
    private val mediaSourceId: String,
) : ViewModel() {
    private val _state = MutableStateFlow<PlayerUiState>(PlayerUiState.Loading)
    val state: StateFlow<PlayerUiState> = _state.asStateFlow()

    init { route() }

    private fun route() {
        viewModelScope.launch {
            _state.value = PlayerUiState.Loading
            runCatching { api.getPlaybackRoute(mediaSourceId) }
                .onSuccess { route ->
                    val streamPath = route.streamPath()
                    _state.value = when {
                        !route.isPlayable || streamPath.isNullOrBlank() ->
                            PlayerUiState.Unplayable(reason = route.status ?: route.route)
                        else -> PlayerUiState.Ready(
                            route = route,
                            absoluteUrl = api.baseUrl + streamPath,
                            sessionToken = api.sessionToken,
                            isHls = route.isHls,
                        )
                    }
                }
                .onFailure { _state.value = PlayerUiState.Error(it.message ?: "playback route failed") }
        }
    }

    class Factory(
        private val api: XuvaAuthApi,
        private val mediaSourceId: String,
    ) : ViewModelProvider.Factory {
        @Suppress("UNCHECKED_CAST")
        override fun <T : ViewModel> create(modelClass: Class<T>): T =
            PlayerViewModel(api, mediaSourceId) as T
    }
}
