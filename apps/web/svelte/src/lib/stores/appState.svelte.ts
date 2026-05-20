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
  /** Whether the hero should autoplay trailers. Defaults to true; set from
   *  the bootstrap features.trailers flag once the server responds. */
  trailersEnabled = $state(true);
}

export const appState = new AppState();
