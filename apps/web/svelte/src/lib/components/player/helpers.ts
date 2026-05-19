export interface ThumbnailCue {
  start: number; end: number;
  x: number; y: number; w: number; h: number;
}

export interface ChapterCue {
  start: number; end: number; title: string;
}

export function fmt(s: number): string {
  const h = Math.floor(s / 3600);
  const m = Math.floor((s % 3600) / 60);
  const sec = Math.floor(s % 60);
  if (h > 0) return `${h}:${String(m).padStart(2, '0')}:${String(sec).padStart(2, '0')}`;
  return `${m}:${String(sec).padStart(2, '0')}`;
}

export function parseTimestampVTT(ts: string): number {
  const parts = ts.trim().split(':');
  if (parts.length === 3) {
    return parseFloat(parts[0]) * 3600 + parseFloat(parts[1]) * 60 + parseFloat(parts[2]);
  }
  if (parts.length === 2) {
    return parseFloat(parts[0]) * 60 + parseFloat(parts[1]);
  }
  return parseFloat(parts[0]);
}

export function thumbForTime(t: number, thumbCues: ThumbnailCue[]): ThumbnailCue | null {
  if (!thumbCues.length) return null;
  return thumbCues.find(c => t >= c.start && t < c.end) ?? thumbCues[thumbCues.length - 1];
}

export interface ParsedVTTCue {
  startStr: string; endStr: string; payload: string;
}

export function parseVTTCues(text: string): ParsedVTTCue[] {
  const cues: ParsedVTTCue[] = [];
  const lines = text.split('\n');
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i].trim();
    if (!line.includes('-->')) continue;
    const arrowIdx = line.indexOf('-->');
    const startStr = line.slice(0, arrowIdx).trim();
    const endStr = line.slice(arrowIdx + 3).trim();
    const payload = (lines[i + 1] ?? '').trim();
    cues.push({ startStr, endStr, payload });
  }
  return cues;
}
