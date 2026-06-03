package com.xuva.tv.ui.browse

import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.viewModelScope
import com.xuva.tv.core.api.XuvaAuthApi
import com.xuva.tv.core.model.HomePayload
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

sealed interface BrowseUiState {
    data object Loading : BrowseUiState
    data class Ready(val payload: HomePayload) : BrowseUiState
    data class Error(val message: String) : BrowseUiState
}

class BrowseViewModel(private val api: XuvaAuthApi) : ViewModel() {
    private val _state = MutableStateFlow<BrowseUiState>(BrowseUiState.Loading)
    val state: StateFlow<BrowseUiState> = _state.asStateFlow()

    init {
        refresh()
    }

    fun refresh() {
        viewModelScope.launch {
            _state.value = BrowseUiState.Loading
            runCatching { api.getHome() }
                .onSuccess { _state.value = BrowseUiState.Ready(it) }
                .onFailure { _state.value = BrowseUiState.Error(it.message ?: "load failed") }
        }
    }

    class Factory(private val api: XuvaAuthApi) : ViewModelProvider.Factory {
        @Suppress("UNCHECKED_CAST")
        override fun <T : ViewModel> create(modelClass: Class<T>): T =
            BrowseViewModel(api) as T
    }
}
