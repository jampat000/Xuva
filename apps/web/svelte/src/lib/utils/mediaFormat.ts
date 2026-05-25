/**
 * Shared formatters for media technical details — used by movie + TV detail
 * pages and anywhere else we surface codec/track info to the user.
 *
 * Keep these display-only: they should never affect API requests or playback
 * decisions, only the human-readable label.
 */

import type { ProbeTrack } from '$lib/api/details';

export function formatResolution(w?: number, h?: number): string {
	if (!h) return '';
	if (h >= 2100) return '4K';
	if (h >= 1000) return '1080p';
	if (h >= 700) return '720p';
	if (h >= 400) return '480p';
	return `${w}×${h}`;
}

export function formatBitrate(bps?: number): string {
	if (!bps || bps <= 0) return '';
	if (bps >= 1_000_000) return `${(bps / 1_000_000).toFixed(1)} Mbps`;
	if (bps >= 1_000) return `${Math.round(bps / 1_000)} kbps`;
	return `${bps} bps`;
}

export function formatChannels(n?: number): string {
	if (!n || n <= 0) return '';
	if (n === 1) return 'Mono';
	if (n === 2) return 'Stereo';
	if (n === 6) return '5.1';
	if (n === 8) return '7.1';
	return `${n}ch`;
}

const codecLabels: Record<string, string> = {
	h264: 'H.264', hevc: 'HEVC', av1: 'AV1', vp9: 'VP9', mpeg4: 'MPEG-4',
	aac: 'AAC', ac3: 'AC3', eac3: 'E-AC3', dts: 'DTS', truehd: 'TrueHD',
	flac: 'FLAC', mp3: 'MP3', opus: 'Opus', vorbis: 'Vorbis', alac: 'ALAC',
	pgs: 'PGS', srt: 'SRT', subrip: 'SRT', webvtt: 'WebVTT', vtt: 'WebVTT',
	ass: 'ASS', ssa: 'SSA', mov_text: 'MOV Text',
	hdmv_pgs_subtitle: 'PGS', dvd_subtitle: 'VobSub', dvb_subtitle: 'DVB',
};

export function formatCodec(codec?: string): string {
	if (!codec) return '';
	const lower = codec.toLowerCase();
	return codecLabels[lower] ?? codec.toUpperCase();
}

const languageLabels: Record<string, string> = {
	en: 'English', eng: 'English',
	es: 'Spanish', spa: 'Spanish',
	fr: 'French', fre: 'French', fra: 'French',
	de: 'German', ger: 'German', deu: 'German',
	it: 'Italian', ita: 'Italian',
	ja: 'Japanese', jpn: 'Japanese',
	ko: 'Korean', kor: 'Korean',
	zh: 'Chinese', chi: 'Chinese', zho: 'Chinese',
	pt: 'Portuguese', por: 'Portuguese',
	ru: 'Russian', rus: 'Russian',
	hi: 'Hindi', hin: 'Hindi',
	ar: 'Arabic', ara: 'Arabic',
	nl: 'Dutch', dut: 'Dutch', nld: 'Dutch',
	sv: 'Swedish', swe: 'Swedish',
	no: 'Norwegian', nor: 'Norwegian',
	da: 'Danish', dan: 'Danish',
	fi: 'Finnish', fin: 'Finnish',
	pl: 'Polish', pol: 'Polish',
	tr: 'Turkish', tur: 'Turkish',
	he: 'Hebrew', heb: 'Hebrew',
	und: 'Unknown',
};

export function formatLanguage(code?: string): string {
	if (!code) return '';
	return languageLabels[code.toLowerCase()] ?? code.toUpperCase();
}

export function formatFileSize(bytes?: number): string {
	if (!bytes) return '';
	if (bytes >= 1_073_741_824) return `${(bytes / 1_073_741_824).toFixed(1)} GB`;
	if (bytes >= 1_048_576) return `${Math.round(bytes / 1_048_576)} MB`;
	return `${bytes} B`;
}

/**
 * One-line audio summary from the FIRST track, e.g. "AC3 5.1" or "AAC Stereo".
 * Used in the File Info pills row where we don't want a full track listing.
 */
export function audioSummary(tracks: ProbeTrack[]): string {
	const first = tracks?.[0];
	if (!first) return '';
	return [formatCodec(first.codec), formatChannels(first.channels)].filter(Boolean).join(' ');
}

// ─── Playability badges ────────────────────────────────────────────────────
// Translates a playback decision mode (server-side enum from
// internal/playback/decision.go) into a user-facing label, blurb, and RAG
// colour so movie/TV detail pages can show "Plays instantly" /
// "Will repackage in ~30s" / "Needs full transcode" per version.
//
// RAG semantics match the Library Codecs panel on the dashboard:
//   green  = no server work (direct play)
//   amber  = container repackage only (remux)
//   amber+ = video copy + audio re-encode (a bit more work)
//   red    = full video re-encode required
//   muted  = blocked / deferred (file not probed, policy block, etc.)
export type PlayabilityBadge = {
	label: string;
	blurb: string;
	color: string;       // OKLCH literal for inline styles
	tone: 'green' | 'amber' | 'red' | 'muted';
};

export function playabilityBadge(mode?: string): PlayabilityBadge {
	const m = (mode || '').toLowerCase();
	if (m === 'direct play') {
		return {
			label: 'Plays instantly',
			blurb: 'No server work needed — browser decodes the file natively.',
			color: 'oklch(0.78 0.22 145)',
			tone: 'green',
		};
	}
	if (m === 'remux') {
		return {
			label: 'Fast repackage',
			blurb: 'Server rewraps the container (≈30s for a 2h film). Video stream is untouched.',
			color: 'oklch(0.85 0.22 75)',
			tone: 'amber',
		};
	}
	if (m === 'audio transcode') {
		return {
			label: 'Audio transcode',
			blurb: 'Video stream passes through, only audio is re-encoded (≈2–3 min for a 2h film).',
			color: 'oklch(0.85 0.22 75)',
			tone: 'amber',
		};
	}
	if (m === 'video transcode' || m === 'adaptive stream') {
		return {
			label: 'Needs full transcode',
			blurb: 'No browser can decode this directly. The server has to re-encode the video — slow and CPU-heavy. Convert once with HandBrake for instant playback later.',
			color: 'oklch(0.68 0.26 22)',
			tone: 'red',
		};
	}
	if (m === 'subtitle burn') {
		return {
			label: 'Subtitle burn-in',
			blurb: 'The selected subtitle track is image-based — playing it requires re-encoding the video with subs baked in.',
			color: 'oklch(0.68 0.26 22)',
			tone: 'red',
		};
	}
	if (m === 'decision deferred') {
		return {
			label: 'Not yet analysed',
			blurb: 'Xuva needs to inspect this file once before it can pick the best playback path.',
			color: 'oklch(0.65 0.02 280)',
			tone: 'muted',
		};
	}
	return {
		label: 'Unknown',
		blurb: 'Playback path not determined.',
		color: 'oklch(0.65 0.02 280)',
		tone: 'muted',
	};
}

// Short 1-2 word label for use in compact matrix cells.
export function playabilityShortLabel(mode?: string): string {
	const m = (mode || '').toLowerCase();
	if (m === 'direct play') return 'Direct';
	if (m === 'remux') return 'Remux';
	if (m === 'audio transcode') return 'Audio';
	if (m === 'video transcode' || m === 'adaptive stream') return 'Transcode';
	if (m === 'subtitle burn') return 'Sub burn';
	if (m === 'decision deferred') return 'Pending';
	return '—';
}
