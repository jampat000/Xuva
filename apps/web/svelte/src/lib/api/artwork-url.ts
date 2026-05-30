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

// Mirror the server's maxResizeWidth (see server/internal/api/image_resize.go).
// Anything above this gets clamped server-side, so emitting a 3x descriptor
// past it wastes a srcset slot. Bumped alongside the server cap so 4K screens
// at 2-3x DPR can request a sharp poster.
const ARTWORK_MAX_WIDTH = 2048;

/**
 * Build a `srcset` attribute with 1x / 2x / 3x descriptors. The browser picks
 * the variant whose pixel width best matches the rendered CSS size × the
 * device's pixel ratio. Three rungs cover the realistic range:
 *
 *   - 1x: standard-density desktop monitors (DPR=1)
 *   - 2x: most retina laptops and modern phones (DPR=2)
 *   - 3x: high-density phones and 4K monitors at 200% scaling (DPR=3)
 *
 * Each rung is capped at ARTWORK_MAX_WIDTH because the proxy clamps above
 * that. Duplicate rungs (e.g. 2x and 3x both hit the cap) are deduped so we
 * don't ship a srcset with two identical URLs at different descriptors.
 */
export function artworkSrcset(media: Media, type: ArtType, width: number): string | undefined {
	const url1x = sizedArtworkUrl(media, type, width);
	if (!url1x) return undefined;
	const w2 = Math.min(width * 2, ARTWORK_MAX_WIDTH);
	const w3 = Math.min(width * 3, ARTWORK_MAX_WIDTH);
	const url2x = sizedArtworkUrl(media, type, w2);
	const url3x = w3 > w2 ? sizedArtworkUrl(media, type, w3) : undefined;
	const parts: string[] = [`${url1x} 1x`];
	if (url2x && url2x !== url1x) parts.push(`${url2x} 2x`);
	if (url3x && url3x !== url2x && url3x !== url1x) parts.push(`${url3x} 3x`);
	return parts.length > 1 ? parts.join(', ') : undefined;
}
