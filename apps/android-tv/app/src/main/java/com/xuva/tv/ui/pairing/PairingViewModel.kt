package com.xuva.tv.ui.pairing

import android.os.Build
import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.viewModelScope
import com.xuva.tv.core.discovery.ServerDiscovery
import com.xuva.tv.core.model.DiscoveredServer
import com.xuva.tv.core.model.PairingItem
import com.xuva.tv.core.pairing.PairingRepository
import com.xuva.tv.core.storage.SecureStore
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

sealed interface PairingUiState {
    /** No paired server yet — show server list (NSD discoveries + manual entry). */
    data class ServerSelection(
        val discovered: List<DiscoveredServer> = emptyList(),
        val manualUrl: String = "",
        val error: String? = null,
    ) : PairingUiState

    /** Pairing request submitted — show the code and poll for approval. */
    data class WaitingForApproval(
        val item: PairingItem,
        val baseUrl: String,
    ) : PairingUiState

    /** Pairing rejected, expired, or cancelled — back to selection w/ a reason. */
    data class Rejected(val reason: String) : PairingUiState

    /** Approved — auth grant is saved, the host activity should swap to the
     *  browse experience. Slice 3 wires that up. */
    data object Paired : PairingUiState
}

class PairingViewModel(
    private val repository: PairingRepository,
    private val discovery: ServerDiscovery,
    private val secureStore: SecureStore,
) : ViewModel() {

    private val _state = MutableStateFlow<PairingUiState>(
        if (secureStore.isPaired) PairingUiState.Paired else PairingUiState.ServerSelection()
    )
    val state: StateFlow<PairingUiState> = _state.asStateFlow()

    private var discoveryJob: Job? = null
    private var pollJob: Job? = null

    init {
        if (_state.value is PairingUiState.ServerSelection) startDiscovery()
    }

    private fun startDiscovery() {
        discoveryJob?.cancel()
        discoveryJob = viewModelScope.launch {
            runCatching {
                discovery.discover().collect { server ->
                    _state.update { current ->
                        when (current) {
                            is PairingUiState.ServerSelection -> current.copy(
                                discovered = (current.discovered + server)
                                    .distinctBy { "${it.host}:${it.port}" },
                            )
                            else -> current
                        }
                    }
                }
            }
        }
    }

    fun setManualUrl(value: String) {
        _state.update { current ->
            if (current is PairingUiState.ServerSelection) current.copy(manualUrl = value, error = null)
            else current
        }
    }

    fun pair(server: DiscoveredServer) = pair(server.baseUrl)

    fun pairManualEntry() {
        val current = _state.value as? PairingUiState.ServerSelection ?: return
        val url = normaliseUrl(current.manualUrl)
        if (url.isNullOrBlank()) {
            _state.value = current.copy(error = "Enter a server URL like http://192.168.1.10:8097")
            return
        }
        pair(url)
    }

    private fun pair(baseUrl: String) {
        discoveryJob?.cancel()
        pollJob?.cancel()
        pollJob = viewModelScope.launch {
            runCatching {
                val item = repository.start(baseUrl, deviceName = defaultDeviceName())
                _state.value = PairingUiState.WaitingForApproval(item = item, baseUrl = baseUrl)
                val final = repository.await(baseUrl, item.id)
                _state.value = when (final.status) {
                    "approved" -> PairingUiState.Paired
                    else -> PairingUiState.Rejected(reason = final.status)
                }
            }.onFailure { e ->
                if (e is CancellationException) throw e
                _state.value = PairingUiState.Rejected(reason = e.message ?: "network error")
            }
        }
    }

    fun cancel() {
        val waiting = _state.value as? PairingUiState.WaitingForApproval ?: return
        viewModelScope.launch { repository.cancel(waiting.baseUrl, waiting.item.id) }
        pollJob?.cancel()
        _state.value = PairingUiState.ServerSelection()
        startDiscovery()
    }

    fun resetAfterRejection() {
        _state.value = PairingUiState.ServerSelection()
        startDiscovery()
    }

    fun unpair() {
        repository.unpair()
        _state.value = PairingUiState.ServerSelection()
        startDiscovery()
    }

    override fun onCleared() {
        discoveryJob?.cancel()
        pollJob?.cancel()
    }

    private fun defaultDeviceName(): String =
        "${Build.MODEL ?: "Android TV"} (${Build.MANUFACTURER ?: "unknown"})".trim()

    private fun normaliseUrl(raw: String): String? {
        val trimmed = raw.trim().ifBlank { return null }
        val withScheme = if ("://" in trimmed) trimmed else "http://$trimmed"
        return withScheme.trimEnd('/')
    }

    class Factory(
        private val repository: PairingRepository,
        private val discovery: ServerDiscovery,
        private val secureStore: SecureStore,
    ) : ViewModelProvider.Factory {
        @Suppress("UNCHECKED_CAST")
        override fun <T : ViewModel> create(modelClass: Class<T>): T =
            PairingViewModel(repository, discovery, secureStore) as T
    }
}
