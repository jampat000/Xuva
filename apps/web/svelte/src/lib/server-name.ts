export const DEFAULT_SERVER_NAME = 'Lorivo';

export function normalizeServerName(value: unknown): string {
	const trimmed = String(value ?? '').trim();
	return trimmed || DEFAULT_SERVER_NAME;
}

export function lorivoTitle(value: unknown): string {
	const name = normalizeServerName(value);
	return name === DEFAULT_SERVER_NAME ? DEFAULT_SERVER_NAME : `${name} · Lorivo`;
}
