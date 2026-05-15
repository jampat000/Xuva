import {
	getAuthSessionIfAvailable,
	type AuthSessionResponse,
	type AuthSessionUser
} from '$lib/api/auth';
import { ApiClientError, apiClient } from '$lib/api/client';
import {
	getClientHome,
	getLibraries,
	getPlaybackRecent,
	type ClientHomeResponse,
	type ClientHomeRow,
	type LibrariesResponse,
	type PlaybackRecentResponse
} from '$lib/api/home';
import {
	buildHomeViewModel,
	type HomeDisplayItem,
	type HomeViewModel
} from '$lib/home/model';

export interface SecondaryRouteContext {
	user: AuthSessionUser | null;
	libraries: NonNullable<LibrariesResponse['libraries']>;
	homePayload: ClientHomeResponse;
	playbackRecentPayload: PlaybackRecentResponse;
	model: HomeViewModel;
}

export type SecondaryRouteLoadState =
	| {
			kind: 'ready';
			context: SecondaryRouteContext;
	  }
	| {
			kind: 'auth';
			message: string;
	  }
	| {
			kind: 'error';
			error: unknown;
	  };

export async function loadSecondaryRouteContext({
	forceEmpty = false,
	limit = 48
}: {
	forceEmpty?: boolean;
	limit?: number;
}): Promise<SecondaryRouteContext> {
	const sessionPayload = await getAuthSessionIfAvailable(apiClient).catch((error: unknown) => {
		if (isApiStatus(error, 401)) return null;
		throw error;
	});

	let homePayload: ClientHomeResponse;
	let playbackRecentPayload: PlaybackRecentResponse;
	let librariesPayload: LibrariesResponse;

	try {
		[homePayload, playbackRecentPayload, librariesPayload] = await Promise.all([
			getClientHome(apiClient, limit),
			getPlaybackRecent(apiClient, Math.max(12, Math.round(limit / 2))),
			getLibraries(apiClient)
		]);
	} catch (error) {
		if (isApiStatus(error, 401) && forceEmpty) {
			homePayload = {};
			playbackRecentPayload = { recent: [] };
			librariesPayload = { libraries: [] };
		} else {
			throw error;
		}
	}

	const model = buildHomeViewModel({
		homePayload,
		playbackRecentPayload,
		librariesPayload,
		sessionPayload: sessionPayload as AuthSessionResponse | null,
		forceEmpty
	});

	return {
		user: model.user || null,
		libraries: model.libraries || [],
		homePayload,
		playbackRecentPayload,
		model
	};
}

export async function loadSecondaryRouteContextSafe({
	forceEmpty = false,
	limit = 48
}: {
	forceEmpty?: boolean;
	limit?: number;
}): Promise<SecondaryRouteLoadState> {
	try {
		const context = await loadSecondaryRouteContext({ forceEmpty, limit });
		return { kind: 'ready', context };
	} catch (error) {
		if (isApiStatus(error, 401) && !forceEmpty) {
			return {
				kind: 'auth',
				message: 'Your session has expired. Sign in again to continue.'
			};
		}
		return { kind: 'error', error };
	}
}

export function findRow(homePayload: ClientHomeResponse, rowID: string): ClientHomeRow | null {
	const target = asText(rowID).toLowerCase();
	for (const row of homePayload.rows || []) {
		if (asText(row?.id).toLowerCase() === target) return row;
	}
	return null;
}

export function formatLoadError(error: unknown, fallbackLabel: string): string {
	if (error instanceof ApiClientError) return error.userMessage || error.message;
	if (isApiStatus(error, 401)) return 'Your session is no longer active. Sign in again to continue.';
	if (isApiStatus(error, 403))
		return 'This account does not have access to this route in the current session.';
	if (error instanceof Error) return error.message;
	return `${fallbackLabel} could not load.`;
}

export function isApiStatus(error: unknown, expectedStatus: number): boolean {
	if (error instanceof ApiClientError) return error.status === expectedStatus;
	if (typeof error !== 'object' || !error) return false;
	const candidate = (error as { status?: unknown }).status;
	return Number(candidate) === expectedStatus;
}

export function initialsForName(name: string): string {
	const words = asText(name).split(/\s+/).filter(Boolean);
	if (words.length === 0) return 'V';
	if (words.length === 1) return words[0].slice(0, 1).toUpperCase();
	return `${words[0][0] || ''}${words[1][0] || ''}`.toUpperCase();
}

export function itemDetailHref(item: HomeDisplayItem): string | undefined {
	const id = asText(item.id);
	if (!id) return undefined;
	if (item.kind === 'movie') return `/movies/${encodeURIComponent(id)}`;
	if (item.kind === 'series') return `/tv/${encodeURIComponent(id)}`;
	return undefined;
}

export function itemPlayHref(item: HomeDisplayItem): string | undefined {
	const mediaSourceID = asText(item.playMediaSourceId || item.mediaSourceId);
	if (!mediaSourceID) return undefined;
	return `/play/${encodeURIComponent(mediaSourceID)}`;
}

export function asText(value: unknown): string {
	return String(value ?? '').trim();
}
