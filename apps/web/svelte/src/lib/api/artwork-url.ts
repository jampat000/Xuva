import type { Media } from '$lib/mock-data';

/**
 * The art types the backend /api/artwork endpoint accepts. Map directly to
 * Media field names (`poster`, `backdrop`, `logo`, `thumbnail`, `banner`).
 */
export type ArtType = 'poster' | 'backdrop' | 'thumbnail' | 'logo' | 'banner';

/**
 * Resolve the (kind, id) tuple the artwork endpoint expects for a Media row.
 *
 * Continue-watching cards arrive keyed by media_source_id but expose the
 * parent movie/series identity via parentId/parentKind — use those when
 * present so the proxy can find the right metadata record.
 */
function artworkRoute(media: Media): { kind: 'movie' | 'series'; id: string } | null {
	const parentKind = media.parentKind?.toLowerCase();
	const parentId = media.parentId;
	if ((parentKind === 'movie' || parentKind === 'series') && parentId) {
		return { kind: parentKind, id: parentId };
	}
	if (media.type === 'Movie') return { kind: 'movie', id: media.id };
	if (media.type === 'Series') return { kind: 'series', id: media.id };
	return null; // Live / unknown — no metadata-backed artwork pipeline.
}

/**
 * Build a sized-artwork URL that goes through the backend resize proxy. The
 * proxy caches each (sourceURL, width) pair on disk and serves with immutable
 * Cache-Control, so repeat hits at the same density are free.
 *
 * Returns null when the media row can't be mapped to (kind, id) — callers
 * should fall back to the raw URL from the API response (`media.poster`,
 * `media.backdrop`, etc.) in that case so the user still sees a picture.
 *
 * `width` is the CSS width in pixels at the **rendered** density you want.
 * Pass `width * devicePixelRatio` for retina-sharp output.
 */
export function sizedArtworkUrl(media: Media, type: ArtType, width: number): string | null {
	const route = artworkRoute(media);
	if (!route) return null;
	const safeId = encodeURIComponent(route.id);
	return `/api/artwork/${route.kind}/${safeId}?type=${type}&w=${width}`;
}

/**
 * Best-of-two helper: prefer the proxy URL when available, otherwise return
 * the raw URL the caller already has. Components use this so they never have
 * to write the conditional themselves.
 */
export function artworkSrc(
	media: Media,
	type: ArtType,
	width: number,
	rawUrl?: string,
): string | undefined {
	return sizedArtworkUrl(media, type, width) ?? rawUrl;
}

/**
 * Build a `srcset` attribute with the requested width and its 2× variant.
 * Use alongside `sizes` for retina sharpness without wasted bandwidth on
 * standard-density screens.
 */
export function artworkSrcset(media: Media, type: ArtType, width: number): string | undefined {
	const url1x = sizedArtworkUrl(media, type, width);
	if (!url1x) return undefined;
	const url2x = sizedArtworkUrl(media, type, Math.min(width * 2, 1024));
	if (!url2x || url2x === url1x) return undefined;
	return `${url1x} 1x, ${url2x} 2x`;
}

/**
 * URL of the backend blurhash endpoint for a Media row. Returns JSON
 * shaped `{ hash: string }`. An empty string means the backend hasn't
 * computed one yet (frontend should fall back to the palette gradient).
 *
 * Returns null when (kind, id) can't be resolved — same fallback rules as
 * sizedArtworkUrl.
 */
export function blurhashUrl(media: Media, type: ArtType): string | null {
	const url = sizedArtworkUrl(media, type, 32);
	if (!url) return null;
	// Append meta=1 to swap image bytes for a JSON envelope.
	return url + '&meta=1';
}

const _blurhashMemoryCache = new Map<string, string>();
const _blurhashInflight = new Map<string, Promise<string>>();

/**
 * Fetch the backend blurhash for a Media row, with per-tab memoization to
 * avoid duplicate fetches when the same poster appears in multiple grids
 * (e.g. on the home page rows). Returns empty string on any failure; the
 * frontend treats that as "no blurhash, use the palette gradient."
 */
export async function fetchBlurhash(media: Media, type: ArtType): Promise<string> {
	const url = blurhashUrl(media, type);
	if (!url) return '';
	const cached = _blurhashMemoryCache.get(url);
	if (cached !== undefined) return cached;
	const inflight = _blurhashInflight.get(url);
	if (inflight) return inflight;
	const promise = (async () => {
		try {
			const resp = await fetch(url, { credentials: 'include' });
			if (!resp.ok) return '';
			const body = await resp.json() as { hash?: string };
			const hash = String(body.hash ?? '');
			_blurhashMemoryCache.set(url, hash);
			return hash;
		} catch {
			return '';
		} finally {
			_blurhashInflight.delete(url);
		}
	})();
	_blurhashInflight.set(url, promise);
	return promise;
}
