import type { MovieListItem, SeriesListItem } from '$lib/api/browse';

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

export type MovieFilter = 'all' | 'review' | 'metadata' | 'versions';
export type MovieSort = 'title' | 'year' | 'versions' | 'review';
export type SeriesFilter = 'all' | 'multi-season' | 'with-episodes' | 'unknown-year';
export type SeriesSort = 'title' | 'year' | 'seasons' | 'episodes';

export interface MovieCardModel {
	id: string;
	title: string;
	year: number;
	yearLabel: string;
	needsReview: boolean;
	versionCount: number;
	hasMetadata: boolean;
	meta: string;
	posterUrl: string;
	searchText: string;
}

export interface SeriesCardModel {
	id: string;
	title: string;
	year: number;
	yearLabel: string;
	seasonCount: number;
	episodeCount: number;
	meta: string;
	posterUrl: string;
	searchText: string;
}

export function buildMovieCards(items: MovieListItem[]): MovieCardModel[] {
	return items
		.map((item) => {
			const id = asText(item.id);
			if (!id) return null;

			const metadataTitle = asText(item.metadata?.title);
			const rawTitle = metadataTitle || asText(item.title);
			const title = cleanDisplayTitle(rawTitle || 'Untitled');
			const year = resolveYear(item.year, item.metadata?.year, rawTitle);
			const versionCount = Math.max(0, Number(item.versionCount || 0));
			const needsReview = Boolean(item.needsReview);
			const hasMetadata = Boolean(item.metadata);
			const yearLabel = year > 0 ? String(year) : 'Unknown year';
			const versionLabel = `${versionCount} version${versionCount === 1 ? '' : 's'}`;
			const metaParts = [year > 0 ? yearLabel : '', versionLabel, needsReview ? 'Needs review' : '']
				.filter(Boolean)
				.join(' - ');
			const posterUrl = resolvePosterUrl('movie', id, asText(item.metadata?.posterUrl));
			const searchText = [title, yearLabel, metaParts].join(' ').toLowerCase();

			return {
				id,
				title,
				year,
				yearLabel,
				needsReview,
				versionCount,
				hasMetadata,
				meta: metaParts,
				posterUrl,
				searchText
			};
		})
		.filter((item): item is MovieCardModel => Boolean(item));
}

export function filterAndSortMovieCards(
	items: MovieCardModel[],
	needle: string,
	filter: MovieFilter,
	sort: MovieSort
): MovieCardModel[] {
	const loweredNeedle = asText(needle).toLowerCase();
	let output = items;
	if (loweredNeedle) {
		output = output.filter((item) => item.searchText.includes(loweredNeedle));
	}
	if (filter === 'review') {
		output = output.filter((item) => item.needsReview);
	}
	if (filter === 'metadata') {
		output = output.filter((item) => !item.hasMetadata);
	}
	if (filter === 'versions') {
		output = output.filter((item) => item.versionCount > 1);
	}
	return [...output].sort(movieSorter(sort));
}

export function buildSeriesCards(items: SeriesListItem[]): SeriesCardModel[] {
	return items
		.map((item) => {
			const id = asText(item.id);
			if (!id) return null;

			const metadataTitle = asText(item.metadata?.title);
			const rawTitle = metadataTitle || asText(item.title);
			const title = cleanDisplayTitle(rawTitle || 'Untitled');
			const year = resolveYear(0, item.metadata?.year, rawTitle);
			const yearLabel = year > 0 ? String(year) : '';
			const seasonCount = Math.max(0, Number(item.seasonCount || 0));
			const episodeCount = Math.max(0, Number(item.episodeCount || 0));
			const seasonLabel = `${seasonCount} season${seasonCount === 1 ? '' : 's'}`;
			const episodeLabel = `${episodeCount} episode${episodeCount === 1 ? '' : 's'}`;
			const meta = [seasonLabel, episodeLabel].filter(Boolean).join(' - ');
			const posterUrl = resolvePosterUrl('series', id, asText(item.metadata?.posterUrl));
			const searchText = [title, yearLabel, meta].join(' ').toLowerCase();

			return {
				id,
				title,
				year,
				yearLabel,
				seasonCount,
				episodeCount,
				meta,
				posterUrl,
				searchText
			};
		})
		.filter((item): item is SeriesCardModel => Boolean(item));
}

export function filterSeriesCards(items: SeriesCardModel[], needle: string): SeriesCardModel[] {
	const loweredNeedle = asText(needle).toLowerCase();
	if (!loweredNeedle) return items;
	return items.filter((item) => item.searchText.includes(loweredNeedle));
}

export function filterAndSortSeriesCards(
	items: SeriesCardModel[],
	needle: string,
	filter: SeriesFilter,
	sort: SeriesSort
): SeriesCardModel[] {
	const loweredNeedle = asText(needle).toLowerCase();
	let output = items;

	if (loweredNeedle) {
		output = output.filter((item) => item.searchText.includes(loweredNeedle));
	}
	if (filter === 'multi-season') {
		output = output.filter((item) => item.seasonCount > 1);
	}
	if (filter === 'with-episodes') {
		output = output.filter((item) => item.episodeCount > 0);
	}
	if (filter === 'unknown-year') {
		output = output.filter((item) => item.year <= 0);
	}

	return [...output].sort(seriesSorter(sort));
}

function movieSorter(sort: MovieSort): (a: MovieCardModel, b: MovieCardModel) => number {
	if (sort === 'year') {
		return (a, b) => b.year - a.year || a.title.localeCompare(b.title);
	}
	if (sort === 'versions') {
		return (a, b) => b.versionCount - a.versionCount || a.title.localeCompare(b.title);
	}
	if (sort === 'review') {
		return (a, b) => Number(b.needsReview) - Number(a.needsReview) || a.title.localeCompare(b.title);
	}
	return (a, b) => a.title.localeCompare(b.title);
}

function seriesSorter(sort: SeriesSort): (a: SeriesCardModel, b: SeriesCardModel) => number {
	if (sort === 'year') {
		return (a, b) => b.year - a.year || a.title.localeCompare(b.title);
	}
	if (sort === 'seasons') {
		return (a, b) => b.seasonCount - a.seasonCount || a.title.localeCompare(b.title);
	}
	if (sort === 'episodes') {
		return (a, b) => b.episodeCount - a.episodeCount || a.title.localeCompare(b.title);
	}
	return (a, b) => a.title.localeCompare(b.title);
}

function resolvePosterUrl(kind: 'movie' | 'series', id: string, metadataPosterUrl: string): string {
	if (metadataPosterUrl) return metadataPosterUrl;
	return '';
}

function resolveYear(
	primaryYear: number | undefined,
	metadataYear: number | undefined,
	title: string
): number {
	const metadata = Number(metadataYear || 0);
	if (metadata > 1800) return metadata;
	const item = Number(primaryYear || 0);
	if (item > 1800) return item;
	const extracted = Number(extractYear(title) || 0);
	return extracted > 1800 ? extracted : 0;
}

function cleanDisplayTitle(rawValue: string): string {
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

function extractYear(value: string): string {
	const match = asText(value).match(/\b(19|20)\d{2}\b/);
	return match ? match[0] : '';
}

function asText(value: unknown): string {
	return String(value ?? '').trim();
}

function escapeRegex(value: string): string {
	return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}
