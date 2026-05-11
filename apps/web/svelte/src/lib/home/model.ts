import type { AuthSessionResponse } from '$lib/api/auth';
import type {
	ClientHomeItem,
	ClientHomeResponse,
	ClientHomeRow,
	LibrariesResponse,
	PlaybackRecentItem,
	PlaybackRecentResponse
} from '$lib/api/home';
import { previewBackdrop, previewPoster } from '$lib/preview/artwork';
import {
	homePreviewBackdrop,
	homePreviewContinueBackdrop,
	homePreviewHeroBackdrop,
	homePreviewHeroPoster,
	homePreviewPoster
} from '$lib/preview/home-artwork';

const PREVIEW_TRUE_VALUES = new Set(['1', 'true', 'on', 'yes']);
const PREVIEW_FALSE_VALUES = new Set(['0', 'false', 'off', 'no']);

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

export interface HomeDisplayItem {
	id: string;
	kind: string;
	mediaSourceId: string;
	title: string;
	subtitle: string;
	meta: string;
	description: string;
	posterUrl: string;
	backdropUrl: string;
	progressPercent: number;
	searchText: string;
	playMediaSourceId: string;
	isPreview: boolean;
}

export interface HomeSummary {
	libraryCount: number;
	movieCount: number;
	tvCount: number;
	inProgressCount: number;
	watchlistCount: number;
}

type LibraryList = NonNullable<LibrariesResponse['libraries']>;

export interface HomeViewModel {
	previewMode: boolean;
	usingPreviewData: boolean;
	trueEmpty: boolean;
	user: AuthSessionResponse['user'] | null;
	libraries: LibraryList;
	hero: HomeDisplayItem;
	continueItems: HomeDisplayItem[];
	movieItems: HomeDisplayItem[];
	tvItems: HomeDisplayItem[];
	watchlistItems: HomeDisplayItem[];
	summary: HomeSummary;
}

export interface BuildHomeViewModelInput {
	homePayload: ClientHomeResponse;
	playbackRecentPayload: PlaybackRecentResponse;
	librariesPayload: LibrariesResponse;
	sessionPayload: AuthSessionResponse | null;
	previewMode: boolean;
	forceEmpty: boolean;
}

interface ParsedHomeRows {
	continueItems: HomeDisplayItem[];
	movieItems: HomeDisplayItem[];
	tvItems: HomeDisplayItem[];
	recentlyAddedItems: HomeDisplayItem[];
	watchlistItems: HomeDisplayItem[];
	lookup: Map<string, HomeDisplayItem>;
}

export function createEmptyHomeViewModel(): HomeViewModel {
	return {
		previewMode: false,
		usingPreviewData: false,
		trueEmpty: true,
		user: null,
		libraries: [],
		hero: buildEmptyHero(),
		continueItems: [],
		movieItems: [],
		tvItems: [],
		watchlistItems: [],
		summary: {
			libraryCount: 0,
			movieCount: 0,
			tvCount: 0,
			inProgressCount: 0,
			watchlistCount: 0
		}
	};
}

export function resolvePreviewMode(searchParams: URLSearchParams): boolean {
	const preview = parseBooleanFlag(searchParams.get('preview'));
	return preview ?? false;
}

export function resolveForceEmptyMode(searchParams: URLSearchParams): boolean {
	const forceEmpty = parseBooleanFlag(searchParams.get('forceEmpty'));
	return forceEmpty ?? false;
}

export function buildHomeViewModel(input: BuildHomeViewModelInput): HomeViewModel {
	const previewBundle = buildPreviewBundle();
	const parsedRows = parseHomeRows(input.homePayload.rows || []);
	const recentItems = mapRecentPlaybackItems(
		input.playbackRecentPayload.recent || [],
		parsedRows.lookup
	);

	const realMovies = uniqueLimit(
		parsedRows.movieItems.length > 0
			? parsedRows.movieItems
			: parsedRows.recentlyAddedItems.filter((item) => item.kind === 'movie'),
		12
	);
	const realTV = uniqueLimit(
		parsedRows.tvItems.length > 0
			? parsedRows.tvItems
			: parsedRows.recentlyAddedItems.filter((item) => item.kind === 'series'),
		12
	);
	const realContinue = fillRailItems(
		recentItems.length > 0 ? recentItems : parsedRows.continueItems,
		[...realMovies, ...realTV, ...parsedRows.recentlyAddedItems],
		6
	);
	const realWatchlist = uniqueLimit(parsedRows.watchlistItems, 6);

	const usePreviewContinue = input.previewMode && realContinue.length === 0;
	const usePreviewMovies = input.previewMode && realMovies.length === 0;
	const usePreviewTV = input.previewMode && realTV.length === 0;
	const usePreviewWatchlist = input.previewMode && realWatchlist.length === 0;

	const continueItems = usePreviewContinue ? previewBundle.continueItems : realContinue;
	const movieItems = usePreviewMovies ? previewBundle.movieItems : realMovies;
	const tvItems = usePreviewTV ? previewBundle.tvItems : realTV;
	const watchlistItems = usePreviewWatchlist ? previewBundle.watchlistItems : realWatchlist;

	const realLibraries = input.librariesPayload.libraries || [];
	const hasRealMedia = realContinue.length > 0 || realMovies.length > 0 || realTV.length > 0;
	const shouldForceEmpty = input.forceEmpty;
	const trueEmpty = shouldForceEmpty || (!input.previewMode && realLibraries.length === 0 && !hasRealMedia);

	const usingPreviewData =
		!trueEmpty && (usePreviewContinue || usePreviewMovies || usePreviewTV || usePreviewWatchlist);

	let hero = trueEmpty
		? buildEmptyHero()
		: selectHero({
				hero: input.homePayload.hero,
				movies: movieItems,
				tv: tvItems,
				continueItems,
				previewHero: previewBundle.hero
			});

	if (!hero.id && !trueEmpty && input.previewMode) {
		hero = previewBundle.hero;
	}

	const realSummary: HomeSummary = {
		libraryCount: realLibraries.length,
		movieCount: realMovies.length,
		tvCount: realTV.length,
		inProgressCount: realContinue.length,
		watchlistCount: realWatchlist.length
	};

	const summary: HomeSummary = trueEmpty
		? {
				libraryCount: 0,
				movieCount: 0,
				tvCount: 0,
				inProgressCount: 0,
				watchlistCount: 0
			}
		: usingPreviewData
			? {
					libraryCount: Math.max(realSummary.libraryCount, 2),
					movieCount: Math.max(realSummary.movieCount, movieItems.length),
					tvCount: Math.max(realSummary.tvCount, tvItems.length),
					inProgressCount: Math.max(realSummary.inProgressCount, continueItems.length),
					watchlistCount: Math.max(realSummary.watchlistCount, watchlistItems.length)
				}
			: realSummary;

	return {
		previewMode: input.previewMode,
		usingPreviewData,
		trueEmpty,
		user: input.sessionPayload?.user || null,
		libraries: realLibraries,
		hero,
		continueItems,
		movieItems,
		tvItems,
		watchlistItems,
		summary
	};
}

export function updateHeroWithDetail(
	hero: HomeDisplayItem,
	detail: { overview?: string; playMediaSourceId?: string }
): HomeDisplayItem {
	if (!hero.id) return hero;
	const description = cleanDescription(detail.overview || hero.description);
	const playMediaSourceId = asText(detail.playMediaSourceId) || hero.playMediaSourceId;
	return {
		...hero,
		description,
		playMediaSourceId
	};
}

function parseHomeRows(rows: ClientHomeRow[]): ParsedHomeRows {
	const lookup = new Map<string, HomeDisplayItem>();
	const output: ParsedHomeRows = {
		continueItems: [],
		movieItems: [],
		tvItems: [],
		recentlyAddedItems: [],
		watchlistItems: [],
		lookup
	};

	for (const row of rows) {
		const rowID = asText(row?.id).toLowerCase();
		const mapped = Array.isArray(row?.items)
			? row.items.map((item) => mapClientHomeItem(item)).filter((item) => Boolean(item.id || item.title))
			: [];

		for (const item of mapped) {
			addLookupItem(lookup, item);
		}

		if (rowID === 'continue') output.continueItems = mapped;
		if (rowID === 'movies') output.movieItems = mapped;
		if (rowID === 'tv') output.tvItems = mapped;
		if (rowID === 'recently-added') output.recentlyAddedItems = mapped;
		if (rowID === 'watchlist') output.watchlistItems = mapped;
	}

	return output;
}

function mapClientHomeItem(item: ClientHomeItem): HomeDisplayItem {
	const id = asText(item.id || item.mediaSourceId);
	const kind = normalizeKind(asText(item.kind));
	const mediaSourceId = asText(item.mediaSourceId);
	const rawTitle = asText(item.title);
	const title = cleanDisplayTitle(rawTitle || asText(item.mediaSourceId) || 'Untitled');
	const subtitle = cleanDisplaySubtitle(asText(item.subtitle), rawTitle);
	const year = extractYear(asText(item.subtitle)) || extractYear(rawTitle);
	const meta = buildMeta(kind, subtitle, year, '');
	const posterUrl = resolvePosterUrl(kind, id, asText(item.posterUrl));
	const backdropUrl = resolveBackdropUrl(kind, id, asText(item.backdropUrl), posterUrl);
	const description = cleanDescription(asText(item.description || item.overview));
	const progressPercent = normalizeProgress(item.progressPercent ?? item.percent);
	const searchText = buildSearchText(title, subtitle, year, meta);

	return {
		id,
		kind,
		mediaSourceId,
		title,
		subtitle,
		meta,
		description,
		posterUrl,
		backdropUrl,
		progressPercent,
		searchText,
		playMediaSourceId: mediaSourceId,
		isPreview: false
	};
}

function mapRecentPlaybackItems(
	items: PlaybackRecentItem[],
	lookup: Map<string, HomeDisplayItem>
): HomeDisplayItem[] {
	return items
		.map((item) => {
			const matched = lookupRecentLookupItem(item, lookup);
			const id = asText(item.mediaSourceId);
			const fallbackTitle = cleanDisplayTitle(asText(item.name) || asText(item.relPath) || 'Resume playback');
			const title = matched?.title || fallbackTitle;
			const subtitle =
				extractEpisodeCode(`${asText(item.name)} ${asText(item.relPath)}`) ||
				extractYear(`${asText(item.name)} ${asText(item.relPath)}`) ||
				matched?.subtitle ||
				'';
			const kind = matched?.kind || normalizeKind(asText(item.kind || 'unknown'));
			const progressPercent = normalizeProgress(item.percent);
			const progressLabel = progressPercent > 0 ? `Resume from ${progressPercent}%` : '';
			const meta = buildMeta(kind, subtitle, '', progressLabel);
			const searchText = buildSearchText(title, subtitle, '', meta);

			return {
				id,
				kind,
				mediaSourceId: id,
				title,
				subtitle,
				meta,
				description: '',
				posterUrl: matched?.posterUrl || '',
				backdropUrl: matched?.backdropUrl || matched?.posterUrl || '',
				progressPercent,
				searchText,
				playMediaSourceId: id,
				isPreview: false
			};
		})
		.filter((item) => Boolean(item.id));
}

function lookupRecentLookupItem(
	item: PlaybackRecentItem,
	lookup: Map<string, HomeDisplayItem>
): HomeDisplayItem | null {
	const candidates = new Set<string>();
	candidates.add(normalizeLookupKey(asText(item.name), false));
	candidates.add(normalizeLookupKey(asText(item.name), true));

	const relPath = asText(item.relPath);
	if (relPath) {
		const segments = relPath.split(/[\\/]/).filter(Boolean);
		for (const segment of segments) {
			candidates.add(normalizeLookupKey(segment, false));
			candidates.add(normalizeLookupKey(segment, true));
		}
	}

	for (const key of candidates) {
		if (key && lookup.has(key)) return lookup.get(key) || null;
	}
	return null;
}

function selectHero({
	hero,
	movies,
	tv,
	continueItems,
	previewHero
}: {
	hero?: ClientHomeItem;
	movies: HomeDisplayItem[];
	tv: HomeDisplayItem[];
	continueItems: HomeDisplayItem[];
	previewHero: HomeDisplayItem;
}): HomeDisplayItem {
	const mappedHero = hero ? mapClientHomeItem(hero) : null;
	if (mappedHero && (mappedHero.id || mappedHero.title)) {
		return mappedHero;
	}

	const heroFromMovies =
		movies.find((item) => Boolean(item.backdropUrl || item.posterUrl)) || movies[0] || null;
	if (heroFromMovies) return heroFromMovies;

	const heroFromTV = tv.find((item) => Boolean(item.backdropUrl || item.posterUrl)) || tv[0] || null;
	if (heroFromTV) return heroFromTV;

	const heroFromContinue = continueItems[0] || null;
	if (heroFromContinue) return heroFromContinue;

	return previewHero;
}

function buildPreviewBundle(): {
	hero: HomeDisplayItem;
	continueItems: HomeDisplayItem[];
	movieItems: HomeDisplayItem[];
	tvItems: HomeDisplayItem[];
	watchlistItems: HomeDisplayItem[];
} {
	const hero = previewItem({
		id: 'hero-feature-ember-harbor',
		kind: 'movie',
		title: 'Ember Harbor',
		subtitle: '2025',
		meta: '2025 • Movie • 1h 54m',
		description:
			'When long-range navigation lights flicker across the bay, one captain is forced to choose between silence and a broadcast that could reshape the coast.',
		progressPercent: 64
	});
	hero.backdropUrl = homePreviewHeroBackdrop();
	hero.posterUrl = homePreviewHeroPoster();
	hero.playMediaSourceId = 'preview-ember-harbor';

	const continueItems = [
		previewItem({
			id: 'continue-atlas-of-dawn',
			kind: 'series',
			title: 'Atlas of Dawn',
			subtitle: 'S2 E4',
			meta: 'Resume from 68%',
			progressPercent: 68
		}),
		previewItem({
			id: 'continue-the-last-orchard',
			kind: 'movie',
			title: 'The Last Orchard',
			subtitle: '2023',
			meta: 'Resume from 42%',
			progressPercent: 42
		}),
		previewItem({
			id: 'continue-violet-signal',
			kind: 'series',
			title: 'Violet Signal',
			subtitle: 'S1 E7',
			meta: 'Resume from 58%',
			progressPercent: 58
		}),
		previewItem({
			id: 'continue-coastline',
			kind: 'series',
			title: 'Coastline',
			subtitle: 'S1 E3',
			meta: 'Resume from 52%',
			progressPercent: 52
		}),
		previewItem({
			id: 'continue-polar-night',
			kind: 'movie',
			title: 'Polar Night',
			subtitle: '2024',
			meta: 'Resume from 35%',
			progressPercent: 35
		}),
		previewItem({
			id: 'continue-return-vector',
			kind: 'series',
			title: 'Return Vector',
			subtitle: 'S3 E2',
			meta: 'Resume from 21%',
			progressPercent: 21
		})
	];
	for (const item of continueItems) {
		item.backdropUrl = homePreviewContinueBackdrop(item.title);
	}

	const movieItems = [
		previewItem({ id: 'movie-ember-harbor', kind: 'movie', title: 'Ember Harbor', subtitle: '2025' }),
		previewItem({
			id: 'movie-atlas-of-dawn',
			kind: 'movie',
			title: 'Atlas of Dawn',
			subtitle: '2024'
		}),
		previewItem({
			id: 'movie-polar-night',
			kind: 'movie',
			title: 'Polar Night',
			subtitle: '2024'
		}),
		previewItem({
			id: 'movie-night-archive',
			kind: 'movie',
			title: 'Night Archive',
			subtitle: '2022'
		}),
		previewItem({
			id: 'movie-the-last-orchard',
			kind: 'movie',
			title: 'The Last Orchard',
			subtitle: '2023'
		}),
		previewItem({
			id: 'movie-violet-signal',
			kind: 'movie',
			title: 'Violet Signal',
			subtitle: '2022'
		}),
		previewItem({ id: 'movie-broken-current', kind: 'movie', title: 'Broken Current', subtitle: '2021' }),
		previewItem({ id: 'movie-glass-canyon', kind: 'movie', title: 'Glass Canyon', subtitle: '2020' })
	];

	const tvItems = [
		previewItem({
			id: 'tv-coastline',
			kind: 'series',
			title: 'Coastline',
			subtitle: 'Latest: S2 E6'
		}),
		previewItem({
			id: 'tv-return-vector',
			kind: 'series',
			title: 'Return Vector',
			subtitle: 'Latest: S3 E2'
		}),
		previewItem({ id: 'tv-night-orbit', kind: 'series', title: 'Night Orbit', subtitle: 'Latest: S1 E4' }),
		previewItem({ id: 'tv-littoral', kind: 'series', title: 'Littoral', subtitle: 'Latest: S2 E5' }),
		previewItem({ id: 'tv-violet-signal', kind: 'series', title: 'Violet Signal', subtitle: 'Latest: S3 E1' }),
		previewItem({ id: 'tv-sunward', kind: 'series', title: 'Sunward', subtitle: 'Latest: S1 E10' }),
		previewItem({ id: 'tv-archipelago', kind: 'series', title: 'Archipelago', subtitle: 'Latest: S2 E1' })
	];
	for (const item of tvItems) {
		if (item.id === 'tv-violet-signal') {
			item.posterUrl = '/preview-art/lorivo/violet-signal-tv-poster.webp';
			item.backdropUrl = '/preview-art/lorivo/violet-signal-wide.webp';
		}
	}

	const watchlistItems = [
		previewItem({ id: 'watchlist-ember-harbor', kind: 'movie', title: 'Ember Harbor', subtitle: '2025' }),
		previewItem({ id: 'watchlist-violet-signal', kind: 'series', title: 'Violet Signal', subtitle: 'S1 E7' }),
		previewItem({ id: 'watchlist-polar-night', kind: 'movie', title: 'Polar Night', subtitle: '2024' })
	];

	return {
		hero,
		continueItems,
		movieItems,
		tvItems,
		watchlistItems
	};
}

function previewItem({
	id,
	kind,
	title,
	subtitle = '',
	meta = '',
	description = '',
	progressPercent = 0
}: {
	id: string;
	kind: string;
	title: string;
	subtitle?: string;
	meta?: string;
	description?: string;
	progressPercent?: number;
}): HomeDisplayItem {
	const cleanTitle = cleanDisplayTitle(title);
	const cleanSubtitle = cleanDisplaySubtitle(subtitle, cleanTitle);
	const year = extractYear(subtitle);
	return {
		id,
		kind,
		mediaSourceId: '',
		title: cleanTitle,
		subtitle: cleanSubtitle,
		meta: meta || buildMeta(kind, cleanSubtitle, year, ''),
		description: cleanDescription(description),
		posterUrl: homePreviewPoster(cleanTitle) || previewPoster(cleanTitle),
		backdropUrl: homePreviewBackdrop(cleanTitle) || previewBackdrop(cleanTitle),
		progressPercent: normalizeProgress(progressPercent),
		searchText: buildSearchText(cleanTitle, cleanSubtitle, year, meta),
		playMediaSourceId: '',
		isPreview: true
	};
}

function buildEmptyHero(): HomeDisplayItem {
	return {
		id: 'empty',
		kind: 'empty',
		mediaSourceId: '',
		title: 'Add your first library',
		subtitle: '',
		meta: 'Setup',
		description:
			'Choose a Movies or TV folder and Lorivo will build your personal streaming home.',
		posterUrl: '',
		backdropUrl: '',
		progressPercent: 0,
		searchText: 'add your first library',
		playMediaSourceId: '',
		isPreview: false
	};
}

function fillRailItems(
	primary: HomeDisplayItem[],
	fallback: HomeDisplayItem[],
	limit: number
): HomeDisplayItem[] {
	const output: HomeDisplayItem[] = [];
	const seen = new Set<string>();

	const append = (items: HomeDisplayItem[]): void => {
		for (const item of items) {
			const key = uniqueKey(item);
			if (!key || seen.has(key)) continue;
			seen.add(key);
			output.push(item);
			if (output.length >= limit) return;
		}
	};

	append(primary);
	if (output.length < limit) append(fallback);
	return output;
}

function uniqueLimit(items: HomeDisplayItem[], limit: number): HomeDisplayItem[] {
	const output: HomeDisplayItem[] = [];
	const seen = new Set<string>();
	for (const item of items) {
		const key = uniqueKey(item);
		if (!key || seen.has(key)) continue;
		seen.add(key);
		output.push(item);
		if (output.length >= limit) break;
	}
	return output;
}

function uniqueKey(item: HomeDisplayItem): string {
	const titleKey = normalizeLookupKey(item.title, false);
	if (titleKey) return `${item.kind}:${titleKey}`;
	if (item.id) return `${item.kind}:${item.id}`;
	return '';
}

function normalizeLookupKey(value: string, includeYear: boolean): string {
	const text = asText(value);
	if (!text) return '';
	const cleanedTitle = cleanDisplayTitle(text);
	const year = includeYear ? extractYear(text) : '';
	const base = [cleanedTitle, year].filter(Boolean).join(' ').trim();
	return base
		.toLowerCase()
		.replace(/[^a-z0-9\s']/g, ' ')
		.replace(/\b(the|a|an)\b/g, ' ')
		.replace(/\s+/g, ' ')
		.trim();
}

function addLookupItem(lookup: Map<string, HomeDisplayItem>, item: HomeDisplayItem): void {
	const primary = normalizeLookupKey(item.title, false);
	const withYear = normalizeLookupKey(`${item.title} ${item.subtitle}`, true);
	const titleAndSubtitle = normalizeLookupKey(`${item.title} ${item.subtitle}`, false);
	for (const key of [primary, withYear, titleAndSubtitle]) {
		if (!key || lookup.has(key)) continue;
		lookup.set(key, item);
	}
}

function normalizeKind(rawKind: string): string {
	const value = rawKind.toLowerCase();
	if (value === 'movie' || value === 'movies') return 'movie';
	if (value === 'series' || value === 'tv' || value === 'show') return 'series';
	if (value === 'episode') return 'episode';
	if (value === 'empty') return 'empty';
	return value || 'unknown';
}

function resolvePosterUrl(kind: string, id: string, provided: string): string {
	const direct = asText(provided);
	if (direct) return direct;
	if (!isArtworkEntity(kind) || !id) return '';
	return `/api/artwork/${encodeURIComponent(kind)}/${encodeURIComponent(id)}?style=neutral&type=poster`;
}

function resolveBackdropUrl(kind: string, id: string, provided: string, posterUrl: string): string {
	const direct = asText(provided);
	if (direct) return direct;
	if (!isArtworkEntity(kind) || !id) return posterUrl;
	return `/api/artwork/${encodeURIComponent(kind)}/${encodeURIComponent(id)}?style=neutral&type=backdrop`;
}

function isArtworkEntity(kind: string): boolean {
	return kind === 'movie' || kind === 'series';
}

function parseBooleanFlag(value: string | null): boolean | undefined {
	const trimmed = asText(value).toLowerCase();
	if (!trimmed) return undefined;
	if (PREVIEW_TRUE_VALUES.has(trimmed)) return true;
	if (PREVIEW_FALSE_VALUES.has(trimmed)) return false;
	return undefined;
}

function normalizeProgress(value: unknown): number {
	const numberValue = Number(value ?? 0);
	if (!Number.isFinite(numberValue) || numberValue <= 0) return 0;
	if (numberValue <= 1) return Math.round(Math.max(0, Math.min(1, numberValue)) * 100);
	return Math.round(Math.max(0, Math.min(100, numberValue)));
}

function buildMeta(kind: string, subtitle: string, year: string, suffix: string): string {
	const parts: string[] = [];
	if (subtitle) parts.push(subtitle);
	if (!subtitle && year) parts.push(year);
	const label = kindLabel(kind);
	if (label && !parts.includes(label)) parts.push(label);
	if (suffix) parts.push(suffix);
	return parts.join(' - ');
}

function kindLabel(kind: string): string {
	if (kind === 'movie') return 'Movie';
	if (kind === 'series') return 'TV';
	if (kind === 'episode') return 'Episode';
	if (kind === 'empty') return 'Setup';
	return '';
}

function buildSearchText(title: string, subtitle: string, year: string, meta: string): string {
	return [title, subtitle, year, meta]
		.filter(Boolean)
		.join(' ')
		.toLowerCase()
		.replace(/\s+/g, ' ')
		.trim();
}

function asText(value: unknown): string {
	return String(value ?? '').trim();
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

function cleanDisplaySubtitle(rawSubtitle: string, fallbackTitle: string): string {
	const subtitle = asText(rawSubtitle);
	if (!subtitle) return '';
	const episodeCode = extractEpisodeCode(`${fallbackTitle} ${subtitle}`);
	if (episodeCode) return episodeCode;
	const year = extractYear(subtitle);
	if (year) return year;
	if (/resume from /i.test(subtitle)) return '';
	return subtitle.replace(/\s+/g, ' ').trim();
}

function extractYear(value: string): string {
	const match = asText(value).match(/\b(19|20)\d{2}\b/);
	return match ? match[0] : '';
}

function extractEpisodeCode(value: string): string {
	const match = asText(value).match(/\bS(\d{1,2})E(\d{1,2})(?:E(\d{1,2}))?\b/i);
	if (!match) return '';
	const season = Number(match[1]);
	const episode = Number(match[2]);
	const end = Number(match[3] || 0);
	if (!Number.isFinite(season) || !Number.isFinite(episode)) return '';
	if (end && end !== episode) return `S${season} E${episode}-${end}`;
	return `S${season} E${episode}`;
}

function cleanDescription(value: string): string {
	const text = asText(value).replace(/\s+/g, ' ').trim();
	if (!text) return '';
	if (text.length <= 220) return text;
	return `${text.slice(0, 217).trimEnd()}...`;
}

function escapeRegex(value: string): string {
	return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}
