package com.xuva.tv.ui.detail

import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.viewModelScope
import com.xuva.tv.core.api.XuvaAuthApi
import com.xuva.tv.core.model.MovieDetail
import com.xuva.tv.core.model.SeriesDetail
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

sealed interface DetailUiState {
    data object Loading : DetailUiState
    data class Movie(val detail: MovieDetail) : DetailUiState
    data class Series(val detail: SeriesDetail) : DetailUiState
    data class Error(val message: String) : DetailUiState
}

class DetailViewModel(
    private val api: XuvaAuthApi,
    private val kind: String,
    private val id: String,
) : ViewModel() {
    private val _state = MutableStateFlow<DetailUiState>(DetailUiState.Loading)
    val state: StateFlow<DetailUiState> = _state.asStateFlow()

    init {
        refresh()
    }

    fun refresh() {
        viewModelScope.launch {
            _state.value = DetailUiState.Loading
            runCatching {
                when (kind) {
                    "movie" -> DetailUiState.Movie(api.getMovieDetail(id))
                    "series" -> DetailUiState.Series(api.getSeriesDetail(id))
                    else -> error("unknown kind: $kind")
                }
            }
                .onSuccess { _state.value = it }
                .onFailure { _state.value = DetailUiState.Error(it.message ?: "load failed") }
        }
    }

    class Factory(
        private val api: XuvaAuthApi,
        private val kind: String,
        private val id: String,
    ) : ViewModelProvider.Factory {
        @Suppress("UNCHECKED_CAST")
        override fun <T : ViewModel> create(modelClass: Class<T>): T =
            DetailViewModel(api, kind, id) as T
    }
}
