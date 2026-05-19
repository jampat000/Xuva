/**
 * Client capability auto-detection (issue #64).
 *
 * Builds a capability report for the current browser by probing
 * canPlayType() results and media queries. The result is cached in memory
 * so repeated calls (player setup, route pre-flight) are free.
 */

export interface ClientCapabilities {
	containers: string[];
	videoCodecs: string[];
	audioCodecs: string[];
	subtitleCodecs: string[];
	maxVideoBitDepth: number;
	maxVideoFrameRate: number;
	supportsHdr: boolean;
	supportsDolbyVision: boolean;
	supportsHls: boolean;
}

let _cached: ClientCapabilities | null = null;

/**
 * Build (or return cached) capability report for this browser session.
 * Must be called in a browser context (document must be available).
 */
export function buildCapabilityReport(): ClientCapabilities {
	if (_cached) return _cached;

	const video = document.createElement('video');
	const can = (type: string): boolean => {
		const r = video.canPlayType(type);
		return r === 'probably' || r === 'maybe';
	};

	// ── Containers ────────────────────────────────────────────────────────────
	const containers: string[] = [];
	if (can('video/mp4')) containers.push('mp4');
	if (can('video/webm')) containers.push('webm');
	if (can('video/ogg')) containers.push('ogg');
	if (can('video/mp2t') || can('video/mpegts')) containers.push('mpegts');

	// ── Video codecs ──────────────────────────────────────────────────────────
	const videoCodecs: string[] = [];
	if (can('video/mp4; codecs="avc1.42E01E"') || can('video/mp4; codecs="avc1.4D401F"')) {
		videoCodecs.push('h264');
	}
	if (
		can('video/mp4; codecs="hvc1.1.6.L93.90"') ||
		can('video/mp4; codecs="hev1.1.6.L93.90"') ||
		can('video/mp4; codecs="hvc1"') ||
		can('video/mp4; codecs="hev1"')
	) {
		videoCodecs.push('hevc');
	}
	if (can('video/webm; codecs="vp9"') || can('video/mp4; codecs="vp09.00.10.08"')) {
		videoCodecs.push('vp9');
	}
	if (
		can('video/mp4; codecs="av01.0.05M.08"') ||
		can('video/webm; codecs="av01.0.05M.08"') ||
		can('video/mp4; codecs="av01"')
	) {
		videoCodecs.push('av1');
	}

	// ── Audio codecs ──────────────────────────────────────────────────────────
	const audioCodecs: string[] = [];
	if (can('audio/mp4; codecs="mp4a.40.2"') || can('audio/aac')) audioCodecs.push('aac');
	if (can('audio/mp4; codecs="ac-3"') || can('audio/mp4; codecs="ac3"')) audioCodecs.push('ac3');
	if (can('audio/mp4; codecs="ec-3"') || can('audio/mp4; codecs="ec3"')) audioCodecs.push('eac3');
	if (can('audio/ogg; codecs="opus"') || can('audio/webm; codecs="opus"')) audioCodecs.push('opus');
	if (can('audio/mpeg')) audioCodecs.push('mp3');
	if (can('audio/flac')) audioCodecs.push('flac');

	// ── Subtitle formats ──────────────────────────────────────────────────────
	// All modern browsers support text tracks rendered by JS; assume WebVTT/SRT.
	const subtitleCodecs = ['webvtt', 'srt'];

	// ── HDR / bit-depth detection ─────────────────────────────────────────────
	// Best-effort: media queries may not be available on all platforms.
	let supportsHdr = false;
	try {
		supportsHdr =
			window.matchMedia('(dynamic-range: high)').matches ||
			window.matchMedia('(color-gamut: p3)').matches ||
			screen.colorDepth >= 30;
	} catch {
		// matchMedia unavailable (SSR, test environment)
	}
	const maxVideoBitDepth = supportsHdr ? 10 : 8;

	// ── HLS support ───────────────────────────────────────────────────────────
	// Native (Safari) or via hls.js (which we bundle for Chrome/Firefox).
	const supportsHlsNative =
		can('application/x-mpegURL') || can('application/vnd.apple.mpegurl');
	// hls.js is always bundled in the web player, so HLS is always available.
	const supportsHls = true || supportsHlsNative;

	_cached = {
		containers: containers.length > 0 ? containers : ['mp4'],
		videoCodecs: videoCodecs.length > 0 ? videoCodecs : ['h264'],
		audioCodecs: audioCodecs.length > 0 ? audioCodecs : ['aac'],
		subtitleCodecs,
		maxVideoBitDepth,
		maxVideoFrameRate: 60,
		supportsHdr,
		supportsDolbyVision: false, // No reliable browser detection available yet
		supportsHls,
	};

	return _cached;
}

/** Clear the cached capability report (useful for testing). */
export function clearCapabilityCache(): void {
	_cached = null;
}
