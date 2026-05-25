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
