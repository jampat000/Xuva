/**
 * App-level shared reactive state (Svelte 5 runes).
 * Import `appState` anywhere — mutations are reflected everywhere.
 *
 * Usage:
 *   import { appState } from '$lib/stores/appState.svelte';
 *   appState.serverName          // read
 *   appState.serverName = 'Foo' // write (triggers reactivity)
 */
class AppState {
  serverName = $state('Xuva');
}

export const appState = new AppState();
