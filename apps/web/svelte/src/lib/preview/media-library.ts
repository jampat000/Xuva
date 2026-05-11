import type { MovieListItem, SeriesListItem } from '$lib/api/browse';
import { previewBackdrop, previewPoster } from '$lib/preview/artwork';

const LOCKED = '/preview-art/lorivo/locked';
const LOCAL = '/preview-art/lorivo';

export interface PreviewMovieFixture {
	id: string;
	title: string;
	year: number;
	overview: string;
	posterUrl: string;
	backdropUrl: string;
	versionCount: number;
	needsReview?: boolean;
}

export interface PreviewSeriesFixture {
	id: string;
	title: string;
	year: number;
	overview: string;
	posterUrl: string;
	backdropUrl: string;
	seasonCount: number;
	episodeCount: number;
}

export const PREVIEW_MOVIES: PreviewMovieFixture[] = [
	{
		id: 'preview-movie-dune-part-two',
		title: 'Dune: Part Two',
		year: 2024,
		overview:
			'Paul Atreides unites with the Fremen and wages war against the conspirators who destroyed his family, torn between love and a prophecy that could ignite a galaxy-spanning holy war.',
		posterUrl: `${LOCKED}/mv-topgun.jpg`,
		backdropUrl: `${LOCKED}/cw-dune.png`,
		versionCount: 2
	},
	{
		id: 'preview-movie-mufasa',
		title: 'Mufasa: The Lion King',
		year: 2024,
		overview: 'A young lion rises from exile toward the legacy that will shape the Pride Lands.',
		posterUrl: `${LOCKED}/mv-mufasa.jpg`,
		backdropUrl: `${LOCKED}/mv-mufasa.png`,
		versionCount: 1
	},
	{
		id: 'preview-movie-twisters',
		title: 'Twisters',
		year: 2024,
		overview: 'Storm chasers push toward a dangerous new weather system as the sky turns unpredictable.',
		posterUrl: `${LOCKED}/mv-twisters.jpg`,
		backdropUrl: `${LOCKED}/mv-twisters.jpg`,
		versionCount: 1
	},
	{
		id: 'preview-movie-quiet-place-day-one',
		title: 'A Quiet Place: Day One',
		year: 2024,
		overview: 'A city falls silent as survivors learn how to move through the first hours of an invasion.',
		posterUrl: `${LOCKED}/mv-quietplace-dayone.jpg`,
		backdropUrl: `${LOCKED}/mv-quietplace-dayone.jpg`,
		versionCount: 1
	},
	{
		id: 'preview-movie-fall-guy',
		title: 'The Fall Guy',
		year: 2024,
		overview: 'A stunt performer is pulled into a missing-person mystery behind the scenes of a massive production.',
		posterUrl: `${LOCKED}/mv-fallguy-real.jpg`,
		backdropUrl: `${LOCKED}/mv-kingdom.png`,
		versionCount: 2
	},
	{
		id: 'preview-movie-kingdom-apes',
		title: 'Kingdom of the Planet of the Apes',
		year: 2024,
		overview: 'A young ape crosses a transformed world where memory, power, and survival collide.',
		posterUrl: `${LOCKED}/mv-civilwar.jpg`,
		backdropUrl: `${LOCKED}/mv-civilwar.png`,
		versionCount: 1
	},
	{
		id: 'preview-movie-civil-war',
		title: 'Civil War',
		year: 2024,
		overview: 'A group of journalists races across a fractured country toward the center of a national crisis.',
		posterUrl: `${LOCKED}/mv-dune2.jpg`,
		backdropUrl: `${LOCKED}/mv-dune2.png`,
		versionCount: 1
	},
	{
		id: 'preview-movie-top-gun-maverick',
		title: 'Top Gun: Maverick',
		year: 2022,
		overview: 'An elite pilot returns to train a new generation for a mission that demands everything.',
		posterUrl: `${LOCKED}/mv-deadpool.jpg`,
		backdropUrl: `${LOCKED}/mv-deadpool.png`,
		versionCount: 1
	},
	{
		id: 'preview-movie-deadpool-wolverine',
		title: 'Deadpool & Wolverine',
		year: 2024,
		overview: 'Two volatile heroes collide across timelines, grudges, and a mission neither can ignore.',
		posterUrl: `${LOCKED}/mv-wildrobot.jpg`,
		backdropUrl: `${LOCKED}/mv-wildrobot.png`,
		versionCount: 1
	},
	{
		id: 'preview-movie-wild-robot',
		title: 'The Wild Robot',
		year: 2024,
		overview: 'A stranded robot learns to survive and care for a wild island that slowly becomes home.',
		posterUrl: previewPoster('The Wild Robot'),
		backdropUrl: previewBackdrop('The Wild Robot'),
		versionCount: 1
	},
	{
		id: 'preview-movie-atlas-of-dawn',
		title: 'Atlas of Dawn',
		year: 2024,
		overview: 'A crew follows a fractured star map toward a discovery buried in morning light.',
		posterUrl: `${LOCAL}/atlas-of-dawn-poster.webp`,
		backdropUrl: `${LOCAL}/atlas-of-dawn-wide.webp`,
		versionCount: 2
	},
	{
		id: 'preview-movie-ember-harbor',
		title: 'Ember Harbor',
		year: 2025,
		overview:
			'When long-range navigation lights flicker across the bay, one captain is forced to choose between silence and a broadcast that could reshape the coast.',
		posterUrl: `${LOCAL}/ember-harbor-poster.webp`,
		backdropUrl: `${LOCAL}/ember-harbor-backdrop.webp`,
		versionCount: 2
	}
];

export const PREVIEW_SERIES: PreviewSeriesFixture[] = [
	{
		id: 'preview-tv-violet-signal',
		title: 'Violet Signal',
		year: 2025,
		overview:
			'A late-night observatory feed reveals repeating anomalies, and a fragmented crew races to decode what returns with every tide.',
		posterUrl: `${LOCAL}/violet-signal-tv-poster.webp`,
		backdropUrl: `${LOCAL}/violet-signal-wide.webp`,
		seasonCount: 2,
		episodeCount: 5
	},
	{
		id: 'preview-tv-penguin',
		title: 'The Penguin',
		year: 2024,
		overview: 'A crime saga follows a ruthless ascent through the city after the old order fractures.',
		posterUrl: `${LOCKED}/tv-penguin.jpg`,
		backdropUrl: `${LOCKED}/tv-penguin.png`,
		seasonCount: 1,
		episodeCount: 8
	},
	{
		id: 'preview-tv-house-of-the-dragon',
		title: 'House of the Dragon',
		year: 2024,
		overview: 'A divided dynasty draws closer to war as rival claims burn across the realm.',
		posterUrl: `${LOCKED}/tv-hod.jpg`,
		backdropUrl: `${LOCKED}/tv-hod.png`,
		seasonCount: 2,
		episodeCount: 18
	},
	{
		id: 'preview-tv-the-boys',
		title: 'The Boys',
		year: 2024,
		overview: 'A brutal power struggle escalates as corrupt heroes and their hunters cross new lines.',
		posterUrl: `${LOCKED}/tv-boys.jpg`,
		backdropUrl: `${LOCKED}/tv-boys.png`,
		seasonCount: 4,
		episodeCount: 32
	},
	{
		id: 'preview-tv-slow-horses',
		title: 'Slow Horses',
		year: 2024,
		overview: 'A misfit intelligence unit stumbles into another crisis with consequences far above its pay grade.',
		posterUrl: `${LOCKED}/tv-slowhorses.jpg`,
		backdropUrl: `${LOCKED}/tv-slowhorses.png`,
		seasonCount: 4,
		episodeCount: 24
	},
	{
		id: 'preview-tv-reacher',
		title: 'Reacher',
		year: 2025,
		overview: 'A wandering investigator follows a new trail of violence, leverage, and buried motive.',
		posterUrl: `${LOCKED}/tv-reacher.jpg`,
		backdropUrl: `${LOCKED}/tv-reacher.png`,
		seasonCount: 3,
		episodeCount: 24
	},
	{
		id: 'preview-tv-severance',
		title: 'Severance',
		year: 2025,
		overview: 'Office memory and private life fracture further as workers uncover the shape of the system.',
		posterUrl: `${LOCKED}/tv-severance.jpg`,
		backdropUrl: `${LOCKED}/tv-severance.png`,
		seasonCount: 2,
		episodeCount: 19
	},
	{
		id: 'preview-tv-silo',
		title: 'Silo',
		year: 2024,
		overview: 'A sealed society faces truths hidden below its routines, rules, and buried machinery.',
		posterUrl: `${LOCKED}/tv-silo.jpg`,
		backdropUrl: `${LOCKED}/tv-silo.png`,
		seasonCount: 2,
		episodeCount: 20
	},
	{
		id: 'preview-tv-three-body-problem',
		title: '3 Body Problem',
		year: 2024,
		overview: 'A scientific mystery unfolds across decades as humanity confronts an impossible signal.',
		posterUrl: `${LOCKED}/tv-3body.jpg`,
		backdropUrl: `${LOCKED}/tv-3body.png`,
		seasonCount: 1,
		episodeCount: 8
	},
	{
		id: 'preview-tv-monarch',
		title: 'Monarch: Legacy of Monsters',
		year: 2023,
		overview: 'A family history opens into a global record of massive creatures and hidden institutions.',
		posterUrl: `${LOCKED}/tv-monarch.jpg`,
		backdropUrl: `${LOCKED}/tv-monarch.png`,
		seasonCount: 1,
		episodeCount: 10
	}
];

export function previewMovieRows(): MovieListItem[] {
	return PREVIEW_MOVIES.map((item) => ({
		id: item.id,
		title: item.title,
		year: item.year,
		versionCount: item.versionCount,
		needsReview: Boolean(item.needsReview),
		metadata: {
			title: item.title,
			year: item.year,
			overview: item.overview,
			posterUrl: item.posterUrl,
			backdropUrl: item.backdropUrl
		}
	}));
}

export function previewSeriesRows(): SeriesListItem[] {
	return PREVIEW_SERIES.map((item) => ({
		id: item.id,
		title: item.title,
		seasonCount: item.seasonCount,
		episodeCount: item.episodeCount,
		metadata: {
			title: item.title,
			year: item.year,
			overview: item.overview,
			posterUrl: item.posterUrl,
			backdropUrl: item.backdropUrl
		}
	}));
}

export function getPreviewMovieFixture(id: string): PreviewMovieFixture {
	const normalized = id === 'preview' ? 'preview-movie-dune-part-two' : id;
	return PREVIEW_MOVIES.find((item) => item.id === normalized) || fallbackMovieFixture(normalized);
}

export function getPreviewSeriesFixture(id: string): PreviewSeriesFixture {
	const normalized = id === 'preview' ? 'preview-tv-penguin' : id;
	return PREVIEW_SERIES.find((item) => item.id === normalized) || fallbackSeriesFixture(normalized);
}

function fallbackMovieFixture(id: string): PreviewMovieFixture {
	const title = cleanPreviewTitle(id, 'preview-movie-', 'Movie Preview');
	return {
		id: id || 'preview-movie',
		title,
		year: 2025,
		overview: 'This preview movie is available for local playback testing.',
		posterUrl: previewPoster(title),
		backdropUrl: previewBackdrop(title),
		versionCount: 1
	};
}

function fallbackSeriesFixture(id: string): PreviewSeriesFixture {
	const title = cleanPreviewTitle(id, 'preview-tv-', 'TV Preview');
	return {
		id: id || 'preview-tv',
		title,
		year: 2025,
		overview: 'This preview series is available for local playback testing.',
		posterUrl: previewPoster(title),
		backdropUrl: previewBackdrop(title),
		seasonCount: 1,
		episodeCount: 3
	};
}

function cleanPreviewTitle(value: string, prefix: string, fallback: string): string {
	const raw = String(value || '').replace(prefix, '').replace(/-/g, ' ').trim();
	if (!raw) return fallback;
	return raw.replace(/\b\w/g, (letter) => letter.toUpperCase());
}
