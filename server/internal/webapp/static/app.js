const state = { activeView: "dashboard", activeSession: "", playbackSelections: {} };

const view = document.getElementById("view");
const viewTitle = document.getElementById("viewTitle");
const serverStatus = document.getElementById("serverStatus");
const serverDot = document.getElementById("serverDot");
const themeToggle = document.getElementById("themeToggle");
const densityMenu = document.getElementById("densityMenu");
const densityButton = document.getElementById("densityButton");
const densityLabel = document.getElementById("densityLabel");
const densityOptions = document.getElementById("densityOptions");
const densityNames = { compact: "Compact", balanced: "Balanced", comfortable: "Comfortable", cinematic: "Cinematic" };

const savedTheme = localStorage.getItem("vyrden-theme") || "dark";
applyTheme(savedTheme);
themeToggle?.addEventListener("click", () => {
  applyTheme(document.body.dataset.theme === "light" ? "dark" : "light");
});
const savedDensity = localStorage.getItem("vyrden-density") || "comfortable";
applyDensity(savedDensity);
densityButton?.addEventListener("click", () => {
  const open = densityMenu?.classList.toggle("open");
  densityButton.setAttribute("aria-expanded", open ? "true" : "false");
  if (open) positionDensityMenu();
});
densityOptions?.querySelectorAll("[data-density-option]").forEach(button => {
  button.addEventListener("click", () => {
    applyDensity(button.dataset.densityOption);
    closeDensityMenu();
  });
});
document.addEventListener("click", event => {
  if (!densityMenu?.contains(event.target)) closeDensityMenu();
});
document.addEventListener("keydown", event => {
  if (event.key === "Escape") closeDensityMenu();
});
window.addEventListener("resize", positionDensityMenu);
window.addEventListener("scroll", positionDensityMenu, true);

document.querySelectorAll(".nav-item").forEach(button => {
  button.addEventListener("click", () => navigate(button.dataset.view));
});

async function api(path, options) {
  const response = await fetch(path, options);
  const text = await response.text();
  const payload = text ? JSON.parse(text) : {};
  if (!response.ok) throw new Error(payload.error || response.statusText);
  return payload;
}

async function send(path, body, method = "POST") {
  return api(path, { method, headers: { "Content-Type": "application/json" }, body: JSON.stringify(body || {}) });
}

async function navigate(name) {
  state.activeView = name;
  document.querySelectorAll(".nav-item").forEach(button => button.classList.toggle("active", button.dataset.view === name));
  viewTitle.textContent = title(name);
  view.innerHTML = `<div class="card"><pre>Loading ${title(name)}...</pre></div>`;
  try {
    await render(name);
  } catch (error) {
    view.innerHTML = `<div class="card error">${escapeHTML(error.message)}</div>`;
  }
}

function title(name) {
  return { dashboard: "Home", movies: "Movies", tv: "TV", libraries: "Libraries", activity: "Activity", health: "Review", playback: "Playback Lab", remote: "Remote Access", settings: "Settings" }[name] || name;
}

function applyTheme(theme) {
  const next = theme === "light" ? "light" : "dark";
  document.body.dataset.theme = next;
  localStorage.setItem("vyrden-theme", next);
  if (themeToggle) themeToggle.textContent = next === "light" ? "Light" : "Dark";
}

function applyDensity(density) {
  const next = densityNames[density] ? density : "comfortable";
  document.documentElement.dataset.density = next;
  document.body.dataset.density = next;
  localStorage.setItem("vyrden-density", next);
  if (densityLabel) densityLabel.textContent = densityNames[next];
  densityOptions?.querySelectorAll("[data-density-option]").forEach(button => {
    const selected = button.dataset.densityOption === next;
    button.classList.toggle("selected", selected);
    button.setAttribute("aria-checked", selected ? "true" : "false");
  });
  positionDensityMenu();
}

function closeDensityMenu() {
  densityMenu?.classList.remove("open");
  densityButton?.setAttribute("aria-expanded", "false");
}

function positionDensityMenu() {
  if (!densityMenu?.classList.contains("open") || !densityButton || !densityOptions) return;
  const rect = densityButton.getBoundingClientRect();
  const styles = getComputedStyle(document.body);
  const densityScale = Number.parseFloat(styles.getPropertyValue("--density-scale")) || 1;
  const gap = Math.max(8, Math.round(8 * densityScale));
  const minWidth = Math.max(rect.width, Math.round(190 * densityScale));
  const left = Math.min(Math.max(12, rect.left), window.innerWidth - minWidth - 12);
  densityOptions.style.minWidth = `${minWidth}px`;
  densityOptions.style.left = `${left}px`;
  densityOptions.style.top = `${rect.bottom + gap}px`;
}

async function refreshShell() {
  const health = await api("/api/health");
  serverStatus.textContent = health.status || "unknown";
  serverDot.classList.toggle("ok", health.status === "ok");
}

async function render(name) {
  if (name === "dashboard") return renderDashboard();
  if (name === "movies") return renderMovies();
  if (name === "tv") return renderTV();
  if (name === "libraries") return renderLibraries();
  if (name === "activity") return renderActivity();
  if (name === "health") return renderHealth();
  if (name === "playback") return renderPlaybackLab();
  if (name === "remote") return renderRemote();
  if (name === "settings") return renderSettings();
}

async function renderDashboard() {
  const [summary, libraries, scans, probes, work, downloads, recent, sessions, health, versions, sources, system] = await Promise.all([
    api("/api/catalog/summary"),
    api("/api/libraries"),
    api("/api/scans"),
    api("/api/probes"),
    api("/api/work"),
    api("/api/downloads"),
    api("/api/playback/recent"),
    api("/api/sessions"),
    api("/api/catalog/health"),
    api("/api/versions"),
    api("/api/media-sources?limit=200"),
    api("/api/system/status"),
  ]);
  const reviewCount = health.needsReview || 0;
  const directPlayable = summary.mediaSources ? Math.max(0, Math.round(((summary.mediaSources - (health.unsupported || 0)) / summary.mediaSources) * 100)) : 0;
  const movieCount = summary.movies || 0;
  const seriesCount = summary.series || 0;
  const episodeCount = summary.episodes || 0;
  const totalTitles = movieCount + seriesCount;
  const activeSessions = sessions.sessions || [];
  const scanJobs = scans.scans || [];
  const probeJobs = probes.probes || [];
  const workJobs = work.work || [];
  const downloadJobs = downloads.downloads || [];
  const mediaSources = sources.mediaSources || [];
  const quality = sourceQuality(mediaSources);
  const activeJobs = [...scanJobs, ...probeJobs, ...workJobs, ...downloadJobs].filter(isActiveJob);
  view.innerHTML = `
    <div class="dashboard stack">
      <section class="hero">
        <article class="feature">
          <div class="feature-content">
            <p class="eyebrow">Live media command centre</p>
            <h1>${dashboardCommandTitle(activeSessions, summary)}</h1>
            <p class="lead">${dashboardCommandCopy(activeSessions, summary, health, quality)}</p>
            <div class="meta-line">
              <span class="badge">${summary.mediaSources || 0} sources</span>
              <span class="badge">${summary.libraries || 0} libraries</span>
              <span class="badge ${summary.unprobed ? "warn" : "good"}">${summary.unprobed || 0} unprobed</span>
              <span class="badge ${directPlayable >= 90 ? "good" : "warn"}">${directPlayable}% playable</span>
              <span class="badge route">Live</span>
            </div>
            <div class="actions">
              <button class="primary" onclick="navigate('${movieCount ? "movies" : "tv"}')">${movieCount ? "Movies" : "TV"}</button>
              <button onclick="navigate('playback')">Playback</button>
              <button onclick="navigate('activity')">Activity</button>
              <button onclick="navigate('health')">Review ${reviewCount}</button>
            </div>
          </div>
          ${dashboardSnapshot({ activeSessions, activeJobs, summary, health, quality, system })}
        </article>

        <aside class="panel pad">
          <div class="panel-title"><strong>Playing Now</strong><span class="badge route">${activeSessions.length ? "Live" : "Idle"}</span></div>
          ${playingNow(activeSessions)}
        </aside>
      </section>

      <section class="insight-grid">
        ${metric("Movies", movieCount, "Feature films indexed")}
        ${metric("TV", `${seriesCount} / ${episodeCount}`, "Series and episodes")}
        ${metric("Files", summary.mediaSources || 0, `${quality.totalSize ? formatBytes(quality.totalSize) : "Size pending"} sampled/indexed`)}
        ${metric("Direct Playable", `${directPlayable}%`, `${health.unsupported || 0} unsupported sources`)}
      </section>

      <section class="command-grid">
        <div class="panel pad">
          <div class="panel-title"><strong>File Intelligence</strong><button onclick="navigate('playback')">Inspect</button></div>
          ${fileIntelligence(quality, summary, health)}
        </div>
        <div class="panel pad">
          <div class="panel-title"><strong>Operations</strong><span class="badge ${activeJobs.length ? "warn" : "good"}">${activeJobs.length ? `${activeJobs.length} running` : "Clear"}</span></div>
          ${operationsPanel(scanJobs, probeJobs, workJobs, downloadJobs)}
        </div>
        <div class="panel pad">
          <div class="panel-title"><strong>Needs Attention</strong><div class="inline-actions"><button onclick="startProbe()">Probe</button><button onclick="navigate('health')">Review</button></div></div>
          ${attentionPanel(summary, health, versions.versions || [])}
        </div>
        <div class="panel pad">
          <div class="panel-title"><strong>Hardware</strong><span class="badge route">${system.cpu.cores} cores</span></div>
          ${hardwarePanel(system)}
        </div>
      </section>

      <section class="shelf-section">
        <div class="shelf-head">
          <div>
            <div class="section-title">Continue Watching</div>
            <p>Ordered by local activity, not outside recommendations.</p>
          </div>
          <button onclick="navigate('activity')">All activity</button>
        </div>
        ${mediaCards(recent.recent)}
      </section>

      <section class="dashboard-grid">
        <div class="panel pad">
          <div class="panel-title"><strong>Libraries & Storage</strong><button onclick="navigate('libraries')">Manage</button></div>
          ${libraryCards(libraries.libraries)}
        </div>
        <div class="panel pad"><div class="panel-title"><strong>Recent Scans</strong><button onclick="startScan('/api/libraries/scan')">Scan all</button></div>${jobCards(scanJobs)}</div>
      </section>

      <section class="dashboard-grid">
        <div class="panel pad"><div class="panel-title"><strong>Version Intelligence</strong><button onclick="navigate('health')">Review</button></div>${versionGroups(versions.versions, totalTitles)}</div>
        <div class="panel pad"><div class="panel-title"><strong>Runtime Folders</strong><button onclick="navigate('settings')">Move</button></div>${runtimeFolders(system.disks || [])}</div>
      </section>

      <section class="panel pad">
        <div class="panel-title"><strong>Live Signals</strong><span class="live-stamp">Updated ${new Date().toLocaleTimeString()}</span></div>
        <div class="signal-stack">${signalPill("Sessions", activeSessions.length)}${signalPill("Jobs", activeJobs.length)}${signalPill("Downloads", downloadJobs.length)}${signalPill("Review", reviewCount)}${signalPill("CPU", `${Math.round(system.cpu.percent || 0)}%`)}</div>
      </section>
    </div>`;
}

async function renderMovies() {
  const payload = await api("/api/movies?limit=500");
  view.innerHTML = `
    <div class="stack">
      <div class="shelf-head">
        <div>
          <h2>Movie Wall</h2>
          <p>Poster-first browsing with version and review signals kept close to the title.</p>
        </div>
      </div>
      <div class="toolbar">
        <button class="primary" onclick="startScan('/api/libraries/movies/scan')">Scan Movies</button>
        <input class="search" id="movieFilter" placeholder="Filter movies" oninput="filterCards(this.value)">
      </div>
      <div class="poster-grid">${payload.movies.map(movieCard).join("") || empty("No movies yet. Run a scan first.")}</div>
    </div>`;
}

function movieCard(item) {
  return `<article class="poster-card" data-filter="${escapeAttr(item.title)} ${item.year || ""}" data-initial="${escapeAttr(initials(item.title))}">
    <img alt="" src="${artworkURL("movie", item.id)}" loading="lazy">
    <button class="poster-action" onclick="showMovie('${item.id}')">
      <span class="poster-title">${escapeHTML(item.title)}</span>
      <small>${item.year || "Unknown year"} - ${item.versionCount || 0} version${item.versionCount === 1 ? "" : "s"}</small>
      ${item.needsReview ? `<em>Review</em>` : ""}
    </button>
  </article>`;
}

async function showMovie(id) {
  const movie = await api(`/api/movies/${id}`);
  const rows = await Promise.all(movie.versions.map(async (version, index) => {
    const state = await api(`/api/playback/state/${version.mediaSourceId}`);
    return versionCard(version, state, index === 0);
  }));
  const selected = movie.versions[0];
  const tracks = selected ? await api(`/api/media-sources/${selected.mediaSourceId}/tracks`).catch(() => ({ audioTracks: [], subtitleTracks: [] })) : { audioTracks: [], subtitleTracks: [] };
  if (selected) seedPlaybackSelection(selected.mediaSourceId, tracks);
  viewTitle.textContent = movie.title;
  view.innerHTML = `
    <div class="detail-command">
      <div class="detail-backdrop"><img alt="" src="${artworkURL("movie", movie.id)}"></div>
      <section class="detail-main">
        <button class="back-button" onclick="navigate('movies')">Back to Movies</button>
        <div class="eyebrow">Featured from your library</div>
        <h1>${escapeHTML(movie.title)}</h1>
        <div class="meta-line">
          <span class="badge">${movie.year || "Unknown year"}</span>
          <span class="badge">${movie.versionCount} version${movie.versionCount === 1 ? "" : "s"}</span>
          <span class="badge ${movie.needsReview ? "warn" : "good"}">${movie.needsReview ? "Needs Review" : "Matched"}</span>
          <span class="badge route">${selected ? "Direct Play" : "No Source"}</span>
          ${metadataBadges(movie.metadata)}
        </div>
        <p class="lead">${movieOverview(movie)}</p>
        ${ratingsRail(movie.metadata)}
        ${selected ? `<div class="source-strip"><div><span>Selected source</span><strong>${escapeHTML(selected.qualityLabel || "Original source")}</strong></div><div><span>File size</span><strong>${formatBytes(selected.sizeBytes)}</strong></div><div><span>Container</span><strong>${escapeHTML((selected.relPath || "").split(".").pop() || "media")}</strong></div></div>` : ""}
        <div class="actions">
          ${selected ? `<a class="button primary focusable" href="/play/${selected.mediaSourceId}" target="_blank">Resume</a>` : ""}
          ${selected ? `<a class="button focusable" href="/play/${selected.mediaSourceId}?start=0" target="_blank">Play From Start</a>` : ""}
          ${selected ? `<button class="button focusable" onclick="markWatched('${selected.mediaSourceId}', true)">Mark Watched</button>` : ""}
          ${selected ? `<button class="button focusable" onclick="markWatched('${selected.mediaSourceId}', false)">Mark Unwatched</button>` : ""}
          <button class="button focusable" onclick="refreshMetadata('movie','${movie.id}','${escapeAttr(movie.title)}',${movie.year || 0})">Fetch Ratings</button>
          ${movie.needsReview ? `<button class="button focusable" onclick="openMetadataFix('movie','${movie.id}','${escapeAttr(movie.title)}',${movie.year || 0})">Fix Match</button>` : ""}
        </div>
        <section class="section">
          <div class="section-head">
            <div class="section-title">Versions</div>
            <div class="section-note">Choose source quality before playback</div>
          </div>
          <div class="version-grid">${rows.join("") || empty("No playable versions found.")}</div>
        </section>
        <section class="section">
          <div class="section-head">
            <div class="section-title">Audio & Subtitles</div>
            <div class="section-note">Selected tracks affect compatibility</div>
          </div>
          <div class="track-grid">
            <div class="track-panel"><strong>Audio</strong>${audioTrackRows(tracks.audioTracks, selected?.mediaSourceId)}</div>
            <div class="track-panel"><strong>Subtitles</strong>${subtitleTrackRows(tracks.subtitleTracks, selected?.mediaSourceId)}</div>
          </div>
        </section>
      </section>

      <aside class="side detail-side">
        <section class="panel pad">
          <div class="panel-title"><strong>Playback Forecast</strong><span class="badge route" id="forecastLoad">0% load</span></div>
          <div class="decision"><strong id="forecastMode">${selected ? "Direct Play" : "No Source"}</strong><span id="forecastReason">${selected ? "Client supports the selected source. No server transcode required unless tracks change." : "Scan or attach a source before playback."}</span></div>
          <div class="kv">
            <div><span>Reason</span><span id="forecastReasonShort">${selected ? "Source available" : "Missing source"}</span></div>
            <div><span>Video</span><span id="forecastVideo">${selected ? escapeHTML(selected.qualityLabel || "Source") : "None"}</span></div>
            <div><span>Audio</span><span id="forecastAudio">Passthrough ready</span></div>
            <div><span>Subtitles</span><span id="forecastSubtitles">SRT optional</span></div>
            <div><span>Server</span><span id="forecastServer">Low impact route</span></div>
          </div>
        </section>
        <section class="panel pad">
          <div class="panel-title"><strong>Download</strong></div>
          <div class="download">
            ${selected ? `<button class="download-option" onclick="startDownload('${selected.mediaSourceId}','original')"><span>Original</span><span>${formatBytes(selected.sizeBytes)}</span></button>` : ""}
            ${selected ? `<button class="download-option" onclick="startDownload('${selected.mediaSourceId}','balanced')"><span>Balanced</span><span>1080p prepared</span></button>` : ""}
            ${selected ? `<button class="download-option" onclick="startDownload('${selected.mediaSourceId}','travel')"><span>Travel</span><span>720p small</span></button>` : ""}
            ${movie.needsReview ? `<button onclick="openMetadataFix('movie','${movie.id}','${escapeAttr(movie.title)}',${movie.year || 0})">Fix Match</button>` : ""}
          </div>
        </section>
        <section class="panel pad">
          <div class="panel-title"><strong>Metadata Source</strong><span class="badge route">${escapeHTML(providerLabel(movie.metadata?.provider || "pending"))}</span></div>
          <div class="decision"><strong>${escapeHTML(movie.metadata?.title || movie.title)}</strong><span>${metadataSummary(movie.metadata)}</span></div>
          <div class="kv">
            <div><span>Provider</span><span>${escapeHTML(providerLabel(movie.metadata?.provider || "none"))}</span></div>
            <div><span>Confidence</span><span>${metadataConfidenceLabel(movie.metadata)}</span></div>
            <div><span>External ID</span><span>${escapeHTML(metadataExternalID(movie.metadata))}</span></div>
          </div>
          <div class="inline-actions"><button onclick="refreshMetadata('movie','${movie.id}','${escapeAttr(movie.title)}',${movie.year || 0})">Fetch metadata</button>${movie.needsReview ? `<button onclick="openMetadataFix('movie','${movie.id}','${escapeAttr(movie.title)}',${movie.year || 0})">Fix match</button>` : ""}</div>
        </section>
        <section class="panel pad">
          <div class="panel-title"><strong>Media Intelligence</strong><span class="badge good">Clean</span></div>
          <div class="cast-list">
            <div><i></i><strong>Container</strong><span>${selected ? escapeHTML((selected.relPath || "").split(".").pop() || "media") : "none"}</span></div>
            <div><i></i><strong>Storage</strong><span>${selected ? escapeHTML(selected.relPath) : "No linked file"}</span></div>
            <div><i></i><strong>Remote</strong><span>Use alternate encode if needed</span></div>
          </div>
        </section>
      </aside>
    </div>`;
  if (selected) updatePlaybackForecast(selected.mediaSourceId);
}

function versionCard(version, state, selected) {
  const watched = state.watched ? "Watched" : state.progressSeconds > 0 ? `Resume ${Math.round(state.percent * 100)}%` : "Unplayed";
  return `<article class="version ${selected ? "is-selected" : ""}">
    <div class="version-title">${escapeHTML(version.qualityLabel || "Source")}<span class="version-path">${selected ? "Direct Play" : watched}</span></div>
    <div class="version-details"><span>${escapeHTML(version.relPath)}</span><span>${formatBytes(version.sizeBytes)}</span></div>
    <div class="inline-actions">
      <a class="button primary" href="/play/${version.mediaSourceId}" target="_blank">${state.progressSeconds > 5 && !state.watched ? "Resume" : "Play"}</a>
      <a class="button" href="/play/${version.mediaSourceId}?start=0" target="_blank">Start Over</a>
      <button onclick="markWatched('${version.mediaSourceId}', true)">Watched</button>
      <button onclick="markWatched('${version.mediaSourceId}', false)">Unwatched</button>
    </div>
  </article>`;
}

function seedPlaybackSelection(mediaSourceId, tracks = {}) {
  const audioTracks = Array.isArray(tracks.audioTracks) ? tracks.audioTracks : [];
  const subtitleTracks = Array.isArray(tracks.subtitleTracks) ? tracks.subtitleTracks : [];
  const current = state.playbackSelections[mediaSourceId] || {};
  if (!current.audio && audioTracks.length) {
    current.audio = audioTracks.find(item => item.default) || audioTracks[0];
  }
  if (current.subtitle === undefined) {
    current.subtitle = subtitleTracks.find(item => item.default && item.forced) || null;
  }
  state.playbackSelections[mediaSourceId] = current;
}

function trackRow(label, meta, selected, action = "") {
  return `<button class="track-row ${selected ? "selected" : ""}" type="button" ${action}><span>${escapeHTML(label)}</span><em>${escapeHTML(meta)}</em></button>`;
}

function audioTrackRows(items = [], mediaSourceId = "") {
  items = Array.isArray(items) ? items : [];
  if (!items.length) return trackRow("Probe required", "No audio track data yet", true);
  const selected = state.playbackSelections[mediaSourceId]?.audio;
  return items.map((item, index) => {
    const label = `${trackLanguage(item)} - ${String(item.codec || "audio").toUpperCase()}${item.channels ? ` ${item.channels}ch` : ""}`;
    const meta = item.default ? "Default" : index === 0 ? "Primary" : "Selectable";
    const isSelected = selected ? selected.index === item.index : index === 0 || item.default;
    return trackRow(label, meta, isSelected, `onclick="selectPlaybackTrack('audio','${mediaSourceId}',${escapeAttr(JSON.stringify(item))})"`);
  }).join("");
}

function subtitleTrackRows(items = [], mediaSourceId = "") {
  items = Array.isArray(items) ? items : [];
  const selected = state.playbackSelections[mediaSourceId]?.subtitle;
  const off = trackRow("Subtitles off", "Best direct-play route", !selected, `onclick="selectPlaybackTrack('subtitle','${mediaSourceId}',null)"`);
  if (!items.length) return off + trackRow("No embedded subtitles", "Sidecars checked separately", false);
  return off + items.map((item, index) => {
    const label = `${trackLanguage(item)} - ${String(item.codec || "subtitle").toUpperCase()}${item.forced ? " forced" : ""}`;
    const meta = imageSubtitle(item.codec) ? "May burn in" : "Direct/convert";
    return trackRow(label, meta, selected?.index === item.index, `onclick="selectPlaybackTrack('subtitle','${mediaSourceId}',${escapeAttr(JSON.stringify(item))})"`);
  }).join("");
}

function selectPlaybackTrack(kind, mediaSourceId, track) {
  const current = state.playbackSelections[mediaSourceId] || {};
  current[kind] = track;
  state.playbackSelections[mediaSourceId] = current;
  event?.currentTarget?.parentElement?.querySelectorAll(".track-row").forEach(row => row.classList.remove("selected"));
  event?.currentTarget?.classList.add("selected");
  updatePlaybackForecast(mediaSourceId);
}

async function updatePlaybackForecast(mediaSourceId) {
  const selected = state.playbackSelections[mediaSourceId] || {};
  const params = new URLSearchParams({ mediaSourceId, clientProfile: "web" });
  if (selected.audio) {
    params.set("audioTrackIndex", selected.audio.index ?? 0);
    params.set("audioCodec", selected.audio.codec || "");
    params.set("audioChannels", selected.audio.channels || 0);
  }
  if (selected.subtitle) {
    params.set("subtitleTrackActive", "true");
    params.set("subtitleTrackIndex", selected.subtitle.index ?? 0);
    params.set("subtitleCodec", selected.subtitle.codec || "");
  }
  const decision = await api(`/api/playback/decision?${params.toString()}`);
  setText("forecastMode", decision.mode);
  setText("forecastReason", decision.reason);
  setText("forecastReasonShort", decision.containerAction || "selected source");
  setText("forecastAudio", decision.audioAction || "pending");
  setText("forecastSubtitles", decision.subtitleAction || "none");
  setText("forecastServer", serverImpact(decision));
  setText("forecastLoad", decision.estimatedCpuCost === "high" ? "high load" : decision.estimatedCpuCost === "medium" ? "some load" : "low load");
}

function setText(id, value) {
  const target = document.getElementById(id);
  if (target) target.textContent = value ?? "";
}

function serverImpact(decision = {}) {
  if (decision.mode === "Subtitle Burn" || decision.mode === "Video Transcode") return "Transcode required";
  if (decision.mode === "Audio Transcode") return "Audio transcode";
  if (decision.mode === "Remux") return "Container remux";
  if (decision.mode === "Direct Play") return "Low impact route";
  return "Decision pending";
}

function trackLanguage(item = {}) {
  return item.language && item.language !== "und" ? item.language.toUpperCase() : "Unknown";
}

function imageSubtitle(codec = "") {
  return ["hdmv_pgs_subtitle", "dvd_subtitle", "dvb_subtitle", "pgs"].includes(String(codec).toLowerCase());
}

async function renderTV() {
  const payload = await api("/api/series?limit=500");
  view.innerHTML = `
    <div class="stack">
      <div class="shelf-head">
        <div>
          <h2>Series Wall</h2>
          <p>Shows stay grouped by season with immediate episode playback.</p>
        </div>
      </div>
      <div class="toolbar">
        <button class="primary" onclick="startScan('/api/libraries/tv/scan')">Scan TV</button>
        <input class="search" placeholder="Filter shows" oninput="filterCards(this.value)">
      </div>
      <div class="poster-grid">${payload.series.map(seriesCard).join("") || empty("No TV yet. Run a scan first.")}</div>
    </div>`;
}

function seriesCard(item) {
  return `<article class="poster-card" data-filter="${escapeAttr(item.title)}" data-initial="${escapeAttr(initials(item.title))}">
    <img alt="" src="${artworkURL("series", item.id)}" loading="lazy">
    <button class="poster-action" onclick="showSeries('${item.id}')">
      <span class="poster-title">${escapeHTML(item.title)}</span>
      <small>${item.seasonCount || 0} seasons - ${item.episodeCount || 0} episodes</small>
    </button>
  </article>`;
}

async function showSeries(id) {
  const series = await api(`/api/series/${id}`);
  const seasonSections = await Promise.all(series.seasons.map(async season => `<section class="card"><h2>Season ${season.seasonNumber}</h2>${await episodeList(season.episodes)}</section>`));
  viewTitle.textContent = series.title;
  view.innerHTML = `
    <div class="detail-shell">
      <section class="detail-hero">
        <img alt="" src="${artworkURL("series", series.id)}">
        <div class="detail-poster"><img alt="" src="${artworkURL("series", series.id)}"></div>
        <div class="detail-copy">
          <button class="ghost" onclick="navigate('tv')">Back to TV</button>
          <h2>${escapeHTML(series.title)}</h2>
          <p>${series.seasonCount} seasons - ${series.episodeCount} episodes</p>
        </div>
      </section>
      <div class="stack">
        ${seasonSections.join("")}
      </div>
    </div>`;
}

async function episodeList(episodes) {
  if (!episodes.length) return empty("No episodes in this season.");
  const rows = await Promise.all(episodes.map(async episode => {
    const version = episode.versions && episode.versions[0];
    const state = version ? await api(`/api/playback/state/${version.mediaSourceId}`).catch(() => ({})) : {};
    const playLabel = state.progressSeconds > 5 && !state.watched ? "Resume" : "Play";
    const watchedLabel = state.watched ? "Watched" : state.progressSeconds > 5 ? `${Math.round((state.percent || 0) * 100)}%` : `${episode.versionCount || 0} version${episode.versionCount === 1 ? "" : "s"}`;
    const play = version ? `<a class="button primary" href="/play/${version.mediaSourceId}" target="_blank">${playLabel}</a><a class="button" href="/play/${version.mediaSourceId}?start=0" target="_blank">Start Over</a><button onclick="markWatched('${version.mediaSourceId}', true)">Watched</button><button onclick="markWatched('${version.mediaSourceId}', false)">Unwatched</button>` : `<span class="muted">No source</span>`;
    const label = episode.episodeEnd && episode.episodeEnd !== episode.episodeNumber ? `E${episode.episodeNumber}-E${episode.episodeEnd}` : `E${episode.episodeNumber}`;
    return `<div class="episode-row">
      <strong>${label}</strong>
      <span>${escapeHTML(episode.title || "Episode")}</span>
      <small>${escapeHTML(watchedLabel)}</small>
      <div class="inline-actions">${play}${episode.needsReview ? `<button onclick="openMetadataFix('episode','${episode.id}','${escapeAttr(episode.title || "Episode")}',0)">Fix</button>` : ""}</div>
    </div>`;
  }));
  return rows.join("");
}

async function renderLibraries() {
  const payload = await api("/api/libraries");
  view.innerHTML = `
    <div class="stack">
      <div class="library-panel">
        <div>
          <h2>Add Library</h2>
          <p class="muted">Use any local, USB, mapped, NAS, SMB, NFS, or mounted path that the logged-in user can access.</p>
        </div>
        <form class="library-form" onsubmit="saveLibrary(event)">
          <select name="kind"><option value="movies">Movies</option><option value="tv">TV</option></select>
          <input name="name" placeholder="Library name">
          <input name="path" placeholder="D:\\Media\\Movies or \\\\NAS\\Share\\TV" required>
          <button class="primary">Add</button>
        </form>
      </div>
      <div class="library-grid">${libraryCards(payload.libraries, true)}</div>
    </div>`;
}

function libraryCards(items = [], controls = false) {
  if (!items.length) return empty("No libraries configured yet.");
  return items.map(item => `<article class="library-card">
    <div class="library-card-head"><div><strong>${escapeHTML(item.name)}</strong><span>${escapeHTML(item.kind)} library</span></div><span class="badge">${escapeHTML(item.storageType || "storage")}</span></div>
    <code>${escapeHTML(item.path)}</code>
    ${controls ? `<div class="inline-actions"><button onclick="scanLibrary('${item.id}')">Scan</button><button onclick="deleteLibrary('${item.id}')">Remove</button></div>` : ""}
  </article>`).join("");
}

async function saveLibrary(event) {
  event.preventDefault();
  const data = new FormData(event.currentTarget);
  await send("/api/libraries", { kind: data.get("kind"), name: data.get("name"), path: data.get("path") });
  navigate("libraries");
}

async function scanLibrary(id) {
  await send(`/api/libraries/${id}/scan`, {});
  navigate("activity");
}

async function deleteLibrary(id) {
  if (!confirm("Remove this library from Vyrden? Media files are not deleted.")) return;
  await api(`/api/libraries/${id}`, { method: "DELETE" });
  navigate("libraries");
}

async function renderActivity() {
  const [scans, probes, work, downloads, sessions] = await Promise.all([api("/api/scans"), api("/api/probes"), api("/api/work"), api("/api/downloads"), api("/api/sessions")]);
  view.innerHTML = `
    <div class="stack">
      <div class="hero-strip">
        <div class="hero-copy">
          <span>Operations</span>
          <strong>Background work stays visible without taking over playback.</strong>
          <p>Scanning, probing, and transcode jobs run in separate queues so library maintenance does not fight active playback.</p>
        </div>
        <div class="signal-stack">
          ${signalPill("Sessions", sessions.sessions.length)}
          ${signalPill("Scans", scans.scans.length)}
          ${signalPill("Work jobs", work.work.length)}
          ${signalPill("Downloads", downloads.downloads.length)}
        </div>
      </div>
      <div class="toolbar">
        <button class="primary" onclick="startScan('/api/libraries/scan')">Scan Movies + TV</button>
        <button onclick="startProbe()">Probe Unprobed</button>
      </div>
      <div class="grid two">
        <div class="card"><h2>Active Sessions</h2>${sessionCards(sessions.sessions)}</div>
        <div class="card"><h2>Work</h2>${jobCards(work.work)}</div>
      </div>
      <div class="grid two">
        <div class="card"><h2>Scans</h2>${jobCards(scans.scans)}</div>
        <div class="card"><h2>Probes</h2>${jobCards(probes.probes)}</div>
      </div>
      <div class="card"><h2>Downloads</h2>${downloadCards(downloads.downloads)}</div>
    </div>`;
}

async function renderHealth() {
  const [health, review, versions] = await Promise.all([api("/api/catalog/health"), api("/api/review"), api("/api/versions")]);
  view.innerHTML = `
    <div class="stack">
      <div class="hero-strip">
        <div class="hero-copy">
          <span>Review Signal</span>
          <strong>Fix noisy files before they become a bad library.</strong>
          <p>${health.needsReview || 0} items need review, ${health.unprobed || 0} are unprobed, and ${health.unsupported || 0} may need playback help.</p>
        </div>
        <div class="signal-stack">
          ${signalPill("Needs review", health.needsReview)}
          ${signalPill("High bitrate", health.highBitrate)}
          ${signalPill("Subtitles", health.withSubtitles)}
        </div>
      </div>
      <div class="grid">
        ${metric("Needs Review", health.needsReview)}
        ${metric("Unprobed", health.unprobed)}
        ${metric("Unsupported", health.unsupported)}
        ${metric("High Bitrate", health.highBitrate)}
        ${metric("Subtitles", health.withSubtitles)}
        ${metric("Versions", versions.versions.length)}
      </div>
      <div class="grid two">
        <div class="card"><h2>Review Queue</h2>${reviewCards(review.items)}</div>
        <div class="card"><h2>Version Groups</h2>${versionGroups(versions.versions)}</div>
      </div>
    </div>`;
}

async function renderPlaybackLab() {
  const [payload, profiles, downloads] = await Promise.all([api("/api/media-sources?limit=200"), api("/api/devices/profiles"), api("/api/downloads")]);
  view.innerHTML = `<div class="stack">
    <div class="card"><h2>Client Profiles</h2>${profileCards(profiles.profiles)}</div>
    <div class="card"><h2>Playback Lab</h2>${playbackCards(payload.mediaSources)}</div>
    <div class="card"><h2>Offline Downloads</h2>${downloadCards(downloads.downloads)}</div>
  </div>`;
}

async function renderRemote() {
  const payload = await api("/api/remote/access");
  view.innerHTML = `<div class="stack">
    <div class="hero-strip">
      <div class="hero-copy"><span>Remote Access</span><strong>Your server stays yours. Vyrden helps you expose it safely.</strong><p>Vyrden shows network facts and guidance, but users stay responsible for their own VPN, reverse proxy, or port forwarding.</p></div>
      <button class="primary" onclick="lookupWan()">Detect WAN IP</button>
    </div>
    <div class="grid two">
      <div class="card"><h2>LAN addresses</h2>${payload.lanAddresses.length ? `<div class="mini-list">${payload.lanAddresses.map(url => `<a href="${url}" target="_blank"><strong>${escapeHTML(url)}</strong><span>Reachable inside your network if firewall allows it</span></a>`).join("")}</div>` : empty("No LAN address detected.")}</div>
      <div class="card"><h2>WAN address</h2><pre id="wanResult">Not checked. Click Detect WAN IP to call an external IP service.</pre></div>
    </div>
    <div class="card"><h2>Recommended paths</h2>
      <div class="remote-options">
        <div><strong>VPN mesh</strong><span>Tailscale, WireGuard, ZeroTier. Lowest operational burden for most users.</span></div>
        <div><strong>Reverse proxy</strong><span>Caddy, Nginx, Traefik. Good when the user controls DNS and TLS.</span></div>
        <div><strong>Port forward</strong><span>Works, but users must own firewall, router, and update hygiene.</span></div>
      </div>
    </div>
  </div>`;
}

async function lookupWan() {
  const target = document.getElementById("wanResult");
  target.textContent = "Checking WAN address through api.ipify.org...";
  const payload = await send("/api/remote/wan", {});
  target.textContent = JSON.stringify(payload, null, 2);
}

async function renderSettings() {
  const [settings, performance, system] = await Promise.all([api("/api/settings"), api("/api/settings/performance"), api("/api/system/status")]);
  view.innerHTML = `<div class="stack">
    <div class="hero-strip">
      <div class="hero-copy"><span>Runtime</span><strong>Local-first settings before tray and installer.</strong><p>These values are persisted locally and give the future tray app a real product runtime to control.</p></div>
      <div class="signal-stack">${signalPill("Profile", performance.profile)}${signalPill("Queues", performance.queues.length)}${signalPill("Libraries", settings.libraries.length)}</div>
    </div>
    <div class="card"><h2>Runtime Folders</h2>${runtimePathForm(settings.runtimePaths)}</div>
    <div class="card"><h2>Folder Capacity</h2>${runtimeFolders(system.disks || [])}</div>
    <div class="card"><h2>Server Config</h2>${settingsGrid(settings.config)}</div>
    <div class="card"><h2>Performance</h2><pre>${escapeHTML(JSON.stringify(performance, null, 2))}</pre></div>
  </div>`;
}

function runtimePathForm(paths = {}) {
  const fields = [
    ["data", "Database and settings"],
    ["transcode", "Transcode temp"],
    ["downloads", "Prepared downloads"],
    ["metadata", "Metadata and artwork"],
    ["cache", "Cache"],
    ["temp", "Scratch temp"],
  ];
  return `<form class="path-form" onsubmit="saveRuntimePaths(event)">
    ${fields.map(([key, label]) => `<label><span>${escapeHTML(label)}</span><input name="${key}" value="${escapeAttr(paths[key] || "")}" autocomplete="off"></label>`).join("")}
    <div class="inline-actions"><button class="primary" type="submit">Save folders</button><span class="muted">Applies fully after restart so active jobs do not lose files.</span></div>
  </form>`;
}

function settingsGrid(config) {
  return `<div class="settings-grid">${Object.entries(config).map(([key, value]) => `<div><span>${escapeHTML(key)}</span><strong>${escapeHTML(value)}</strong></div>`).join("")}</div>`;
}

async function saveRuntimePaths(event) {
  event.preventDefault();
  const data = new FormData(event.currentTarget);
  const payload = {
    dataDir: data.get("data"),
    transcodeDir: data.get("transcode"),
    downloadsDir: data.get("downloads"),
    metadataDir: data.get("metadata"),
    cacheDir: data.get("cache"),
    tempDir: data.get("temp"),
  };
  const result = await send("/api/settings", payload, "PUT");
  alert(result.restartRequired ? "Saved. Restart Vyrden to move active runtime folders." : "Saved.");
  navigate("settings");
}

async function markWatched(mediaSourceId, watched) {
  await send(`/api/playback/state/${mediaSourceId}`, { watched, progressSeconds: 0 }, "PUT");
  navigate(state.activeView);
}

async function startWork(mediaSourceId, mode) {
  await send("/api/work", { mediaSourceId, mode });
  navigate("activity");
}

async function startDownload(mediaSourceId, targetProfile) {
  await send("/api/downloads", { mediaSourceId, targetProfile });
  navigate("playback");
}

async function startScan(path) {
  await send(path, { sampleLimit: 50 });
  navigate("activity");
}

async function startProbe() {
  await send("/api/probes", { limit: 50 });
  navigate("activity");
}

function openMetadataFix(kind, id, currentTitle, year) {
  const title = prompt("Correct title", currentTitle);
  if (!title) return;
  const parsedYear = kind === "movie" ? Number(prompt("Year", year || "")) || 0 : 0;
  send("/api/metadata/match", { kind, id, title, year: parsedYear, review: false }, "PUT").then(() => navigate(state.activeView));
}

async function refreshMetadata(kind, id, title, year = 0) {
  const button = event?.currentTarget;
  const original = button?.textContent;
  if (button) {
    button.disabled = true;
    button.textContent = "Fetching...";
  }
  try {
    const result = await send("/api/metadata/refresh", { kind, id, title, year });
    if (result.warnings?.length) {
      alert(result.warnings.join("\n"));
    }
    if (kind === "movie") {
      await showMovie(id);
    } else if (kind === "series") {
      await showSeries(id);
    } else {
      await navigate(state.activeView);
    }
  } catch (error) {
    alert(error.message);
  } finally {
    if (button) {
      button.disabled = false;
      button.textContent = original;
    }
  }
}

function filterCards(value) {
  const needle = value.trim().toLowerCase();
  document.querySelectorAll("[data-filter]").forEach(card => {
    card.hidden = needle && !card.dataset.filter.toLowerCase().includes(needle);
  });
}

function dashboardCommandTitle(sessions, summary) {
  if (sessions.length) return `${sessions.length} stream${sessions.length === 1 ? "" : "s"} active`;
  if ((summary.mediaSources || 0) > 0) return "Library live";
  return "Media command centre";
}

function dashboardCommandCopy(sessions, summary, health, quality) {
  if (sessions.length) {
    const transcoding = sessions.filter(item => item.mode && item.mode !== "direct").length;
    return `${sessions.length} playback session${sessions.length === 1 ? "" : "s"} running now. ${transcoding ? `${transcoding} may need server work.` : "Everything currently looks direct-play friendly."} Keep route, progress, and background jobs visible from here.`;
  }
  if ((summary.mediaSources || 0) === 0) {
    return "Add Movies and TV libraries to start seeing live playback, file quality, probe health, storage status, review queues, and server work from one place.";
  }
  return `${summary.mediaSources || 0} files indexed across ${summary.libraries || 0} libraries. ${quality.probedPercent}% have media intelligence, ${health.needsReview || 0} need review, and ${summary.unprobed || 0} still need probing.`;
}

function dashboardTitle(featured, summary) {
  if (featured) return escapeHTML(featured.name);
  if ((summary.movies || 0) > 0) return "Movies ready for inspection.";
  if ((summary.series || 0) > 0) return "Your TV library, under control.";
  return "Your media system, at a glance.";
}

function dashboardCopy(summary, health, featured) {
  if (featured) {
    return `${summary.movies || 0} movies, ${summary.series || 0} series, and ${summary.episodes || 0} episodes indexed. Resume playback, inspect source quality, and keep server impact visible from the first screen.`;
  }
  if ((summary.movies || 0) === 0 && (summary.series || 0) > 0) {
    return `${summary.series || 0} series and ${summary.episodes || 0} episodes are indexed. Movie libraries are empty or still scanning, so Vyrden is focusing this dashboard on TV, probes, playback readiness, and storage health.`;
  }
  if ((summary.mediaSources || 0) === 0) {
    return "Add a Movies or TV library to start building the dashboard. Vyrden will surface playback decisions, source quality, storage status, probe health, and review work as soon as media is indexed.";
  }
  return `${summary.mediaSources || 0} media sources are indexed, with ${health.needsReview || 0} review items and ${summary.unprobed || 0} sources still needing probe analysis.`;
}

function metric(label, value, note = "") {
  return `<div class="card metric"><span>${escapeHTML(label)}</span><strong>${escapeHTML(value ?? 0)}</strong>${note ? `<small>${escapeHTML(note)}</small>` : ""}</div>`;
}

function signalPill(label, value) {
  return `<div class="signal-pill"><small>${escapeHTML(label)}</small><b>${escapeHTML(value ?? 0)}</b></div>`;
}

function dashboardSnapshot({ activeSessions = [], activeJobs = [], summary = {}, health = {}, quality = {}, system = {} }) {
  const memory = system.memory || {};
  const process = system.process || {};
  const disk = (system.disks || [])[0] || {};
  const serverLoad = Math.max(
    Number(system.cpu?.percent || 0),
    Number(memory.usedPercent || 0),
    process.goSysBytes && memory.totalBytes ? (process.goSysBytes / memory.totalBytes) * 100 : 0
  );
  const storageDetail = disk.freeBytes ? `${formatBytes(disk.freeBytes)} free` : "Storage pending";
  return `<div class="feature-snapshot">
    ${snapshotTile("Playback", activeSessions.length ? `${activeSessions.length} active` : "Idle", activeSessions.length ? "Live sessions are running now" : "Ready for the next stream", activeSessions.length ? "route" : "good")}
    ${snapshotTile("Probe Health", `${quality.probedPercent || 0}%`, `${summary.unprobed || 0} files still need probing`, summary.unprobed ? "warn" : "good")}
    ${snapshotTile("Server Load", `${Math.round(serverLoad || 0)}%`, `${activeJobs.length} background job${activeJobs.length === 1 ? "" : "s"}`, serverLoad > 75 || activeJobs.length ? "warn" : "good")}
    ${snapshotTile("Storage", disk.usedPercent ? `${Math.round(disk.usedPercent)}% used` : "Ready", storageDetail, disk.usedPercent > 85 ? "warn" : "route")}
  </div>`;
}

function snapshotTile(label, value, note, tone = "") {
  return `<div class="snapshot-tile ${tone}">
    <span>${escapeHTML(label)}</span>
    <strong>${escapeHTML(value)}</strong>
    <small>${escapeHTML(note)}</small>
  </div>`;
}

function sourceQuality(items = []) {
  const probed = items.filter(item => item.probed);
  const codecs = countBy(probed, item => item.videoCodec || "unknown");
  const containers = countBy(probed, item => item.container || item.extension || "unknown");
  const totalSize = items.reduce((sum, item) => sum + Number(item.sizeBytes || 0), 0);
  const highBitrate = probed.filter(item => Number(item.bitrate || 0) > 40000000).length;
  const subtitles = probed.filter(item => Number(item.subtitleStreams || 0) > 0).length;
  const fourK = probed.filter(item => Number(item.width || 0) >= 3800 || Number(item.height || 0) >= 2000).length;
  const remuxLikely = probed.filter(item => Number(item.bitrate || 0) > 65000000).length;
  return {
    total: items.length,
    probed: probed.length,
    probedPercent: items.length ? Math.round((probed.length / items.length) * 100) : 0,
    totalSize,
    highBitrate,
    subtitles,
    fourK,
    remuxLikely,
    topCodec: topCount(codecs),
    topContainer: topCount(containers),
  };
}

function countBy(items, keyFn) {
  return items.reduce((counts, item) => {
    const key = keyFn(item);
    counts[key] = (counts[key] || 0) + 1;
    return counts;
  }, {});
}

function topCount(counts) {
  const entries = Object.entries(counts).sort((a, b) => b[1] - a[1]);
  return entries[0] ? { label: entries[0][0], count: entries[0][1] } : { label: "pending", count: 0 };
}

function isActiveJob(item = {}) {
  const status = String(item.status || "").toLowerCase();
  return status === "queued" || status === "running";
}

function playingNow(items = []) {
  if (!items.length) {
    return `<div class="now-idle">
      <strong>Playback idle</strong>
      <span>Start a movie or episode and this panel becomes a live session monitor with device, route, progress, and server impact.</span>
    </div>
    <div class="kv">
      <div><span>Status</span><span>Ready</span></div>
      <div><span>Server impact</span><span>Idle</span></div>
      <div><span>Route</span><span>LAN</span></div>
    </div>`;
  }
  return `<div class="session-list">${items.slice(0, 4).map(session => {
    const percent = session.durationSeconds ? Math.min(100, Math.round((session.progressSeconds / session.durationSeconds) * 100)) : 0;
    return `<article class="session-card">
      <div><strong>${escapeHTML(session.deviceId || "Device")}</strong><span>${escapeHTML(session.mode || "direct")} - ${escapeHTML(session.status || "playing")}</span></div>
      <b>${percent}%</b>
      <i style="--progress:${Math.max(3, percent)}%"></i>
      <small>${formatDuration(session.progressSeconds)} / ${formatDuration(session.durationSeconds)}</small>
    </article>`;
  }).join("")}</div>`;
}

function fileIntelligence(quality, summary, health) {
  return `<div class="quality-grid">
    ${qualityTile("Probed", `${quality.probedPercent}%`, `${quality.probed} of ${quality.total || summary.mediaSources || 0} sampled`)}
    ${qualityTile("4K / UHD", quality.fourK, "High resolution sources")}
    ${qualityTile("High Bitrate", health.highBitrate || quality.highBitrate, "Likely heavy streams")}
    ${qualityTile("Subtitles", health.withSubtitles || quality.subtitles, "Files with subtitle tracks")}
    ${qualityTile("Top Codec", quality.topCodec.label, `${quality.topCodec.count} sampled`)}
    ${qualityTile("Container", quality.topContainer.label, `${quality.topContainer.count} sampled`)}
  </div>`;
}

function qualityTile(label, value, note) {
  return `<div class="quality-tile"><span>${escapeHTML(label)}</span><strong>${escapeHTML(value ?? 0)}</strong><small>${escapeHTML(note || "")}</small></div>`;
}

function operationsPanel(scans = [], probes = [], work = [], downloads = []) {
  const active = [...scans, ...probes, ...work, ...downloads].filter(isActiveJob);
  if (!active.length) {
    return `<div class="ops-clear"><strong>Background queues clear</strong><span>Scanning, probing, remuxing, and transcoding are not competing with playback.</span></div>`;
  }
  return `<div class="ops-list">${active.slice(0, 5).map(job => `<article>
    <strong>${escapeHTML(job.kind || job.mode || "work")}</strong>
    <span>${escapeHTML(job.status || "running")}</span>
    <small>${escapeHTML(job.lastPath || job.outputPath || job.id || "")}</small>
  </article>`).join("")}</div>`;
}

function attentionPanel(summary, health, versions = []) {
  const items = [
    { label: "Needs metadata review", value: health.needsReview || 0, tone: health.needsReview ? "warn" : "good" },
    { label: "Unprobed files", value: summary.unprobed || 0, tone: summary.unprobed ? "warn" : "good" },
    { label: "Unsupported sources", value: health.unsupported || 0, tone: health.unsupported ? "warn" : "good" },
    { label: "Duplicate versions", value: versions.length || 0, tone: versions.length ? "warn" : "good" },
  ];
  return `<div class="attention-list">${items.map(item => `<div class="${item.tone}">
    <span>${escapeHTML(item.label)}</span>
    <strong>${escapeHTML(item.value)}</strong>
  </div>`).join("")}</div>`;
}

function hardwarePanel(system = {}) {
  const memory = system.memory || {};
  const process = system.process || {};
  return `<div class="hardware-list">
    ${usageBar("CPU", system.cpu?.percent || 0, `${system.cpu?.cores || 0} logical cores`)}
    ${usageBar("Memory", memory.usedPercent || 0, `${formatBytes(memory.usedBytes)} / ${formatBytes(memory.totalBytes)}`)}
    ${usageBar("Vyrden heap", process.goSysBytes && memory.totalBytes ? (process.goSysBytes / memory.totalBytes) * 100 : 0, `${formatBytes(process.goAllocBytes)} allocated, ${process.goroutines || 0} tasks`)}
  </div>`;
}

function runtimeFolders(disks = []) {
  if (!disks.length) return empty("No runtime folders reported.");
  return `<div class="folder-list">${disks.map(disk => `<div class="folder-row ${disk.error || !disk.writable ? "warn" : ""}">
    <div><strong>${escapeHTML(folderLabel(disk.name))}</strong><span>${escapeHTML(disk.path)}${disk.sharedWithData ? " - install/data drive" : ""}</span></div>
    <div>${usageBar("Disk", disk.usedPercent || 0, `${formatBytes(disk.freeBytes)} free`)}</div>
  </div>`).join("")}</div>`;
}

function usageBar(label, percent, detail) {
  const value = Math.max(0, Math.min(100, Number(percent || 0)));
  const tone = value > 90 ? "bad" : value > 75 ? "warn" : "good";
  return `<div class="usage ${tone}">
    <div><span>${escapeHTML(label)}</span><strong>${Math.round(value)}%</strong></div>
    <i><b style="width:${value}%"></b></i>
    <small>${escapeHTML(detail || "")}</small>
  </div>`;
}

function folderLabel(value = "") {
  return { data: "Database", transcode: "Transcode temp", downloads: "Prepared downloads", metadata: "Metadata cache", cache: "App cache", temp: "Scratch temp" }[value] || value;
}

function movieOverview(movie = {}) {
  if (movie.metadata?.overview) return escapeHTML(movie.metadata.overview);
  if (movie.needsReview) return "This item needs metadata review before Vyrden can fully trust its match.";
  return "Choose a version, confirm the playback route, then play directly from your local library. The selected source stays explicit so users understand quality, route, and server impact before pressing play.";
}

function metadataBadges(metadata) {
  if (!metadata) return `<span class="badge warn">Metadata pending</span>`;
  return `<span class="badge route">${escapeHTML(providerLabel(metadata.provider))}</span><span class="badge">${metadataConfidenceLabel(metadata)}</span>`;
}

function ratingsRail(metadata = {}) {
  const ratings = normalizeRatings(metadata);
  return `<div class="ratings-rail">
    ${ratings.map(item => `<div class="rating-tile ${item.pending ? "pending" : ""}">
      <span>${escapeHTML(item.label)}</span>
      <strong>${escapeHTML(item.value)}</strong>
      <small>${escapeHTML(item.note)}</small>
    </div>`).join("")}
  </div>`;
}

function normalizeRatings(metadata = {}) {
  const raw = metadata?.ratings || {};
  const lookup = key => raw[key] ?? metadata?.[key];
  return [
    ratingItem("IMDb", lookup("imdb"), "User score"),
    ratingItem("Rotten Critics", lookup("rottenTomatoesCritics"), "Tomatometer"),
    ratingItem("Rotten Audience", lookup("rottenTomatoesAudience"), "Audience"),
    ratingItem("TMDB", lookup("tmdb"), "Community"),
    ratingItem("Metacritic", lookup("metacritic"), "Critic score"),
  ];
}

function ratingItem(label, value, note) {
  if (value === undefined || value === null || value === "") return { label, value: "Pending", note: "Not fetched yet", pending: true };
  return { label, value: formatRating(value), note, pending: false };
}

function formatRating(value) {
  if (value && typeof value === "object") {
    if (value.displayValue) return value.displayValue;
    if (value.value !== undefined) return formatRating(value.value);
  }
  if (typeof value === "string") return value;
  const number = Number(value);
  if (!Number.isFinite(number)) return String(value);
  if (number <= 1) return `${Math.round(number * 100)}%`;
  if (number <= 10) return `${number.toFixed(1)}/10`;
  return `${Math.round(number)}%`;
}

function metadataSummary(metadata) {
  if (!metadata) return "No metadata record has been selected yet. Filename matching will populate the first local record after scan.";
  const source = providerLabel(metadata.provider);
  const confidence = metadataConfidenceLabel(metadata);
  return `${source} is currently selected with ${confidence.toLowerCase()}.`;
}

function metadataExternalID(metadata = {}) {
  if (metadata.externalId) return metadata.externalId;
  const ids = metadata.externalIds || {};
  if (ids.imdb) return ids.imdb;
  if (ids.tmdb) return `TMDB ${ids.tmdb}`;
  return "Local only";
}

function metadataConfidenceLabel(metadata) {
  if (!metadata) return "0% confidence";
  return `${Math.round(Number(metadata.confidence || 0) * 100)}% confidence`;
}

function providerLabel(provider = "") {
  const labels = {
    filename: "Filename",
    manual: "Manual",
    nfo: "Local NFO",
    tmdb: "TMDB",
    tvdb: "TVDB",
    omdb: "OMDb",
    none: "None",
    pending: "Pending",
  };
  return labels[provider] || provider || "Unknown";
}

function formatBytes(value) {
  const bytes = Number(value || 0);
  if (!bytes) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const index = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
  const amount = bytes / Math.pow(1024, index);
  return `${amount >= 10 || index === 0 ? Math.round(amount) : Math.round(amount * 10) / 10} ${units[index]}`;
}

function formatDuration(value) {
  const total = Math.max(0, Math.round(Number(value || 0)));
  if (!total) return "0:00";
  const hours = Math.floor(total / 3600);
  const minutes = Math.floor((total % 3600) / 60);
  const seconds = total % 60;
  if (hours) return `${hours}:${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`;
  return `${minutes}:${String(seconds).padStart(2, "0")}`;
}

function mediaCards(items = []) {
  if (!items.length) {
    return `<div class="poster-shelf">${["Movies", "TV Shows", "4K Sources", "Direct Play", "New Scans", "Review Queue"].map((label, index) => shelfPoster(label, index)).join("")}</div>`;
  }
  return `<div class="poster-shelf">${items.slice(0, 6).map((item, index) => `<a class="shelf-poster" href="/play/${item.mediaSourceId}" target="_blank" data-initial="${escapeAttr(initials(item.name))}" style="--tone:${index % 6}">
    <strong>${escapeHTML(item.name)}</strong>
    <span>${Math.round((item.percent || 0) * 100)}% watched - ${escapeHTML(item.kind)}</span>
    <i style="--progress:${Math.max(4, Math.round((item.percent || 0) * 100))}%"></i>
  </a>`).join("")}</div>`;
}

function shelfPoster(label, index) {
  return `<button class="shelf-poster" type="button" onclick="navigate('${index < 1 ? "movies" : index < 2 ? "tv" : "playback"}')" data-initial="${escapeAttr(initials(label))}" style="--tone:${index}">
    <strong>${escapeHTML(label)}</strong>
    <span>${index < 2 ? "Browse library" : "Open signal"}</span>
  </button>`;
}

function reviewCards(items = []) {
  if (!items.length) return empty("No review items.");
  return `<div class="version-grid">${items.map(item => `<article class="version-card">
    <div><strong>${escapeHTML(item.title)}</strong><small>${escapeHTML(item.kind)} - ${escapeHTML(item.reviewReason)}</small></div>
    <div class="inline-actions">
      <button onclick="openMetadataFix('${item.kind}','${item.id}','${escapeAttr(item.title)}',0)">Fix Match</button>
      <button onclick="refreshMetadata('${item.kind}','${item.id}','${escapeAttr(item.title)}',0)">Fetch Metadata</button>
    </div>
  </article>`).join("")}</div>`;
}

function versionGroups(items = [], totalTitles = 0) {
  if (!items.length) return empty(totalTitles ? "No duplicate version groups found." : "No version groups yet. Scan libraries to build movie and TV source intelligence.");
  return `<div class="version-grid">${items.map(item => `<article class="version-card">
    <div><strong>${escapeHTML(item.title)}</strong><small>${escapeHTML(item.kind)} - ${item.versionCount} versions</small></div>
  </article>`).join("")}</div>`;
}

function sessionCards(items = []) {
  if (!items.length) return empty("No active sessions.");
  return `<div class="version-grid">${items.map(item => `<article class="version-card">
    <div><strong>${escapeHTML(item.deviceId)}</strong><small>${escapeHTML(item.status)} - ${Math.round(item.progressSeconds || 0)}s</small></div>
    <small>${escapeHTML(item.mediaSourceId)}</small>
  </article>`).join("")}</div>`;
}

function jobCards(items = []) {
  if (!items || !items.length) return empty("No jobs yet. Scans, probes, remuxes, and transcodes will show here when they run.");
  return `<div class="version-grid">${items.slice(0, 8).map(item => `<article class="version-card">
    <div><strong>${escapeHTML(item.status || "queued")}</strong><small>${escapeHTML(item.id || "")}</small></div>
    <small>${escapeHTML(item.lastPath || item.outputPath || item.mediaFiles || item.completed || "")}</small>
  </article>`).join("")}</div>`;
}

function profileCards(items = []) {
  if (!items.length) return empty("No client profiles.");
  return `<div class="version-grid">${items.map(item => `<article class="version-card">
    <div><strong>${escapeHTML(item.name)}</strong><small>${escapeHTML(item.containers.join(", "))}</small></div>
    <small>${escapeHTML(item.videoCodecs.join(", "))}</small>
  </article>`).join("")}</div>`;
}

function playbackCards(items = []) {
  if (!items.length) return empty("No media sources.");
  return `<div class="version-grid">${items.map(item => `<article class="version-card">
    <div><strong>${escapeHTML(item.name)}</strong><small>${escapeHTML(item.kind)} - ${item.probed ? "probed" : "needs probe"}</small></div>
    <div class="inline-actions">
      <a class="button" href="/api/playback/decision?mediaSourceId=${item.id}&clientProfile=web" target="_blank">Decision</a>
      <button onclick="startWork('${item.id}','remux')">Remux</button>
      <button onclick="startWork('${item.id}','transcode')">Transcode</button>
      <button onclick="startDownload('${item.id}','balanced')">Download</button>
      <a class="button primary" href="/play/${item.id}" target="_blank">Play</a>
    </div>
  </article>`).join("")}</div>`;
}

function downloadCards(items = []) {
  if (!items.length) return empty("No prepared downloads yet. Start one from a media source in Playback Lab.");
  return `<div class="version-grid">${items.map(item => `<article class="version-card">
    <div><strong>${escapeHTML(item.targetProfile)}</strong><small>${escapeHTML(item.status)} - ${escapeHTML(item.mediaSourceId)}</small></div>
    <div class="inline-actions">
      ${item.status === "completed" ? `<a class="button primary" href="/api/downloads/${item.id}/file" target="_blank">Save File</a>` : ""}
      <a class="button" href="/api/downloads/${item.id}" target="_blank">Details</a>
    </div>
  </article>`).join("")}</div>`;
}

function table(headers, rows) {
  if (!rows || rows.length === 0) return empty("No items yet.");
  return `<table><thead><tr>${headers.map(header => `<th>${escapeHTML(header)}</th>`).join("")}</tr></thead><tbody>${rows.map(row => `<tr>${row.map(cell => `<td>${cell}</td>`).join("")}</tr>`).join("")}</tbody></table>`;
}

function empty(text) {
  return `<p class="muted">${escapeHTML(text)}</p>`;
}

function link(href, text) {
  return `<a href="${href}">${escapeHTML(text)}</a>`;
}

function escapeAttr(value) {
  return escapeHTML(value).replace(/`/g, "&#096;");
}

function artworkURL(kind, id) {
  return `/api/artwork/${encodeURIComponent(kind)}/${encodeURIComponent(id)}?style=neutral`;
}

function initials(value) {
  const words = String(value || "V").trim().split(/\s+/).filter(Boolean);
  return words.slice(0, 2).map(word => word[0]).join("").toUpperCase() || "V";
}

function escapeHTML(value) {
  return String(value ?? "").replace(/[&<>"']/g, char => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "\"": "&quot;", "'": "&#039;" }[char]));
}

let dashboardRefreshTimer = 0;
function refreshLiveViews() {
  if (!["dashboard", "activity", "health"].includes(state.activeView)) return;
  clearTimeout(dashboardRefreshTimer);
  dashboardRefreshTimer = setTimeout(() => navigate(state.activeView), 180);
}

const events = new EventSource("/api/events");
for (const name of ["scan.queued", "scan.running", "scan.completed", "scan.failed", "probe.queued", "probe.running", "probe.completed", "probe.failed", "transcode.queued", "transcode.running", "transcode.completed", "transcode.failed", "download.queued", "download.running", "download.completed", "download.failed", "session.started", "session.updated", "session.stopped", "playback.state.updated", "metadata.updated"]) {
  events.addEventListener(name, () => {
    refreshLiveViews();
  });
}

refreshShell().catch(() => {});
navigate("dashboard");
