import { describe, it, expect } from 'vitest';
import {
  fmt,
  parseTimestampVTT,
  thumbForTime,
  parseVTTCues,
  type ThumbnailCue,
} from './helpers.js';

// ─── fmt ─────────────────────────────────────────────────────────────────────

describe('fmt', () => {
  it('formats zero as 0:00', () => {
    expect(fmt(0)).toBe('0:00');
  });

  it('formats seconds-only values as M:SS', () => {
    expect(fmt(9)).toBe('0:09');
    expect(fmt(59)).toBe('0:59');
  });

  it('formats minutes correctly', () => {
    expect(fmt(60)).toBe('1:00');
    expect(fmt(65)).toBe('1:05');
    expect(fmt(3599)).toBe('59:59');
  });

  it('formats hours as H:MM:SS once >= 3600 s', () => {
    expect(fmt(3600)).toBe('1:00:00');
    expect(fmt(3661)).toBe('1:01:01');
    expect(fmt(7384)).toBe('2:03:04');
    expect(fmt(36000)).toBe('10:00:00');
  });

  it('pads minutes and seconds with leading zeros', () => {
    expect(fmt(3601)).toBe('1:00:01');
    expect(fmt(3660)).toBe('1:01:00');
  });

  it('floors fractional seconds', () => {
    expect(fmt(65.9)).toBe('1:05');
    expect(fmt(3599.99)).toBe('59:59');
  });
});

// ─── parseTimestampVTT ───────────────────────────────────────────────────────

describe('parseTimestampVTT', () => {
  it('parses HH:MM:SS.mmm', () => {
    expect(parseTimestampVTT('00:01:30.000')).toBeCloseTo(90);
    expect(parseTimestampVTT('01:00:00.000')).toBeCloseTo(3600);
    expect(parseTimestampVTT('01:01:01.500')).toBeCloseTo(3661.5);
  });

  it('parses MM:SS.mmm (two-part)', () => {
    expect(parseTimestampVTT('01:30.000')).toBeCloseTo(90);
    expect(parseTimestampVTT('00:00.000')).toBeCloseTo(0);
    expect(parseTimestampVTT('59:59.999')).toBeCloseTo(3599.999);
  });

  it('parses plain seconds', () => {
    expect(parseTimestampVTT('45.5')).toBeCloseTo(45.5);
    expect(parseTimestampVTT('0')).toBeCloseTo(0);
  });

  it('trims surrounding whitespace', () => {
    expect(parseTimestampVTT(' 00:01:00.000 ')).toBeCloseTo(60);
    expect(parseTimestampVTT('\t01:30.000\t')).toBeCloseTo(90);
  });

  it('handles integer second strings without decimals', () => {
    expect(parseTimestampVTT('00:01:00')).toBeCloseTo(60);
    expect(parseTimestampVTT('00:30')).toBeCloseTo(30);
  });
});

// ─── thumbForTime ────────────────────────────────────────────────────────────

describe('thumbForTime', () => {
  const cues: ThumbnailCue[] = [
    { start: 0,  end: 10, x: 0,   y: 0, w: 160, h: 90 },
    { start: 10, end: 20, x: 160, y: 0, w: 160, h: 90 },
    { start: 20, end: 30, x: 320, y: 0, w: 160, h: 90 },
  ];

  it('returns null when cue list is empty', () => {
    expect(thumbForTime(5, [])).toBeNull();
  });

  it('returns the cue whose range contains t', () => {
    expect(thumbForTime(0, cues)?.x).toBe(0);
    expect(thumbForTime(5, cues)?.x).toBe(0);
    expect(thumbForTime(10, cues)?.x).toBe(160);
    expect(thumbForTime(19.9, cues)?.x).toBe(160);
    expect(thumbForTime(25, cues)?.x).toBe(320);
  });

  it('end boundary is exclusive (t=end falls in next cue)', () => {
    // t=10 is the start of cue[1], not the end of cue[0]
    expect(thumbForTime(10, cues)?.x).toBe(160);
  });

  it('falls back to the last cue when t exceeds all ranges', () => {
    expect(thumbForTime(100, cues)?.x).toBe(320);
    expect(thumbForTime(30, cues)?.x).toBe(320);
  });

  it('works with a single cue', () => {
    const single: ThumbnailCue[] = [{ start: 0, end: 60, x: 0, y: 0, w: 160, h: 90 }];
    expect(thumbForTime(30, single)?.x).toBe(0);
    expect(thumbForTime(999, single)?.x).toBe(0);
  });
});

// ─── parseVTTCues (chapter / thumbnail cue extraction) ───────────────────────

describe('parseVTTCues', () => {
  it('returns empty array for files with no arrow lines', () => {
    expect(parseVTTCues('WEBVTT\n\n')).toHaveLength(0);
  });

  it('extracts start, end, and payload from a basic VTT block', () => {
    const vtt = `WEBVTT

00:00:00.000 --> 00:00:10.000
Opening credits

00:00:10.000 --> 00:01:00.000
Main action
`;
    const cues = parseVTTCues(vtt);
    expect(cues).toHaveLength(2);
    expect(cues[0].startStr).toBe('00:00:00.000');
    expect(cues[0].endStr).toBe('00:00:10.000');
    expect(cues[0].payload).toBe('Opening credits');
    expect(cues[1].payload).toBe('Main action');
  });

  it('extracts xywh sprite coordinates from thumbnail VTT payload', () => {
    const vtt = `WEBVTT

00:00:00.000 --> 00:00:10.000
sprite.jpg#xywh=0,0,160,90

00:00:10.000 --> 00:00:20.000
sprite.jpg#xywh=160,0,160,90
`;
    const cues = parseVTTCues(vtt);
    expect(cues).toHaveLength(2);
    const m0 = cues[0].payload.match(/#xywh=(\d+),(\d+),(\d+),(\d+)/);
    expect(m0).not.toBeNull();
    expect(m0![1]).toBe('0');
    expect(m0![3]).toBe('160');

    const m1 = cues[1].payload.match(/#xywh=(\d+),(\d+),(\d+),(\d+)/);
    expect(m1![1]).toBe('160');
  });

  it('handles Windows-style CRLF line endings', () => {
    const vtt = 'WEBVTT\r\n\r\n00:00:00.000 --> 00:00:05.000\r\nIntro\r\n';
    const cues = parseVTTCues(vtt);
    // Line splitting on \n still finds the --> line; payload trim handles \r
    expect(cues).toHaveLength(1);
  });

  it('ignores lines without -->', () => {
    const vtt = `WEBVTT
NOTE This is a comment
00:00:00.000 --> 00:00:05.000
Title
`;
    expect(parseVTTCues(vtt)).toHaveLength(1);
  });
});
