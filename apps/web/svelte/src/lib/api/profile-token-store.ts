/**
 * Profile token store — mirrors the auth token store pattern.
 *
 * The profile token is a short-lived (24 h) token issued by POST /api/auth/switch-profile.
 * It is stored in localStorage and sent on every API request as X-Profile-Token.
 * It is scoped to the current main auth session and revoked on logout.
 */

const PROFILE_TOKEN_KEY = 'xuva-profile-token';

let memoryProfileToken = '';

function safeRead(storage: Storage | undefined): string {
	try {
		return storage ? String(storage.getItem(PROFILE_TOKEN_KEY) ?? '').trim() : '';
	} catch {
		return '';
	}
}

function safeWrite(storage: Storage | undefined, value: string): void {
	try {
		if (!storage) return;
		if (!value) {
			storage.removeItem(PROFILE_TOKEN_KEY);
		} else {
			storage.setItem(PROFILE_TOKEN_KEY, value);
		}
	} catch {
		// Ignore storage access failures in locked-down browser contexts.
	}
}

export function readProfileToken(): string {
	if (memoryProfileToken) return memoryProfileToken;
	if (typeof window === 'undefined') return '';
	const local = safeRead(window.localStorage);
	if (local) {
		memoryProfileToken = local;
		return local;
	}
	return '';
}

export function writeProfileToken(token: string): void {
	const value = String(token ?? '').trim();
	memoryProfileToken = value;
	if (typeof window === 'undefined') return;
	safeWrite(window.localStorage, value);
}

export function clearProfileToken(): void {
	writeProfileToken('');
}
