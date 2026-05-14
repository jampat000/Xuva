const AUTH_TOKEN_KEY = 'xuva-auth-token';
const WINDOW_NAME_TOKEN_KEY = 'xuvaAuthToken';

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

function safeReadWindowNameToken(): string {
	if (typeof window === 'undefined') return '';
	try {
		const raw = String(window.name ?? '').trim();
		if (!raw) return '';
		const parts = raw.split(';').map((value) => value.trim());
		const pair = parts.find((value) => value.startsWith(`${WINDOW_NAME_TOKEN_KEY}=`));
		if (!pair) return '';
		return decodeURIComponent(pair.slice(`${WINDOW_NAME_TOKEN_KEY}=`.length)).trim();
	} catch {
		return '';
	}
}

function safeWriteWindowNameToken(value: string): void {
	if (typeof window === 'undefined') return;
	try {
		const parts = String(window.name ?? '')
			.split(';')
			.map((item) => item.trim())
			.filter(Boolean)
			.filter((item) => !item.startsWith(`${WINDOW_NAME_TOKEN_KEY}=`));
		if (value) {
			parts.push(`${WINDOW_NAME_TOKEN_KEY}=${encodeURIComponent(value)}`);
		}
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

	const windowValue = safeReadWindowNameToken();
	if (windowValue) {
		memoryAuthToken = windowValue;
		return windowValue;
	}

	return '';
}

export function writeAuthToken(token: string): void {
	const value = String(token ?? '').trim();
	memoryAuthToken = value;
	if (typeof window === 'undefined') return;
	safeWriteStorage(window.localStorage, value);
	safeWriteStorage(window.sessionStorage, value);
	safeWriteWindowNameToken(value);
}

export function clearAuthToken(): void {
	writeAuthToken('');
}

export { AUTH_TOKEN_KEY };
