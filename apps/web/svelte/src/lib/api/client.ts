import { clearAuthToken, readAuthToken, writeAuthToken } from './token-store';
import { readProfileToken } from './profile-token-store';

const SAFE_METHODS = new Set(['GET', 'HEAD', 'OPTIONS']);
const DEFAULT_TIMEOUT_MS = 15_000;
const DEFAULT_RETRIES = 0;

export type ApiMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE' | 'HEAD' | 'OPTIONS';

export interface ApiErrorShape {
	error?: string;
	requestId?: string;
	[key: string]: unknown;
}

export class ApiClientError extends Error {
	readonly status: number;
	readonly path: string;
	readonly requestId: string;
	readonly payload: unknown;
	readonly userMessage: string;

	constructor(
		message: string,
		{
			status = 0,
			path = '',
			requestId = '',
			payload = null,
			userMessage = normalizeErrorMessage(status, message)
		}: {
			status?: number;
			path?: string;
			requestId?: string;
			payload?: unknown;
			userMessage?: string;
		} = {}
	) {
		super(message || 'Request failed');
		this.name = 'ApiClientError';
		this.status = status;
		this.path = path;
		this.requestId = requestId;
		this.payload = payload;
		this.userMessage = userMessage;
	}
}

export interface ApiClientOptions {
	fetchImpl?: typeof fetch;
	timeoutMs?: number;
	retries?: number;
}

export interface ApiRequestOptions<TBody> {
	method?: ApiMethod;
	headers?: HeadersInit;
	body?: TBody;
	timeoutMs?: number;
	retries?: number;
	signal?: AbortSignal;
	authToken?: string;
	csrfToken?: string;
}

export interface ApiClient {
	request<TResponse, TBody = undefined>(path: string, options?: ApiRequestOptions<TBody>): Promise<TResponse>;
	send<TResponse, TBody = Record<string, unknown>>(
		path: string,
		body?: TBody,
		method?: Extract<ApiMethod, 'POST' | 'PUT' | 'PATCH' | 'DELETE'>,
		options?: Omit<ApiRequestOptions<TBody>, 'method' | 'body'>
	): Promise<TResponse>;
}

export function normalizeErrorMessage(status: number, message = ''): string {
	if (status === 0) {
		return 'Xuva could not reach the local server. Check that the server is running and retry.';
	}
	if (status === 401) {
		// Prefer the server's actual message when present (e.g. "invalid username or
		// password" from the login endpoint). The "session no longer active" copy
		// only makes sense for spontaneous 401s on background requests, not for
		// failed login attempts where the user expects to see *why* it failed.
		if (message) return message;
		return 'Your session is no longer active. Sign in again to continue.';
	}
	if (status === 403) return message || 'This action requires permission or a valid CSRF token.';
	if (status === 404) return message || 'Xuva could not find that item.';
	if (status === 409) return message || 'This action conflicts with current server state. Refresh and retry.';
	if (status >= 500) return 'The server failed while handling this action. Retry once and inspect Activity if needed.';
	return message || 'Something went wrong. Retry the action.';
}

// On a spontaneous 401 (the session expired or was revoked mid-use) we clear the
// stored token and bounce the user to the sign-in page so they aren't stranded on
// a now-broken view. Excluded:
//   • the auth endpoints themselves — a 401 from /api/auth/login is a bad-password
//     response the form shows inline, not an expired session; /api/auth/session is
//     the app's own "am I signed in?" probe and is handled by the layout guard.
//   • requests made while already on /signin (avoids a redirect loop).
// The current path is preserved as ?return= so sign-in can send the user back.
function redirectToSignInOn401(path: string): void {
	if (typeof window === 'undefined' || !window.location) return;
	if (path.startsWith('/api/auth/')) return;
	try {
		const current = window.location.pathname || '/';
		if (current.startsWith('/signin') || current.startsWith('/setup')) return;
		const ret = encodeURIComponent(current + (window.location.search || ''));
		window.location.assign(`/signin?return=${ret}`);
	} catch {
		// Navigation may be unavailable in test/SSR contexts — fail quietly.
	}
}

function readCookie(name: string): string {
	if (typeof document === 'undefined' || !document.cookie) return '';
	const escapedName = name.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
	const match = document.cookie.match(new RegExp(`(?:^|; )${escapedName}=([^;]*)`));
	return match ? decodeURIComponent(match[1]) : '';
}

function withTimeout(signal: AbortSignal | undefined, timeoutMs: number): {
	signal: AbortSignal | undefined;
	cancel: () => void;
} {
	if (!timeoutMs || typeof AbortController === 'undefined') {
		return { signal, cancel: () => {} };
	}
	const controller = new AbortController();
	const timeout = setTimeout(() => controller.abort(), timeoutMs);
	if (signal) {
		if (signal.aborted) controller.abort();
		signal.addEventListener('abort', () => controller.abort(), { once: true });
	}
	return {
		signal: controller.signal,
		cancel: () => clearTimeout(timeout)
	};
}

function parsePayload(text: string, path: string, status: number): unknown {
	if (!text) return {};
	try {
		return JSON.parse(text);
	} catch {
		throw new ApiClientError('Server returned unreadable data.', {
			status,
			path,
			userMessage:
				'Xuva could not read media data from the local server. Check that the server is running and retry.'
		});
	}
}

function shouldRetry(error: ApiClientError, attempt: number, retries: number): boolean {
	if (attempt >= retries) return false;
	return error.status === 0 || error.status === 408 || error.status === 429 || error.status >= 500;
}

function toHeaders(options: ApiRequestOptions<unknown>): Headers {
	const headers = new Headers(options.headers);

	if (!headers.has('X-Auth-Token') && !headers.has('Authorization')) {
		const token = options.authToken ? String(options.authToken).trim() : readAuthToken();
		if (token) headers.set('X-Auth-Token', token);
	}

	// Inject active profile token so the server can enforce rating ceilings.
	if (!headers.has('X-Profile-Token')) {
		const profileToken = readProfileToken();
		if (profileToken) headers.set('X-Profile-Token', profileToken);
	}

	const method = String(options.method || 'GET').toUpperCase();
	if (!SAFE_METHODS.has(method) && !headers.has('X-CSRF-Token')) {
		const token = options.csrfToken ? String(options.csrfToken).trim() : readCookie('xuva_csrf');
		if (token) headers.set('X-CSRF-Token', token);
	}

	return headers;
}

function toRequestInit<TBody>(options: ApiRequestOptions<TBody>): RequestInit {
	const method = (options.method || 'GET').toUpperCase() as ApiMethod;
	const headers = toHeaders(options as ApiRequestOptions<unknown>);
	const init: RequestInit = {
		method,
		credentials: 'include',
		headers,
		signal: options.signal
	};

	if (typeof options.body !== 'undefined') {
		const body = options.body as unknown;
		if (
			typeof FormData !== 'undefined' && body instanceof FormData ||
			typeof URLSearchParams !== 'undefined' && body instanceof URLSearchParams ||
			typeof Blob !== 'undefined' && body instanceof Blob ||
			typeof ArrayBuffer !== 'undefined' && body instanceof ArrayBuffer
		) {
			init.body = body as BodyInit;
		} else if (typeof body === 'string') {
			if (!headers.has('Content-Type')) headers.set('Content-Type', 'text/plain;charset=UTF-8');
			init.body = body;
		} else {
			if (!headers.has('Content-Type')) headers.set('Content-Type', 'application/json');
			init.body = JSON.stringify(body ?? {});
		}
	}

	return init;
}

export function createApiClient({
	fetchImpl = globalThis.fetch,
	timeoutMs = DEFAULT_TIMEOUT_MS,
	retries = DEFAULT_RETRIES
}: ApiClientOptions = {}): ApiClient {
	if (!fetchImpl) throw new Error('fetch is required');

	async function requestOnce<TResponse, TBody = undefined>(
		path: string,
		options: ApiRequestOptions<TBody> = {}
	): Promise<TResponse> {
		const requestOptions = toRequestInit(options);
		const { signal, cancel } = withTimeout(requestOptions.signal ?? options.signal, options.timeoutMs ?? timeoutMs);
		requestOptions.signal = signal;

		try {
			const response = await fetchImpl(path, requestOptions);
			const rotatedToken = response.headers?.get?.('X-Auth-Token') || '';
			if (rotatedToken) writeAuthToken(rotatedToken);

			const rawText = await response.text();
			const payload = parsePayload(rawText, path, response.status) as ApiErrorShape;

			if (!response.ok) {
				if (response.status === 401) {
						clearAuthToken();
						redirectToSignInOn401(path);
					}
				throw new ApiClientError(payload?.error || response.statusText, {
					status: response.status,
					path,
					payload,
					requestId: response.headers?.get?.('X-Request-ID') || String(payload?.requestId || '')
				});
			}

			return payload as TResponse;
		} catch (error) {
			if (error instanceof ApiClientError) throw error;
			const aborted = error instanceof Error && error.name === 'AbortError';
			throw new ApiClientError(aborted ? 'Request timed out.' : String((error as Error)?.message || error), {
				status: 0,
				path,
				userMessage: aborted
					? 'The server took too long to answer. Retry the action or inspect Activity for long-running jobs.'
					: normalizeErrorMessage(0)
			});
		} finally {
			cancel();
		}
	}

	// In-flight deduplication: identical simultaneous GET requests share one
	// network call instead of each firing their own fetch. The promise is
	// removed from the map as soon as it settles, so subsequent navigations
	// always issue a fresh request.
	const inflight = new Map<string, Promise<unknown>>();

	async function requestWithRetry<TResponse, TBody = undefined>(
		path: string,
		options: ApiRequestOptions<TBody> = {}
	): Promise<TResponse> {
		const retryCount = Number.isFinite(options.retries) ? Number(options.retries) : retries;
		let attempt = 0;
		for (;;) {
			try {
				return await requestOnce<TResponse, TBody>(path, options);
			} catch (error) {
				const clientError =
					error instanceof ApiClientError
						? error
						: new ApiClientError(String((error as Error)?.message || error), { path, status: 0 });
				if (!shouldRetry(clientError, attempt, retryCount)) throw clientError;
				attempt += 1;
				await new Promise((resolve) => setTimeout(resolve, 120 * attempt));
			}
		}
	}

	async function request<TResponse, TBody = undefined>(
		path: string,
		options: ApiRequestOptions<TBody> = {}
	): Promise<TResponse> {
		const method = (options.method || 'GET').toUpperCase();

		// Only deduplicate GET requests with no custom abort signal.
		if (method === 'GET' && !options.signal) {
			const existing = inflight.get(path) as Promise<TResponse> | undefined;
			if (existing) return existing;
			const promise = requestWithRetry<TResponse, TBody>(path, options);
			inflight.set(path, promise as Promise<unknown>);
			// Clean up the map on settle. Use then(ok, err) instead of .finally() so
			// the cleanup side-chain resolves (not re-throws), avoiding an unhandled
			// rejection that the test environment would treat as a fatal error.
			promise.then(() => inflight.delete(path), () => inflight.delete(path));
			return promise;
		}

		return requestWithRetry<TResponse, TBody>(path, options);
	}

	function send<TResponse, TBody = Record<string, unknown>>(
		path: string,
		body?: TBody,
		method: Extract<ApiMethod, 'POST' | 'PUT' | 'PATCH' | 'DELETE'> = 'POST',
		options: Omit<ApiRequestOptions<TBody>, 'method' | 'body'> = {}
	): Promise<TResponse> {
		return request<TResponse, TBody>(path, { ...options, method, body });
	}

	return { request, send };
}

export const apiClient = createApiClient();
