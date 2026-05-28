import type { Media } from '$lib/mock-data';
import type { MovieListItem, SeriesListItem } from './browse';
import type { ClientHomeItem } from './home';

// Decode HTML entities (e.g. &amp; → &) that TMDb and other metadata sources
// occasionally return. Svelte renders text nodes verbatim so entities show as
// literal text. Decode at the API boundary before storing in reactive state.
function decodeEntities(text: string | undefined): string {
	if (!text) return '';
	if (typeof document === 'undefined') return text;
	const el = document.createElement('textarea');
	el.innerHTML = text;
	return el.value;
}

// ---------------------------------------------------------------------------
// Runtime string → minutes
// ---------------------------------------------------------------------------

/** Parse "2h 13m", "1h 45m", "143", "90m" → minutes. Returns undefined if
 *  the string is falsy or unparseable. */
export function parseRuntimeMins(runtime: string | undefined): number | undefined {
	if (!runtime) return undefined;
	const minsOnly = runtime.match(/^(\d+)$/);
	if (minsOnly) return parseInt(minsOnly[1], 10);
	const hm = runtime.match(/(\d+)h\s*(\d+)m/);
	if (hm) return parseInt(hm[1], 10) * 60 + parseInt(hm[2], 10);
	const hOnly = runtime.match(/(\d+)h/);
	if (hOnly) return parseInt(hOnly[1], 10) * 60;
	const mOnly = runtime.match(/(\d+)m/);
	if (mOnly) return parseInt(mOnly[1], 10);
	return undefined;
}

// ---------------------------------------------------------------------------
// Title / year helpers
// ---------------------------------------------------------------------------

/**
 * Strip quality/codec tags and file extensions from raw filenames so we
 * show "22 Jump Street" instead of "22 Jump Street (2014) (Remux-2160p).mkv".
 */
function cleanTitle(raw: string): string {
	// Remove trailing file extension (.mkv, .mp4, .avi, .m4v, …)
	let t = raw.replace(/\.[a-z0-9]{2,4}$/i, '');
	// Remove parenthesised quality / codec / source tags (keep only year-style tokens)
	t = t.replace(
		/\s*\([^)]*(?:remux|bluray|blu-ray|webrip|web-dl|webdl|hdtv|dvdrip|dvdscr|4k|2160p|1080p|720p|480p|hdr|dv|atmos|dts|aac|ac3|x264|x265|hevc|avc|h\.264|h\.265)[^)]*\)/gi,
		''
	);
	return t.trim();
}

/**
 * Try to extract a 4-digit year (1900–2099) from a string like "Title (2014)".
 * Returns 0 if nothing found.
 */
function extractYear(s: string): number {
	const match = s.match(/\((\d{4})\)/);
	if (match) {
		const yr = parseInt(match[1], 10);
		if (yr >= 1900 && yr <= 2099) return yr;
	}
	return 0;
}

// ---------------------------------------------------------------------------
// Palette table
// ---------------------------------------------------------------------------

// Curated dark palettes for items without poster-derived colours.
// Selected deterministically by hashing the item id so the same item
// always gets the same palette across renders.
const PALETTES: Array<{ palette: [string, string]; accent: string }> = [
	{ palette: ['#0f172a', '#1e3a8a'], accent: '#60a5fa' },
	{ palette: ['#1c1917', '#7c2d12'], accent: '#fb923c' },
	{ palette: ['#0c0a09', '#451a03'], accent: '#fbbf24' },
	{ palette: ['#0f172a', '#4c1d95'], accent: '#a78bfa' },
	{ palette: ['#0a0a0a', '#450a0a'], accent: '#f87171' },
	{ palette: ['#0c1a2e', '#0e4f6e'], accent: '#67e8f9' },
	{ palette: ['#1a0a2e', '#6b21a8'], accent: '#d8b4fe' },
	{ palette: ['#1f1f1f', '#374151'], accent: '#e5e7eb' },
	{ palette: ['#1a1a1a', '#7f1d1d'], accent: '#fca5a5' },
	{ palette: ['#0a0a1a', '#1e3a5f'], accent: '#93c5fd' },
	{ palette: ['#1a0f0a', '#78350f'], accent: '#fde68a' },
	{ palette: ['#0c1a0c', '#14532d'], accent: '#4ade80' },
];

function hashPalette(id: string): { palette: [string, string]; accent: string } {
	let h = 0;
	for (let i = 0; i < id.length; i++) h = (h + id.charCodeAt(i)) % PALETTES.length;
	return PALETTES[h];
}

export function movieToMedia(item: MovieListItem): Media {
	const id = item.id ?? crypto.randomUUID();
	const meta = item.metadata as Record<string, unknown> | undefined;
	const rawTitle = item.title ?? (meta?.title as string | undefined) ?? 'Unknown';
	const runtimeStr = (meta?.runtime as string | undefined) || undefined;
	return {
		id,
		title: cleanTitle(rawTitle),
		year: item.year ?? (item.metadata?.year as number | undefined) ?? 0,
		type: 'Movie',
		genres: (meta?.genres as string[] | undefined) ?? [],
		rating: (meta?.voteAverage as number | undefined) ?? 0,
		runtime: runtimeStr,
		runtimeMins: parseRuntimeMins(runtimeStr),
		synopsis: decodeEntities(item.metadata?.overview ?? ''),
		poster: item.metadata?.posterUrl || undefined,
		backdrop: item.metadata?.backdropUrl || undefined,
		contentRating: (meta?.contentRating as string | undefined) || undefined,
		needsReview: item.needsReview ?? false,
		probed: item.probed ?? true,
		versionCount: item.versionCount ?? 1,
		addedAt: item.addedAt || undefined,
		watched: item.watched ?? false,
		studio: (meta?.studios as string[] | undefined) ?? [],
		...hashPalette(id),
	};
}

export function seriesToMedia(item: SeriesListItem): Media {
	const id = item.id ?? crypto.randomUUID();
	const meta = item.metadata as Record<string, unknown> | undefined;
	const rawTitle = item.title ?? (meta?.title as string | undefined) ?? 'Unknown';
	const unknownItem = item as Record<string, unknown>;
	return {
		id,
		title: cleanTitle(rawTitle),
		year: (item.metadata?.year as number | undefined) ?? 0,
		type: 'Series',
		genres: (meta?.genres as string[] | undefined) ?? [],
		rating: (meta?.voteAverage as number | undefined) ?? 0,
		seasons: item.seasonCount,
		episodes: item.episodeCount,
		unwatchedCount: (unknownItem.unwatchedCount as number | undefined) ?? undefined,
		synopsis: decodeEntities(item.metadata?.overview ?? ''),
		poster: item.metadata?.posterUrl || undefined,
		backdrop: item.metadata?.backdropUrl || undefined,
		contentRating: (meta?.contentRating as string | undefined) || undefined,
		needsReview: item.needsReview ?? false,
		versionCount: unknownItem.versionCount as number | undefined,
		addedAt: item.addedAt || undefined,
		watched: item.watched ?? false,
		// For series, prefer networks (streaming home) then fall back to studios
		studio: ((meta?.networks as string[] | undefined) ?? (meta?.studios as string[] | undefined) ?? []),
		...hashPalette(id),
	};
}

export function clientHomeItemToMedia(item: ClientHomeItem): Media {
	const id = item.id ?? crypto.randomUUID();
	const type = item.kind === 'series' ? 'Series' : 'Movie';
	const raw = item.title as Record<string, unknown> | undefined;
	const rawTitle = (typeof item.title === 'string' ? item.title : null) ?? 'Unknown';
	const unknownFields = item as Record<string, unknown>;
	return {
		id,
		title: cleanTitle(rawTitle),
		// year / rating may be in the loose [key: string] bag the server returns
		year: (unknownFields.year as number | undefined) ?? extractYear(rawTitle) ?? 0,
		type,
		genres: (unknownFields.genres as string[] | undefined) ?? [],
		rating: (unknownFields.rating as number | undefined) ?? (unknownFields.voteAverage as number | undefined) ?? 0,
		seasons: (unknownFields.seasonCount as number | undefined) ?? (unknownFields.seasons as number | undefined),
		episodes: (unknownFields.episodeCount as number | undefined) ?? (unknownFields.episodes as number | undefined),
		synopsis: decodeEntities(item.overview ?? item.description ?? ''),
		// Normalise empty strings to undefined: the API returns "" for items
		// without artwork, and `??` fallbacks downstream only catch null/
		// undefined — leaving "" in place causes <img src=""> to render nothing.
		poster: item.posterUrl || undefined,
		backdrop: item.backdropUrl || undefined,
		logo: (unknownFields.logoUrl as string | undefined) || undefined,
		thumbnail: (unknownFields.thumbnailUrl as string | undefined) || undefined,
		banner: (unknownFields.bannerUrl as string | undefined) || undefined,
		videoKey: (unknownFields.videoKey as string | undefined) || undefined,
		trailerUrl: (unknownFields.trailerUrl as string | undefined) || undefined,
		parentId: (unknownFields.parentId as string | undefined) || undefined,
		parentKind: (unknownFields.parentKind as string | undefined) || undefined,
		seasonNumber: (unknownFields.seasonNumber as number | undefined) ?? undefined,
		episodeNumber: (unknownFields.episodeNumber as number | undefined) ?? undefined,
		episodeTitle: (unknownFields.episodeTitle as string | undefined) || undefined,
		progress: typeof item.percent === 'number' ? item.percent / 100
			: typeof item.progressPercent === 'number' ? item.progressPercent / 100
			: undefined,
		...hashPalette(id),
	};
}
