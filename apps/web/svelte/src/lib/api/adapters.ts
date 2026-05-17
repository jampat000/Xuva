import type { Media } from '$lib/mock-data';
import type { MovieListItem, SeriesListItem } from './browse';
import type { ClientHomeItem } from './home';

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
	return {
		id,
		title: item.title ?? item.metadata?.title ?? 'Unknown',
		year: item.year ?? item.metadata?.year ?? 0,
		type: 'Movie',
		genres: [],
		rating: 0,
		synopsis: item.metadata?.overview ?? '',
		poster: item.metadata?.posterUrl,
		backdrop: item.metadata?.backdropUrl,
		...hashPalette(id),
	};
}

export function seriesToMedia(item: SeriesListItem): Media {
	const id = item.id ?? crypto.randomUUID();
	return {
		id,
		title: item.title ?? item.metadata?.title ?? 'Unknown',
		year: item.metadata?.year ?? 0,
		type: 'Series',
		genres: [],
		rating: 0,
		seasons: item.seasonCount,
		episodes: item.episodeCount,
		synopsis: item.metadata?.overview ?? '',
		poster: item.metadata?.posterUrl,
		backdrop: item.metadata?.backdropUrl,
		...hashPalette(id),
	};
}

export function clientHomeItemToMedia(item: ClientHomeItem): Media {
	const id = item.id ?? crypto.randomUUID();
	const type = item.kind === 'series' ? 'Series' : 'Movie';
	return {
		id,
		title: item.title ?? 'Unknown',
		year: 0,
		type,
		genres: [],
		rating: 0,
		synopsis: item.overview ?? item.description ?? '',
		poster: item.posterUrl,
		backdrop: item.backdropUrl,
		progress: typeof item.percent === 'number' ? item.percent / 100 : undefined,
		...hashPalette(id),
	};
}
