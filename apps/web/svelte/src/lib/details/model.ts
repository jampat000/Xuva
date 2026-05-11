import type { EpisodeBrief, PlaybackDecisionResponse, PlaybackStateResponse } from '$lib/api/details';

const RELEASE_TOKENS = [
	'webrip',
	'web-dl',
	'webdl',
	'bluray',
	'brrip',
	'bdrip',
	'dvdrip',
	'hdrip',
	'remux',
	'proper',
	'repack',
	'extended',
	'criterion',
	'unrated',
	'x264',
	'x265',
	'h264',
	'h265',
	'hevc',
	'av1',
	'aac',
	'ac3',
	'dts',
	'truehd',
	'atmos',
	'2160p',
	'1080p',
	'720p',
	'480p',
	'hdr',
	'uhd',
	'10bit'
];

export function cleanDisplayTitle(rawValue: string): string {
	let value = asText(rawValue);
	if (!value) return 'Untitled';
	value = value.replace(/\.[a-z0-9]{2,4}$/i, ' ');
	value = value.replace(/[_.]/g, ' ');
	value = value.replace(/\[[^\]]*\]/g, ' ');
	value = value.replace(/\bS\d{1,2}E\d{1,2}(?:E\d{1,2})?\b/gi, ' ');
	value = value.replace(/\b(19|20)\d{2}\b/g, ' ');
	for (const token of RELEASE_TOKENS) {
		value = value.replace(new RegExp(`\\b${escapeRegex(token)}\\b`, 'gi'), ' ');
	}
	value = value
		.replace(/\(\s*\)/g, ' ')
		.replace(/\s+-\s+/g, ' ')
		.replace(/[-:]\s*$/g, ' ')
		.replace(/\s+/g, ' ')
		.trim();
	return value || 'Untitled';
}

export function cleanDescription(value: unknown, maxLength = 320): string {
	const text = asText(value).replace(/\s+/g, ' ').trim();
	if (!text) return '';
	if (text.length <= maxLength) return text;
	return `${text.slice(0, Math.max(0, maxLength - 3)).trimEnd()}...`;
}

export function extractYear(value: unknown): number {
	const match = asText(value).match(/\b(19|20)\d{2}\b/);
	if (!match) return 0;
	const parsed = Number(match[0]);
	return Number.isFinite(parsed) ? parsed : 0;
}

export function resolveArtworkUrl(
	kind: 'movie' | 'series',
	id: string,
	metadataValue: unknown,
	type: 'poster' | 'backdrop'
): string {
	const provided = asText(metadataValue);
	if (provided) return provided;
	if (!id) return '';
	return `/api/artwork/${encodeURIComponent(kind)}/${encodeURIComponent(id)}?style=neutral&type=${type}`;
}

export function formatRuntime(durationSeconds: number): string {
	if (!Number.isFinite(durationSeconds) || durationSeconds <= 0) return '';
	const minutes = Math.max(1, Math.round(durationSeconds / 60));
	const hours = Math.floor(minutes / 60);
	const remainder = minutes % 60;
	if (hours <= 0) return `${minutes} min`;
	if (remainder <= 0) return `${hours}h`;
	return `${hours}h ${remainder}m`;
}

export function formatDuration(durationSeconds: number): string {
	if (!Number.isFinite(durationSeconds) || durationSeconds <= 0) return 'Unknown duration';
	const hours = Math.floor(durationSeconds / 3600);
	const minutes = Math.floor((durationSeconds % 3600) / 60);
	const seconds = Math.floor(durationSeconds % 60);
	if (hours > 0) return `${hours}h ${minutes}m ${seconds}s`;
	if (minutes > 0) return `${minutes}m ${seconds}s`;
	return `${seconds}s`;
}

export function formatBytes(value: number): string {
	if (!Number.isFinite(value) || value <= 0) return 'Unknown size';
	const units = ['B', 'KB', 'MB', 'GB', 'TB'];
	let amount = value;
	let unit = 0;
	while (amount >= 1024 && unit < units.length - 1) {
		amount /= 1024;
		unit += 1;
	}
	const digits = amount >= 100 || unit === 0 ? 0 : amount >= 10 ? 1 : 2;
	return `${amount.toFixed(digits)} ${units[unit]}`;
}

export function formatBitrate(value: number): string {
	if (!Number.isFinite(value) || value <= 0) return 'Unknown bitrate';
	const mbps = value / 1_000_000;
	if (mbps >= 1) return `${mbps.toFixed(mbps >= 10 ? 1 : 2)} Mbps`;
	return `${Math.round(value / 1000)} Kbps`;
}

export function formatResolution(width: number, height: number): string {
	if (!Number.isFinite(width) || !Number.isFinite(height) || width <= 0 || height <= 0) {
		return 'Unknown resolution';
	}
	return `${Math.round(width)}x${Math.round(height)}`;
}

export function formatTrackSummary(codec: unknown, language: unknown, channels?: number): string {
	const codecLabel = asText(codec) ? asText(codec).toUpperCase() : 'Unknown codec';
	const languageLabel = asText(language) && asText(language).toLowerCase() !== 'und'
		? asText(language).toUpperCase()
		: 'Unknown language';
	if (Number.isFinite(channels) && Number(channels) > 0) {
		return `${codecLabel} - ${languageLabel} - ${Math.round(Number(channels))}ch`;
	}
	return `${codecLabel} - ${languageLabel}`;
}

export function playbackPercent(state: PlaybackStateResponse | null): number {
	const percentValue = Number(state?.percent ?? 0);
	if (Number.isFinite(percentValue) && percentValue > 0 && percentValue <= 1) {
		return Math.round(percentValue * 100);
	}
	if (Number.isFinite(percentValue) && percentValue > 1) {
		return Math.round(Math.max(0, Math.min(100, percentValue)));
	}
	const progress = Number(state?.progressSeconds ?? 0);
	const duration = Number(state?.durationSeconds ?? 0);
	if (duration > 0 && progress > 0) {
		return Math.round(Math.max(0, Math.min(100, (progress / duration) * 100)));
	}
	return 0;
}

export function isResumeState(state: PlaybackStateResponse | null): boolean {
	if (!state) return false;
	if (state.watched) return false;
	return Number(state.progressSeconds || 0) > 5;
}

export function watchedLabel(state: PlaybackStateResponse | null): string {
	if (!state) return 'Unplayed';
	if (state.watched) return 'Watched';
	const percent = playbackPercent(state);
	if (percent > 0) return `${percent}% watched`;
	return 'Unplayed';
}

export function playbackModeLabel(decision: PlaybackDecisionResponse | null): string {
	const mode = asText(decision?.mode);
	if (mode) return mode;
	return 'Pending decision';
}

export function playbackReasonLabel(decision: PlaybackDecisionResponse | null): string {
	const reasonText = asText(decision?.reasonText) || asText(decision?.reason);
	if (reasonText) return reasonText;
	return 'Playback route has not been resolved yet.';
}

export function episodeLabel(episode: EpisodeBrief): string {
	const season = Number(episode.seasonNumber || 0);
	const number = Number(episode.episodeNumber || 0);
	const end = Number(episode.episodeEnd || 0);
	if (season <= 0 || number <= 0) return 'Episode';
	if (end > number) return `S${pad2(season)} E${pad2(number)}-${pad2(end)}`;
	return `S${pad2(season)} E${pad2(number)}`;
}

export function buildSeriesCardMeta(seasonCount: number, episodeCount: number): string {
	const seasonLabel = `${seasonCount} season${seasonCount === 1 ? '' : 's'}`;
	const episodeLabel = `${episodeCount} episode${episodeCount === 1 ? '' : 's'}`;
	return `${seasonLabel} - ${episodeLabel}`;
}

export function sourceQualityLabel(qualityLabel: unknown, resolutionLabel: string): string {
	const explicitLabel = asText(qualityLabel);
	if (explicitLabel) return explicitLabel;
	if (resolutionLabel && resolutionLabel !== 'Unknown resolution') return resolutionLabel;
	return 'Source';
}

export function listItemTitle(title: unknown, fallback: unknown): string {
	const cleaned = cleanDisplayTitle(asText(title));
	if (cleaned !== 'Untitled') return cleaned;
	return cleanDisplayTitle(asText(fallback) || 'Untitled');
}

function pad2(value: number): string {
	return String(Math.max(0, Math.round(value))).padStart(2, '0');
}

function asText(value: unknown): string {
	return String(value ?? '').trim();
}

function escapeRegex(value: string): string {
	return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}
