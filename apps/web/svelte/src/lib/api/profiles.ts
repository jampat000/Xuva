import { apiClient, type ApiClient } from './client';
import { writeProfileToken, clearProfileToken } from './profile-token-store';

export interface ProfileCard {
	id: string;
	displayName: string;
	avatarUrl?: string;
	avatarPreset?: string;
	avatarColor?: string;
	isRestricted: boolean;
	hasEntryPin: boolean;
	hasExitPin: boolean;
}

export interface SwitchProfileRequest {
	profileUserId: string;
	currentProfilePin?: string;
	targetProfilePin?: string;
}

export interface SwitchProfileResponse {
	profileToken: string;
	profile: ProfileCard;
}

export interface UpdateProfileRequest {
	displayName: string;
	avatarUrl?: string;
	avatarPreset?: string;
	avatarColor?: string;
	isRestricted?: boolean;
	maxRating?: string;
}

export interface SetPinRequest {
	pin: string; // empty = clear pin
}

export async function listProfiles(client: ApiClient = apiClient): Promise<ProfileCard[]> {
	const resp = await client.request<{ profiles: ProfileCard[] }>('/api/profiles');
	return resp.profiles ?? [];
}

export async function switchProfile(
	payload: SwitchProfileRequest,
	currentProfileToken?: string,
	client: ApiClient = apiClient
): Promise<SwitchProfileResponse> {
	const headers: Record<string, string> = {};
	if (currentProfileToken) {
		headers['X-Profile-Token'] = currentProfileToken;
	}
	const resp = await client.request<SwitchProfileResponse, SwitchProfileRequest>('/api/auth/switch-profile', {
		method: 'POST',
		body: payload,
		headers
	});
	// Persist the new profile token.
	if (resp.profileToken) writeProfileToken(resp.profileToken);
	return resp;
}

export function clearActiveProfile(): void {
	clearProfileToken();
}

export async function updateProfileSettings(
	userID: string,
	payload: UpdateProfileRequest,
	client: ApiClient = apiClient
): Promise<{ user: unknown }> {
	return client.send<{ user: unknown }, UpdateProfileRequest>(
		`/api/users/${encodeURIComponent(userID)}`,
		payload,
		'PATCH'
	);
}

export async function setProfilePin(
	userID: string,
	pin: string,
	client: ApiClient = apiClient
): Promise<{ status: string }> {
	return client.send<{ status: string }, SetPinRequest>(
		`/api/users/${encodeURIComponent(userID)}/pin`,
		{ pin }
	);
}

/** All available rating options for the ceiling selector. */
export const RATING_OPTIONS = [
	{ value: '', label: 'Unrestricted' },
	{ value: 'TV-Y', label: 'TV-Y (All children)' },
	{ value: 'TV-Y7', label: 'TV-Y7 (Children 7+)' },
	{ value: 'G', label: 'G / TV-G' },
	{ value: 'PG', label: 'PG / TV-PG' },
	{ value: 'TV-14', label: 'TV-14' },
	{ value: 'PG-13', label: 'PG-13' },
	{ value: 'R', label: 'R (Restricted)' },
	{ value: 'TV-MA', label: 'TV-MA (Mature)' },
	{ value: 'NC-17', label: 'NC-17' },
] as const;

/** Preset avatar identifiers — must match filenames in /avatars/. */
export const AVATAR_PRESETS = [
	'cat', 'dog', 'fox', 'bear', 'rabbit',
	'owl', 'penguin', 'fish', 'lion', 'panda',
] as const;

export type AvatarPreset = typeof AVATAR_PRESETS[number];
