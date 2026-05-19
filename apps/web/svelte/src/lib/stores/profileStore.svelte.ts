/**
 * Profile store — Svelte 5 runes.
 *
 * Manages which profile is active for the current session.
 * The profile token is persisted in localStorage; the active ProfileCard
 * is kept in reactive memory and refetched when the store initialises.
 */
import { readProfileToken, clearProfileToken } from '$lib/api/profile-token-store';
import type { ProfileCard } from '$lib/api/profiles';

class ProfileStore {
	/** Currently active profile, or null if no profile has been selected. */
	activeProfile = $state<ProfileCard | null>(null);

	/** Whether the Who's Watching screen should be shown. */
	showPicker = $state(false);

	/** True while the initial profile token is being validated on mount. */
	loading = $state(false);

	/** Set the active profile (called after a successful switchProfile). */
	setActiveProfile(profile: ProfileCard | null): void {
		this.activeProfile = profile;
		this.showPicker = false;
	}

	/** Clear profile selection (e.g. on logout). */
	clear(): void {
		clearProfileToken();
		this.activeProfile = null;
		this.showPicker = false;
	}

	/** Show the profile picker screen. */
	openPicker(): void {
		this.showPicker = true;
	}

	/** Dismiss the picker without selecting a profile. */
	closePicker(): void {
		this.showPicker = false;
	}

	/** True when a profile token exists in storage (but card may not yet be loaded). */
	get hasToken(): boolean {
		return !!readProfileToken();
	}
}

export const profileStore = new ProfileStore();
