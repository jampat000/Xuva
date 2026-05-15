const DESKTOP_MIN_WIDTH = 980;

function storageAvailable(): boolean {
	return typeof window !== 'undefined' && typeof window.localStorage !== 'undefined';
}

function normalizeBoolean(value: string | null): boolean {
	return value === '1';
}

export function isDesktopSidebarViewport(): boolean {
	return typeof window !== 'undefined' && window.innerWidth >= DESKTOP_MIN_WIDTH;
}

export function readSidebarPinnedPreference(storageKey: string): boolean {
	if (!storageAvailable()) return false;
	try {
		return normalizeBoolean(window.localStorage.getItem(storageKey));
	} catch {
		return false;
	}
}

export function writeSidebarPinnedPreference(storageKey: string, pinned: boolean): void {
	if (!storageAvailable()) return;
	try {
		window.localStorage.setItem(storageKey, pinned ? '1' : '0');
	} catch {
		// Ignore storage write failures and keep the UI responsive.
	}
}
