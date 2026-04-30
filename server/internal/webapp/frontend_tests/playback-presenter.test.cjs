const test = require("node:test");
const assert = require("node:assert/strict");

const playback = require("../static/modules/playback-presenter.js");

test("direct play forecast uses user-facing no-work language", () => {
  const decision = {
    mode: "Direct Play",
    reason: "Container and video codec match client profile",
    estimatedCpuCost: "none",
    videoAction: "direct",
  };

  assert.equal(playback.playbackReadinessLabel(decision), "Ready to play");
  assert.equal(playback.playbackEffortLabel(decision), "No extra work");
  assert.equal(playback.loadLabel(decision), "No server work");
  assert.match(playback.playbackReason(decision), /stream the file directly/);
  assert.equal(playback.sourceVideoLine({ videoCodec: "hevc", width: 3840, height: 2160 }, decision), "HEVC - 3840x2160 - Ready");
});

test("video conversion forecast explains impact without saying the file cannot play", () => {
  const decision = {
    mode: "Video Transcode",
    reason: "Source is not safely direct-playable for this profile",
    estimatedCpuCost: "high",
    videoAction: "transcode",
  };

  assert.equal(playback.playbackReadinessLabel(decision), "Video conversion");
  assert.equal(playback.playbackEffortLabel(decision, { available: true }), "GPU locked");
  assert.equal(playback.loadLabel(decision), "Heavy server work");
  assert.match(playback.playbackReason(decision, { available: true, configured: false }), /This file can still play/);
  assert.doesNotMatch(playback.playbackReason(decision), /cannot use this file/);
});

test("adaptive stream forecast explains remote resilience", () => {
  const decision = {
    mode: "Adaptive Stream",
    reason: "The selected network is below the source bitrate, so Vyrden can use adaptive streaming to step quality down before playback stalls.",
    estimatedCpuCost: "medium",
    videoAction: "adaptive",
    containerAction: "adaptive_hls",
  };

  assert.equal(playback.playbackReadinessLabel(decision), "Adaptive stream");
  assert.equal(playback.playbackActionLabel(decision.videoAction), "Adaptive stream");
  assert.equal(playback.playbackActionLabel(decision.containerAction), "Adaptive HLS");
  assert.match(playback.playbackSummary(decision), /step down before buffering stalls/);
});

test("inspector facts escape source and track data", () => {
  const html = playback.renderInspectorFacts({
    source: { probed: true, container: "<mkv>", videoCodec: "h264", bitrate: 1000, audioStreams: 1, subtitleStreams: 2 },
    tracks: { audioTracks: [{}], subtitleTracks: [{}, {}] },
    decision: { mode: "Remux", estimatedCpuCost: "low" },
    playbackState: { percent: 0.42, progressSeconds: 12 },
    formatBitrate: value => `${value} bps`,
    formatDuration: value => `${value}s`,
  });

  assert.match(html, /&lt;mkv&gt;/);
  assert.match(html, /1 audio \/ 2 subtitles/);
  assert.match(html, /42%/);
  assert.match(html, /Easy on server/);
});
