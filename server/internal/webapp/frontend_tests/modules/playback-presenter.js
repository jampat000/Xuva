(function (root, factory) {
  const api = factory();
  if (typeof module === "object" && module.exports) module.exports = api;
  root.LorivoPlayback = api;
})(typeof globalThis !== "undefined" ? globalThis : window, function () {
  function escapeHTML(value) {
    return String(value ?? "").replace(/[&<>"']/g, char => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "\"": "&quot;", "'": "&#039;" }[char]));
  }

  function serverImpact(decision = {}) {
    if (decision.mode === "Subtitle Burn" || decision.mode === "Video Transcode") return "Video conversion needed";
    if (decision.mode === "Audio Transcode") return "Audio conversion needed";
    if (decision.mode === "Adaptive Stream") return "Adaptive remote stream";
    if (decision.mode === "Remux") return "Live repackage";
    if (decision.mode === "Direct Play") return "Low impact route";
    return "Decision pending";
  }

  function playbackReadinessLabel(decision = {}) {
    const mode = String(decision.mode || "").toLowerCase();
    if (mode === "direct play") return "Ready to play";
    if (mode === "remux") return "Repackage while playing";
    if (mode === "adaptive stream") return "Adaptive stream";
    if (mode === "audio transcode") return "Audio conversion";
    if (mode === "video transcode") return "Video conversion";
    if (mode === "subtitle burn") return "Subtitle conversion";
    if (mode === "decision deferred") return "Needs file check";
    if (!mode) return "Checking";
    return decision.mode;
  }

  function playbackEffortLabel(decision = {}, hardware = {}) {
    const mode = String(decision.mode || "").toLowerCase();
    if (mode === "direct play") return "No extra work";
    if (mode === "remux") return "Low PC load";
    if (mode === "adaptive stream") return hardware.available ? "Adaptive GPU route" : "Adaptive CPU route";
    if (mode === "audio transcode") return "Light PC load";
    if (mode === "video transcode" || mode === "subtitle burn") {
      if (hardware.unlockState === "unlocked" && hardware.configured) return "GPU conversion";
      if (hardware.available) return "GPU locked";
      return "High CPU load";
    }
    if (mode === "decision deferred") return "Waiting for file check";
    return serverImpact(decision);
  }

  function loadLabel(decision = {}) {
    if (String(decision.mode || "").toLowerCase() === "decision deferred") return "Check needed";
    if (decision.estimatedCpuCost === "high") return "Heavy server work";
    if (decision.estimatedCpuCost === "medium") return "Some server work";
    if (decision.estimatedCpuCost === "low") return "Easy on server";
    if (decision.estimatedCpuCost === "none") return "No server work";
    return "Checking";
  }

  function playbackActionLabel(value = "") {
    const labels = {
      probe_required: "Needs file check",
      pending: "Checking",
      none: "None",
      direct: "Ready",
      direct_play: "Ready",
      copy: "No conversion",
      remux: "Live repackage",
      adaptive: "Adaptive stream",
      adaptive_hls: "Adaptive HLS",
      transcode: "Convert",
      burn_in: "Burn subtitles",
      selected_source: "Selected source",
    };
    return labels[String(value).toLowerCase()] || value;
  }

  function playbackReason(decision = {}, hardware = {}) {
    const reason = String(decision.reason || "");
    const mode = String(decision.mode || "").toLowerCase();
    if (!reason) return "Lorivo is checking this file before choosing the best playback path.";
    if (reason.includes("has not been probed")) return "Lorivo needs to check this file once before it can confirm the best playback path.";
    if (reason.includes("Container and video codec match")) return "This player can stream the file directly. Lorivo should not need extra CPU, GPU, or temporary disk work.";
    if (reason.includes("Video can direct play")) return "The video can play as-is, but Lorivo may convert the audio track for this player. Expect a light CPU load.";
    if (reason.includes("video codec is compatible")) return "The video can stay untouched, but Lorivo may need to repackage the file while playing for this device. No permanent file is created.";
    if (reason.includes("subtitle track is image-based")) return "This file can still play, but image subtitles may need to be burned into the video. That is a heavy path and can use significant CPU/GPU.";
    if (reason.includes("not safely direct-playable")) {
      if (mode === "video transcode") return hardware.configured && hardware.unlockState === "unlocked"
        ? "This file can still play, but this player profile needs video conversion. GPU acceleration is unlocked and is the right path for keeping CPU load low."
        : "This file can still play, but this player profile needs video conversion. Without hardware acceleration, expect high CPU use, more power draw, and more heat.";
      if (mode === "subtitle burn") return "This file can still play, but subtitles may need to be burned into the video for this player. This is one of the heaviest playback paths.";
      if (mode === "audio transcode") return "The video can stay intact, but Lorivo may convert audio for this player. This is usually a light PC load.";
      if (mode === "remux") return "The streams can stay intact, but Lorivo may repackage the file while playing for this device. This is temporary and usually low impact.";
      if (mode === "adaptive stream") return "This file can still play through an adaptive remote stream. Lorivo can lower quality during weak network moments instead of hard buffering.";
      return "This file can still play, but Lorivo may need to prepare it before or during playback for this player profile.";
    }
    if (reason.includes("adaptive streaming")) return "Lorivo can use adaptive streaming so remote playback can step quality down before stalls.";
    return reason;
  }

  function playbackSummary(decision = {}, hardware = {}, policy = {}) {
    const mode = String(decision.mode || "").toLowerCase();
    if (!mode || mode === "decision deferred") return playbackReason(decision, hardware);
    if (mode === "direct play") return "This source should play as-is with no conversion work.";
    if (policy && policy.allowed === false) {
      return "This source can play, but this player may need a fallback if the original stream is not accepted.";
    }
    if (mode === "video transcode" || mode === "subtitle burn") {
      return hardware.available
        ? "This source can play with video conversion; GPU support can reduce CPU load when enabled."
        : "This source can play with video conversion, but it will use more CPU.";
    }
    if (mode === "remux") return "This source can play after a live container repack with no quality loss.";
    if (mode === "adaptive stream") return "This source can use adaptive streaming so remote quality can step down before buffering stalls.";
    if (mode === "audio transcode") return "This source can play with lightweight audio conversion.";
    return playbackReason(decision, hardware);
  }

  function sourceVideoLine(source = {}, decision = {}) {
    return [source.videoCodec ? String(source.videoCodec).toUpperCase() : "Video pending", source.width && source.height ? `${source.width}x${source.height}` : "", playbackActionLabel(decision.videoAction || "")].filter(Boolean).join(" - ");
  }

  function sourceCodecLine(source = {}) {
    if (!source.probed) return "File check needed";
    return [source.videoCodec ? String(source.videoCodec).toUpperCase() : "video", source.container ? String(source.container).toUpperCase() : "container", source.audioStreams ? `${source.audioStreams} audio` : "", source.subtitleStreams ? `${source.subtitleStreams} subs` : ""].filter(Boolean).join(" - ");
  }

  function inspectorFact(label, value, note) {
    return `<div class="inspector-fact"><span>${escapeHTML(label)}</span><strong>${escapeHTML(value || "Pending")}</strong><small>${escapeHTML(note || "")}</small></div>`;
  }

  function renderInspectorFacts(model = {}) {
    const source = model.source || {};
    const tracks = model.tracks || {};
    const decision = model.decision || {};
    const playbackState = model.playbackState || {};
    const formatBitrate = model.formatBitrate || (value => `${value || 0} bps`);
    const formatDuration = model.formatDuration || (value => `${value || 0}s`);
    return `<div class="source-inspector-grid">
      ${inspectorFact("Probe", source.probed ? "Complete" : "Needed", source.probed ? "Media facts are cached locally." : "Run file check for this source.")}
      ${inspectorFact("Container", source.container || "Pending", sourceCodecLine(source))}
      ${inspectorFact("Video", sourceVideoLine(source, decision), source.bitrate ? formatBitrate(source.bitrate) : "Bitrate pending")}
      ${inspectorFact("Tracks", `${tracks.audioTracks?.length || 0} audio / ${tracks.subtitleTracks?.length || 0} subtitles`, source.subtitleStreams ? `${source.subtitleStreams} embedded subtitle streams` : "No subtitle facts yet")}
      ${inspectorFact("Progress", playbackState.watched ? "Watched" : `${Math.round((playbackState.percent || 0) * 100)}%`, playbackState.progressSeconds ? formatDuration(playbackState.progressSeconds) : "Not started")}
      ${inspectorFact("Server work", playbackEffortLabel(decision, model.hardware || {}), loadLabel(decision))}
    </div>`;
  }

  return {
    escapeHTML,
    serverImpact,
    playbackReadinessLabel,
    playbackEffortLabel,
    loadLabel,
    playbackActionLabel,
    playbackReason,
    playbackSummary,
    sourceVideoLine,
    sourceCodecLine,
    inspectorFact,
    renderInspectorFacts,
  };
});
