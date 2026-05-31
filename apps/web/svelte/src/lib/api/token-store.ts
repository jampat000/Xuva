const AUTH_TOKEN_KEY = 'xuva-auth-token';
// Legacy key — older builds also mirrored the token into window.name, which
// survives same-tab cross-origin navigation and is readable by any script that
// later loads in the tab (an XSS exfiltration vector). We no longer write it and
// actively scrub any leftover value on write so old sessions don't keep leaking.
const LEGACY_WINDOW_NAME_TOKEN_KEY = 'xuvaAuthToken';

let memoryAuthToken = '';

function safeReadStorage(storage: Storage | undefined): string {
	try {
		return storage ? String(storage.getItem(AUTH_TOKEN_KEY) ?? '').trim() : '';
	} catch {
		return '';
	}
}

function safeWriteStorage(storage: Storage | undefined, value: string): void {
	try {
		if (!storage) return;
		if (!value) {
			storage.removeItem(AUTH_TOKEN_KEY);
			return;
		}
		storage.setItem(AUTH_TOKEN_KEY, value);
	} catch {
		// Ignore storage access failures in locked-down browser contexts.
	}
}

// Strip any legacy `xuvaAuthToken=…` entry from window.name. We never write it
// anymore; this only cleans up tokens left behind by older builds so the value
// doesn't persist as the tab is reused.
function scrubLegacyWindowNameToken(): void {
	if (typeof window === 'undefined') return;
	try {
		const raw = String(window.name ?? '');
		if (!raw.includes(`${LEGACY_WINDOW_NAME_TOKEN_KEY}=`)) return;
		const parts = raw
			.split(';')
			.map((item) => item.trim())
			.filter(Boolean)
			.filter((item) => !item.startsWith(`${LEGACY_WINDOW_NAME_TOKEN_KEY}=`));
		window.name = parts.join(';');
	} catch {
		// Ignore window.name access failures.
	}
}

export function readAuthToken(): string {
	if (memoryAuthToken) return memoryAuthToken;
	if (typeof window === 'undefined') return '';

	const localValue = safeReadStorage(window.localStorage);
	if (localValue) {
		memoryAuthToken = localValue;
		return localValue;
	}

	const sessionValue = safeReadStorage(window.sessionStorage);
	if (sessionValue) {
		memoryAuthToken = sessionValue;
		return sessionValue;
	}

	return '';
}

export function writeAuthToken(token: string): void {
	const value = String(token ?? '').trim();
	memoryAuthToken = value;
	if (typeof window === 'undefined') return;
	safeWriteStorage(window.localStorage, value);
	safeWriteStorage(window.sessionStorage, value);
	// Defence-in-depth: ensure no token lingers in window.name from older builds.
	scrubLegacyWindowNameToken();
}

export function clearAuthToken(): void {
	writeAuthToken('');
}

export { AUTH_TOKEN_KEY };
