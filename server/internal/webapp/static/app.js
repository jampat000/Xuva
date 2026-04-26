const state = { activeView: "dashboard", activeSession: "", playbackSelections: {}, movieWall: [], movieFilter: "all", movieSort: "title" };

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
  return { dashboard: "Home", movies: "Movies", tv: "TV", libraries: "Libraries", activity: "Activity", health: "Review", playback: "Source Inspector", remote: "Remote Access", settings: "Settings" }[name] || name;
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
  const [summary, libraries, scans, probes, work, downloads, recent, sessions, health, versions, sources, system, settings] = await Promise.all([
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
    api("/api/settings"),
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
            <h1 data-live="dashboard-title">${dashboardCommandTitle(activeSessions, summary)}</h1>
            <p class="lead" data-live="dashboard-copy">${dashboardCommandCopy(activeSessions, summary, health, quality)}</p>
            <div class="meta-line">
              <span class="badge">${summary.mediaSources || 0} sources</span>
              <span class="badge">${summary.libraries || 0} libraries</span>
              <span class="badge ${summary.unprobed ? "warn" : "good"}">${summary.unprobed || 0} unprobed</span>
              <span class="badge ${directPlayable >= 90 ? "good" : "warn"}">${directPlayable}% playable</span>
              <span class="badge route">Live</span>
            </div>
            <div class="actions">
              <button class="primary" onclick="navigate('${movieCount ? "movies" : "tv"}')">${movieCount ? "Movies" : "TV"}</button>
              <button onclick="navigate('activity')">Activity</button>
              <button onclick="navigate('health')">Review ${reviewCount}</button>
            </div>
          </div>
          <div data-live="dashboard-snapshot">${dashboardSnapshot({ activeSessions, activeJobs, summary, health, quality, system })}</div>
        </article>

        <aside class="panel pad">
          <div class="panel-title"><strong>Playing Now</strong><span class="badge route" data-live="playing-now-status">${activeSessions.length ? "Live" : "Idle"}</span></div>
          <div data-live="playing-now">${playingNow(activeSessions)}</div>
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
          <div class="panel-title"><strong>File Intelligence</strong><button onclick="navigate('movies')">Browse</button></div>
          ${fileIntelligence(quality, summary, health)}
        </div>
        <div class="panel pad">
          <div class="panel-title"><strong>Operations</strong><span class="badge ${activeJobs.length ? "warn" : "good"}" data-live="operations-status">${activeJobs.length ? `${activeJobs.length} running` : "Clear"}</span></div>
          <div data-live="operations">${operationsPanel(scanJobs, probeJobs, workJobs, downloadJobs)}</div>
        </div>
        <div class="panel pad">
          <div class="panel-title"><strong>Needs Attention</strong><div class="inline-actions"><button onclick="startProbe()">Probe</button><button onclick="navigate('health')">Review</button></div></div>
          ${attentionPanel(summary, health, versions.versions || [])}
        </div>
        <div class="panel pad">
          <div class="panel-title"><strong>Hardware</strong><span class="badge route" data-live="hardware-cores">${system.cpu.cores} cores</span></div>
          <div data-live="hardware">${hardwarePanel(system)}</div>
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
        <div class="panel pad"><div class="panel-title"><strong>Library Automation</strong><button onclick="navigate('settings')">Schedule</button></div>${automationStatus(settings.config || {}, scanJobs, probeJobs)}</div>
      </section>

      <section class="dashboard-grid">
        <div class="panel pad"><div class="panel-title"><strong>Version Intelligence</strong><button onclick="navigate('health')">Review</button></div>${versionGroups(versions.versions, totalTitles)}</div>
        <div class="panel pad"><div class="panel-title"><strong>Runtime Folders</strong><button onclick="navigate('settings')">Move</button></div>${runtimeFolders(system.disks || [])}</div>
      </section>

      <section class="panel pad">
        <div class="panel-title"><strong>Live Signals</strong><span class="live-stamp" data-live="updated-at">Updated ${new Date().toLocaleTimeString()}</span></div>
        <div class="signal-stack" data-live="live-signals">${signalPill("Sessions", activeSessions.length)}${signalPill("Jobs", activeJobs.length)}${signalPill("Downloads", downloadJobs.length)}${signalPill("Review", reviewCount)}${signalPill("CPU", `${Math.round(system.cpu.percent || 0)}%`)}</div>
      </section>
    </div>`;
}

async function renderMovies() {
  const payload = await api("/api/movies?limit=500");
  state.movieWall = payload.movies || [];
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
        <button onclick="refreshMetadataBatch('movie')">Refresh Metadata</button>
        <input class="search" id="movieSearch" placeholder="Filter movies" oninput="applyMovieWallControls()">
      </div>
      <div class="control-panel">
        <div class="filter-group" aria-label="Movie filters">
          ${movieFilterButton("all", "All")}
          ${movieFilterButton("review", "Needs Review")}
          ${movieFilterButton("metadata", "Metadata Pending")}
          ${movieFilterButton("versions", "Multiple Versions")}
        </div>
        <div class="filter-group" aria-label="Movie sorting">
          ${movieSortButton("title", "Title")}
          ${movieSortButton("year", "Year")}
          ${movieSortButton("versions", "Versions")}
          ${movieSortButton("review", "Review")}
        </div>
      </div>
      <div class="poster-grid" id="movieWall">${movieWall(state.movieWall)}</div>
    </div>`;
}

function movieWall(items = []) {
  return items.length ? items.map(movieCard).join("") : empty("No movies match the current view.");
}

function movieFilterButton(filter, label) {
  return `<button class="${state.movieFilter === filter ? "selected" : ""}" data-movie-filter="${filter}" onclick="setMovieFilter('${filter}')">${escapeHTML(label)}</button>`;
}

function movieSortButton(sort, label) {
  return `<button class="${state.movieSort === sort ? "selected" : ""}" data-movie-sort="${sort}" onclick="setMovieSort('${sort}')">${escapeHTML(label)}</button>`;
}

function movieCard(item) {
  return `<article class="poster-card" data-filter="${escapeAttr(item.title)} ${item.year || ""}" data-initial="${escapeAttr(initials(item.title))}" data-review="${item.needsReview ? "true" : "false"}" data-metadata="${item.metadata ? "true" : "false"}" data-versions="${item.versionCount || 0}">
    <img alt="" src="${artworkURL("movie", item.id)}" loading="lazy">
    <button class="poster-action" onclick="showMovie('${item.id}')">
      <span class="poster-title">${escapeHTML(item.title)}</span>
      <small>${item.year || "Unknown year"} - ${item.versionCount || 0} version${item.versionCount === 1 ? "" : "s"}</small>
      ${item.needsReview ? `<em>Review</em>` : ""}
    </button>
  </article>`;
}

function setMovieFilter(filter) {
  state.movieFilter = filter;
  document.querySelectorAll("[data-movie-filter]").forEach(button => button.classList.toggle("selected", button.dataset.movieFilter === filter));
  applyMovieWallControls();
}

function setMovieSort(sort) {
  state.movieSort = sort;
  document.querySelectorAll("[data-movie-sort]").forEach(button => button.classList.toggle("selected", button.dataset.movieSort === sort));
  applyMovieWallControls();
}

function applyMovieWallControls() {
  const target = document.getElementById("movieWall");
  if (!target) return;
  const needle = (document.getElementById("movieSearch")?.value || "").trim().toLowerCase();
  let items = [...state.movieWall];
  if (needle) {
    items = items.filter(item => `${item.title || ""} ${item.year || ""}`.toLowerCase().includes(needle));
  }
  if (state.movieFilter === "review") items = items.filter(item => item.needsReview);
  if (state.movieFilter === "metadata") items = items.filter(item => !item.metadata);
  if (state.movieFilter === "versions") items = items.filter(item => Number(item.versionCount || 0) > 1);
  items.sort(movieSorter(state.movieSort));
  target.innerHTML = movieWall(items);
}

function movieSorter(sort) {
  if (sort === "year") return (a, b) => Number(b.year || 0) - Number(a.year || 0) || String(a.title || "").localeCompare(String(b.title || ""));
  if (sort === "versions") return (a, b) => Number(b.versionCount || 0) - Number(a.versionCount || 0) || String(a.title || "").localeCompare(String(b.title || ""));
  if (sort === "review") return (a, b) => Number(Boolean(b.needsReview)) - Number(Boolean(a.needsReview)) || String(a.title || "").localeCompare(String(b.title || ""));
  return (a, b) => String(a.title || "").localeCompare(String(b.title || ""));
}

async function showMovie(id) {
  const movie = await api(`/api/movies/${id}`);
  const versionModels = await Promise.all(movie.versions.map(loadVersionModel));
  const rows = versionModels.map((model, index) => versionCard(model, index === 0));
  const selected = versionModels[0];
  const tracks = selected?.tracks || { audioTracks: [], subtitleTracks: [] };
  if (selected) seedPlaybackSelection(selected.mediaSourceId, tracks);
  viewTitle.textContent = movie.title;
  view.innerHTML = `
    <div class="detail-command">
      <div class="detail-backdrop"><img alt="" src="${artworkURL("movie", movie.id, "backdrop")}"></div>
      <section class="detail-main">
        <button class="back-button" onclick="navigate('movies')">Back to Movies</button>
        <div class="eyebrow">Featured from your library</div>
        <h1>${escapeHTML(movie.title)}</h1>
        <div class="meta-line">
          <span class="badge">${movie.year || "Unknown year"}</span>
          <span class="badge">${movie.versionCount} version${movie.versionCount === 1 ? "" : "s"}</span>
          <span class="badge ${movie.needsReview ? "warn" : "good"}">${movie.needsReview ? "Needs Review" : "Matched"}</span>
          <span class="badge route">${selected ? escapeHTML(playbackReadinessLabel(selected.decision)) : "No Source"}</span>
          ${metadataBadges(movie.metadata)}
        </div>
        <p class="lead">${movieOverview(movie)}</p>
        ${ratingsRail(movie.metadata)}
        <div class="actions detail-hero-actions">
          ${selected ? `<a class="button primary focusable" href="/play/${selected.mediaSourceId}" target="_blank">Resume</a>` : ""}
          ${selected ? `<a class="button focusable" href="/play/${selected.mediaSourceId}?start=0" target="_blank">Play From Start</a>` : ""}
          ${selected ? `<button class="button focusable" onclick="markWatched('${selected.mediaSourceId}', true)">Watched</button>` : ""}
          <button class="button focusable" onclick="refreshMetadata('movie','${movie.id}','${escapeAttr(movie.title)}',${movie.year || 0})">Fetch Ratings</button>
          ${movie.needsReview ? `<button class="button focusable" onclick="openMetadataFix('movie','${movie.id}','${escapeAttr(movie.title)}',${movie.year || 0})">Fix Match</button>` : ""}
        </div>
        <section class="section">
          <div class="section-head">
            <div class="section-title">Source file</div>
          </div>
          <div class="detail-source-workbench">
            <div class="version-grid">${rows.join("") || empty("No playable versions found.")}</div>
            ${movieDetailRail(movie, selected)}
          </div>
        </section>
        <section class="section">
          <div class="section-head">
            <div class="section-title">Audio & Subtitles</div>
          </div>
          <div class="track-grid">
            <div class="track-panel"><strong>Audio</strong>${audioTrackRows(tracks.audioTracks, selected?.mediaSourceId)}</div>
            <div class="track-panel"><strong>Subtitles</strong>${subtitleTrackRows(tracks.subtitleTracks, selected?.mediaSourceId)}</div>
          </div>
        </section>
      </section>
    </div>`;
  if (selected) updatePlaybackForecast(selected.mediaSourceId);
}

async function loadVersionModel(version = {}) {
  const forecastProfiles = ["web", "android-tv", "apple-tv", "chromecast"];
  const [state, source, tracks, decision, deviceDecisions, performance] = await Promise.all([
    api(`/api/playback/state/${version.mediaSourceId}`).catch(() => ({})),
    api(`/api/media-sources/${version.mediaSourceId}`).catch(() => ({})),
    api(`/api/media-sources/${version.mediaSourceId}/tracks`).catch(() => ({ audioTracks: [], subtitleTracks: [] })),
    api(`/api/playback/decision?mediaSourceId=${version.mediaSourceId}&clientProfile=web`).catch(() => ({})),
    Promise.all(forecastProfiles.map(profile => api(`/api/playback/decision?mediaSourceId=${version.mediaSourceId}&clientProfile=${profile}`).catch(() => ({ clientProfile: profile })))),
    performanceSnapshot(),
  ]);
  return { ...version, state, source, tracks, decision, deviceDecisions, hardware: hardwareTranscodeProfile(performance), policy: playbackPolicyProfile(performance.playbackPolicy) };
}

function versionCard(model, selected) {
  const version = model || {};
  const state = version.state || {};
  const source = version.source || {};
  const decision = version.decision || {};
  const watched = state.watched ? "Watched" : state.progressSeconds > 0 ? `${Math.round((state.percent || 0) * 100)}% watched` : "Unplayed";
  const routeTone = decision.mode ? playbackReadinessLabel(decision) : watched;
  const sourceFacts = [
    source.container ? String(source.container).toUpperCase() : "Container pending",
    source.width && source.height ? `${source.width}x${source.height}` : "Resolution pending",
    source.bitrate ? formatBitrate(source.bitrate) : "Bitrate pending",
  ].filter(Boolean).join(" - ");
  const hardware = version.hardware || {};
  const policy = version.policy || playbackPolicyProfile();
  const tabName = `source-tabs-${String(version.mediaSourceId || "source").replace(/[^a-zA-Z0-9_-]/g, "")}`;
  return `<article class="version source-card ${selected ? "is-selected" : ""}">
    <div class="source-card-head">
      <div>
        <strong>${escapeHTML(version.qualityLabel || sourceQualityLabel(source) || "Source")}</strong>
        <span>${escapeHTML(version.relPath || "Local source")}</span>
      </div>
      <em class="${source.probed ? "ready" : "pending"}">${escapeHTML(source.probed ? "Checked" : "Check needed")}</em>
    </div>
    <div class="source-tabset">
      ${sourceTabInput(tabName, "play", "Play", true)}
      ${sourceTabInput(tabName, "details", "Details")}
      ${sourceTabInput(tabName, "devices", "Devices")}
      ${sourceTabInput(tabName, "optimize", "Optimize")}
      <div class="source-tabs">
        ${sourceTabLabel(tabName, "play", "Play")}
        ${sourceTabLabel(tabName, "details", "Details")}
        ${sourceTabLabel(tabName, "devices", "Devices")}
        ${sourceTabLabel(tabName, "optimize", "Optimize")}
      </div>
      <section class="source-tab-panel source-tab-panel-play">
        <div class="source-card-grid primary">
          ${sourceCardFact("Can it play?", routeTone)}
          ${sourceCardFact("PC work needed", playbackEffortLabel(decision, hardware))}
          ${sourceCardFact("Current policy", policyDecisionLabel(policy, decision))}
        </div>
        <p>${escapeHTML(playbackReason(decision, hardware))}</p>
        ${policyImpactNote(policy, decision)}
        ${hardwareImpactNote(decision, hardware)}
        <div class="inline-actions source-card-actions">
          <a class="button primary" href="/play/${version.mediaSourceId}" target="_blank">${state.progressSeconds > 5 && !state.watched ? "Resume" : "Play"}</a>
          <a class="button" href="/play/${version.mediaSourceId}?start=0" target="_blank">Start Over</a>
          ${source.probed ? "" : `<button onclick="probeSource('${version.mediaSourceId}')">Check file</button>`}
          <button onclick="openSourceInspector('${version.mediaSourceId}')">File details</button>
          <button onclick="markWatched('${version.mediaSourceId}', true)">Watched</button>
          <button onclick="markWatched('${version.mediaSourceId}', false)">Unwatched</button>
        </div>
      </section>
      <section class="source-tab-panel source-tab-panel-details">
        <div class="source-card-grid">
          ${sourceCardFact("File size", formatBytes(version.sizeBytes))}
          ${sourceCardFact("Video file details", sourceFacts)}
          ${sourceCardFact("Probe status", source.probed ? "Checked" : "Check needed")}
          ${sourceCardFact("Storage path", version.relPath || source.relPath || "Local source")}
        </div>
      </section>
      <section class="source-tab-panel source-tab-panel-devices">
        ${deviceForecast(version.deviceDecisions, hardware) || empty("No player compatibility forecast yet.")}
      </section>
      <section class="source-tab-panel source-tab-panel-optimize">
        <div class="source-optimize-copy">
          <strong>Create optimized version</strong>
          <span>Optional stored versions are only created when you choose them. Originals are not modified.</span>
        </div>
        <div class="download-plan">
          ${downloadPlan("Original", formatBytes(version.sizeBytes), "No new file. Use the existing source.", version.mediaSourceId, "original")}
          ${downloadPlan("Remote Optimized", "1080p", "Creates a smaller stored version for remote playback.", version.mediaSourceId, "balanced")}
          ${downloadPlan("Travel Optimized", "720p", "Creates the smallest portable stored version.", version.mediaSourceId, "travel")}
        </div>
      </section>
    </div>
  </article>`;
}

function sourceCardFact(label, value) {
  return `<div><span>${escapeHTML(label)}</span><strong>${escapeHTML(value || "Pending")}</strong></div>`;
}

function sourceTabInput(group, id, label, checked = false) {
  return `<input class="source-tab-input" type="radio" name="${escapeAttr(group)}" id="${escapeAttr(group)}-${escapeAttr(id)}" ${checked ? "checked" : ""} aria-label="${escapeAttr(label)}">`;
}

function sourceTabLabel(group, id, label) {
  return `<label for="${escapeAttr(group)}-${escapeAttr(id)}">${escapeHTML(label)}</label>`;
}

async function performanceSnapshot() {
  if (state.performanceSnapshot) return state.performanceSnapshot;
  state.performanceSnapshot = await api("/api/settings/performance").catch(() => ({}));
  return state.performanceSnapshot;
}

function hardwareTranscodeProfile(performance = {}) {
  const hardware = performance?.hardwareAcceleration || {};
  const encoders = Array.isArray(hardware.encoders) ? hardware.encoders : [];
  const gpuWorkers = Number(hardware.gpuWorkers ?? performance?.limits?.gpuWorkers ?? 0);
  return {
    gpuWorkers,
    encoders,
    available: !!hardware.available,
    configured: gpuWorkers > 0 && !!hardware.available,
    unlockState: hardware.unlockState || "planned",
    status: hardware.status || "unknown",
  };
}

function playbackPolicyProfile(policy = {}) {
  const id = policy.id || "original_only";
  const label = policy.label || playbackPolicyName(id);
  return { id, label, description: policy.description || "" };
}

function playbackPolicyName(value = "original_only") {
  return {
    original_only: "Original Only",
    light: "Light Compatibility",
    full: "Full Compatibility",
    cinema: "Cinema Server",
  }[String(value)] || "Original Only";
}

function policyAllowsDecision(policy = {}, decision = {}) {
  const mode = String(decision.mode || "").toLowerCase();
  if (!mode || mode === "direct play" || mode === "decision deferred") return true;
  const id = policy.id || "original_only";
  if (id === "light") return mode === "remux" || mode === "audio transcode";
  if (id === "full" || id === "cinema") return ["remux", "audio transcode", "video transcode", "subtitle burn"].includes(mode);
  return false;
}

function policyDecisionLabel(policy = {}, decision = {}) {
  if (policyAllowsDecision(policy, decision)) return "Allowed";
  return "Fallback needed";
}

function policyImpactNote(policy = {}, decision = {}) {
  if (policyAllowsDecision(policy, decision)) return "";
  const mode = playbackReadinessLabel(decision);
  return `<div class="policy-impact">
    <strong>${escapeHTML(policy.label || "Original Only")} blocks automatic changes</strong>
    <span>${escapeHTML(`This file may still play on another device, but this player needs ${mode.toLowerCase()}. Vyrden will not convert or repackage it unless you choose a fallback.`)}</span>
    <div>${fallbackOptions(decision).map(item => `<small>${escapeHTML(item)}</small>`).join("")}</div>
  </div>`;
}

function fallbackOptions(decision = {}) {
  const mode = String(decision.mode || "").toLowerCase();
  const options = ["Try a device that supports the original file"];
  if (mode === "remux") options.push("Allow live repackage with no quality loss");
  if (mode === "audio transcode") options.push("Allow temporary audio conversion");
  if (mode === "video transcode" || mode === "subtitle burn") options.push("Allow temporary video conversion");
  options.push("Create optimized version only if you explicitly want a stored file");
  return options;
}

function needsVideoHardware(decision = {}) {
  const mode = String(decision.mode || "").toLowerCase();
  return mode === "video transcode" || mode === "subtitle burn";
}

function hardwareImpactNote(decision = {}, hardware = {}) {
  if (!needsVideoHardware(decision)) return "";
  const unlocked = hardware.available && hardware.unlockState === "unlocked";
  const title = unlocked ? "GPU conversion ready" : hardware.available ? "GPU conversion locked" : "GPU conversion would help";
  const detail = unlocked
    ? "This file can still play. Vyrden can move heavy video work to the GPU for lower CPU load, less heat, and more headroom."
    : hardware.available
    ? "This file can still play. Hardware support is present, but GPU conversion is a paid feature, so this route stays on CPU until unlocked."
    : "This file can still play, but video conversion would run on the CPU. Expect higher CPU use, more heat, and less headroom for other streams.";
  return `<div class="hardware-impact">
    <strong>${escapeHTML(title)}</strong>
    <span>${escapeHTML(detail)}</span>
  </div>`;
}

function deviceForecast(decisions = [], hardware = {}) {
  if (!Array.isArray(decisions) || !decisions.length) return "";
  return `<div class="device-forecast">
    <div class="device-forecast-title">Player impact</div>
    <div class="device-forecast-grid">
      ${decisions.map(decision => `<div>
        <span>${escapeHTML(profileName(decision.clientProfile))}</span>
        <strong>${escapeHTML(playbackReadinessLabel(decision))}</strong>
        <small>${escapeHTML(playbackEffortLabel(decision, hardware))}</small>
      </div>`).join("")}
    </div>
  </div>`;
}

function profileName(value = "") {
  const labels = {
    web: "Web",
    "android-tv": "Android TV",
    "apple-tv": "Apple TV",
    chromecast: "Chromecast",
    ios: "iPhone / iPad",
  };
  return labels[String(value || "").toLowerCase()] || value || "Player";
}

function movieDetailRail(movie = {}, selected = null) {
  return `
    <aside class="panel pad source-control-panel">
      ${sourceControlHeader(selected)}
      ${sourceControlActions(selected)}
    </aside>`;
}

function downloadPlan(label, value, note, mediaSourceId, profile) {
  return `<button class="download-plan-item" onclick="startDownload('${mediaSourceId}','${profile}')">
    <strong>${escapeHTML(label)}</strong>
    <span>${escapeHTML(value)}</span>
    <small>${escapeHTML(note)}</small>
  </button>`;
}

function factRow(label, value) {
  return `<div class="fact-row"><span>${escapeHTML(label)}</span><strong>${escapeHTML(value || "Pending")}</strong></div>`;
}

function sourceControlHeader(selected = null) {
  if (!selected) {
    return `<div class="source-control-hero">
      <span>Ready to play?</span>
      <strong>No source selected</strong>
      <p>Scan this library to attach a playable file.</p>
    </div>`;
  }
  const decision = selected.decision || {};
  return `<div class="source-control-hero">
    <span>Ready to play?</span>
    <strong>${escapeHTML(playbackReadinessLabel(decision))}</strong>
    <p>${escapeHTML(playbackReason(decision))}</p>
  </div>`;
}

function sourceMetric(label, value) {
  return `<div><span>${escapeHTML(label)}</span><strong>${escapeHTML(value || "Pending")}</strong></div>`;
}

function sourceControlActions(selected = null) {
  if (!selected) return "";
  return `<div class="source-control-actions">
    <button class="primary" onclick="probeSource('${selected.mediaSourceId}')">${selected.source?.probed ? "Recheck file" : "Check file now"}</button>
    <button onclick="openSourceInspector('${selected.mediaSourceId}')">File details</button>
  </div>`;
}

function compactFact(label, value, long = false) {
  return `<div class="compact-fact ${long ? "long" : ""}"><span>${escapeHTML(label)}</span><strong>${escapeHTML(value || "Pending")}</strong></div>`;
}

async function openSourceInspector(mediaSourceId) {
  const [source, tracks, decision, playbackState] = await Promise.all([
    api(`/api/media-sources/${mediaSourceId}`).catch(() => ({})),
    api(`/api/media-sources/${mediaSourceId}/tracks`).catch(() => ({ audioTracks: [], subtitleTracks: [] })),
    api(`/api/playback/decision?mediaSourceId=${mediaSourceId}&clientProfile=web`).catch(() => ({})),
    api(`/api/playback/state/${mediaSourceId}`).catch(() => ({})),
  ]);
  closeSourceInspector();
  const overlay = document.createElement("div");
  overlay.className = "source-drawer-backdrop";
  overlay.id = "sourceInspector";
  overlay.innerHTML = `<aside class="source-drawer" role="dialog" aria-modal="true" aria-label="Source inspector">
    <div class="drawer-head">
      <div>
        <span>Source intelligence</span>
        <strong>${escapeHTML(source.name || source.relPath || mediaSourceId)}</strong>
      </div>
      <button type="button" onclick="closeSourceInspector()">Close</button>
    </div>
    <div class="decision"><strong>${escapeHTML(playbackReadinessLabel(decision))}</strong><span>${escapeHTML(playbackReason(decision))}</span></div>
    <div class="source-inspector-grid">
      ${inspectorFact("Probe", source.probed ? "Complete" : "Needed", source.probed ? "Media facts are cached locally." : "Run ffprobe for this one source now.")}
      ${inspectorFact("Container", source.container || "Pending", sourceCodecLine(source))}
      ${inspectorFact("Video", sourceVideoLine(source, decision), source.bitrate ? formatBitrate(source.bitrate) : "Bitrate pending")}
      ${inspectorFact("Tracks", `${tracks.audioTracks?.length || 0} audio / ${tracks.subtitleTracks?.length || 0} subtitles`, source.subtitleStreams ? `${source.subtitleStreams} embedded subtitle streams` : "No subtitle facts yet")}
      ${inspectorFact("Progress", playbackState.watched ? "Watched" : `${Math.round((playbackState.percent || 0) * 100)}%`, playbackState.progressSeconds ? formatDuration(playbackState.progressSeconds) : "Not started")}
      ${inspectorFact("Server work", playbackEffortLabel(decision), loadLabel(decision))}
    </div>
    <div class="drawer-section">
      <div class="panel-title"><strong>Audio Tracks</strong></div>
      <div class="mini-list">${trackSummaryRows(tracks.audioTracks, "No audio tracks cached. Probe this source.")}</div>
    </div>
    <div class="drawer-section">
      <div class="panel-title"><strong>Subtitle Tracks</strong></div>
      <div class="mini-list">${trackSummaryRows(tracks.subtitleTracks, "No subtitle tracks cached or available.")}</div>
    </div>
    <div class="drawer-section">
      <div class="panel-title"><strong>Actions</strong></div>
      <div class="inline-actions">
        <button class="primary" onclick="probeSource('${mediaSourceId}')">${source.probed ? "Re-probe" : "Probe Now"}</button>
        <button onclick="startWork('${mediaSourceId}','remux')">Repackage while playing</button>
        <button onclick="startWork('${mediaSourceId}','transcode')">Convert while playing</button>
        <button onclick="startDownload('${mediaSourceId}','balanced')">Create remote optimized version</button>
        <a class="button" href="/api/playback/decision?mediaSourceId=${mediaSourceId}&clientProfile=web" target="_blank">Raw Decision</a>
      </div>
    </div>
  </aside>`;
  overlay.addEventListener("click", event => {
    if (event.target === overlay) closeSourceInspector();
  });
  document.body.appendChild(overlay);
}

function closeSourceInspector() {
  document.getElementById("sourceInspector")?.remove();
}

function inspectorFact(label, value, note) {
  return `<div class="inspector-fact"><span>${escapeHTML(label)}</span><strong>${escapeHTML(value || "Pending")}</strong><small>${escapeHTML(note || "")}</small></div>`;
}

function trackSummaryRows(items = [], emptyText = "No tracks cached.") {
  if (!items || !items.length) return `<div><strong>${escapeHTML(emptyText)}</strong><span>Source probe required</span></div>`;
  return items.map(item => `<div><strong>${escapeHTML(trackLanguage(item))} - ${escapeHTML(String(item.codec || "track").toUpperCase())}</strong><span>${escapeHTML(trackDescriptor(item))}</span></div>`).join("");
}

function trackDescriptor(item = {}) {
  const details = [];
  if (item.channels) details.push(`${item.channels} channels`);
  if (item.default) details.push("default");
  if (item.forced) details.push("forced");
  if (item.title) details.push(item.title);
  return details.join(" - ") || "Selectable track";
}

function sourceStrip(model = {}) {
  const source = model.source || {};
  return `<div class="source-strip">
    <div><span>Selected source</span><strong>${escapeHTML(model.qualityLabel || sourceQualityLabel(source) || "Original source")}</strong></div>
    <div><span>Playback readiness</span><strong>${escapeHTML(playbackReadinessLabel(model.decision))}</strong></div>
    <div><span>Bitrate / size</span><strong>${escapeHTML(source.bitrate ? formatBitrate(source.bitrate) : "Bitrate pending")} - ${formatBytes(model.sizeBytes)}</strong></div>
  </div>`;
}

function playbackForecastMarkup(model) {
  if (!model) {
    return `<div class="panel-title"><strong>Playback Readiness</strong><span class="badge warn">No Source</span></div>
      <div class="decision"><strong id="forecastMode">No Source</strong><span id="forecastReason">Scan or attach a source before playback.</span></div>`;
  }
  const source = model.source || {};
  const decision = model.decision || {};
  return `<div class="panel-title"><strong>Playback Readiness</strong><span class="badge route" id="forecastLoad">${escapeHTML(loadLabel(decision))}</span></div>
    <div class="decision"><strong id="forecastMode">${escapeHTML(playbackReadinessLabel(decision))}</strong><span id="forecastReason">${escapeHTML(playbackReason(decision))}</span></div>
    <div class="readiness-list">
      ${readinessItem("File check", playbackActionLabel(decision.containerAction || "pending"), "forecastReasonShort")}
      ${readinessItem("Video", sourceVideoLine(source, decision), "forecastVideo")}
      ${readinessItem("Audio", playbackActionLabel(decision.audioAction || "pending"), "forecastAudio")}
      ${readinessItem("Subtitles", playbackActionLabel(decision.subtitleAction || "none"), "forecastSubtitles")}
      ${readinessItem("Network", decision.estimatedNetworkBitrate ? formatBitrate(decision.estimatedNetworkBitrate) : "Not known yet")}
      ${readinessItem("Server", playbackEffortLabel(decision), "forecastServer")}
    </div>
    ${decision.suggestedFixes?.length ? `<div class="mini-list">${decision.suggestedFixes.map(item => `<div><strong>${escapeHTML(friendlyFix(item))}</strong><span>Recommended next step</span></div>`).join("")}</div>` : ""}`;
}

function readinessItem(label, value, id = "") {
  return `<div class="readiness-item"><span>${escapeHTML(label)}</span><strong ${id ? `id="${id}"` : ""}>${escapeHTML(value || "Pending")}</strong></div>`;
}

function sourceQualityLabel(source = {}) {
  const height = Number(source.height || 0);
  const resolution = height >= 2000 ? "4K" : height >= 1000 ? "1080p" : height ? `${height}p` : "";
  return [resolution, source.videoCodec ? String(source.videoCodec).toUpperCase() : "", source.container ? String(source.container).toUpperCase() : ""].filter(Boolean).join(" ");
}

function sourceCodecLine(source = {}) {
  if (!source.probed) return "File check needed";
  return [source.videoCodec ? String(source.videoCodec).toUpperCase() : "video", source.container ? String(source.container).toUpperCase() : "container", source.audioStreams ? `${source.audioStreams} audio` : "", source.subtitleStreams ? `${source.subtitleStreams} subs` : ""].filter(Boolean).join(" - ");
}

function sourceVideoLine(source = {}, decision = {}) {
  return [source.videoCodec ? String(source.videoCodec).toUpperCase() : "Video pending", source.width && source.height ? `${source.width}x${source.height}` : "", playbackActionLabel(decision.videoAction || "")].filter(Boolean).join(" - ");
}

function loadLabel(decision = {}) {
  if (String(decision.mode || "").toLowerCase() === "decision deferred") return "Check needed";
  if (decision.estimatedCpuCost === "high") return "Heavy server work";
  if (decision.estimatedCpuCost === "medium") return "Some server work";
  if (decision.estimatedCpuCost === "low") return "Easy on server";
  if (decision.estimatedCpuCost === "none") return "No server work";
  return "Checking";
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
  if (!items.length) return trackRow("File check needed", "Audio details not available yet", true);
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
  const off = trackRow("Subtitles off", "Simplest playback path", !selected, `onclick="selectPlaybackTrack('subtitle','${mediaSourceId}',null)"`);
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
  setText("forecastMode", playbackReadinessLabel(decision));
  setText("forecastReason", playbackReason(decision));
  setText("forecastReasonShort", playbackActionLabel(decision.containerAction || "selected source"));
  setText("forecastAudio", playbackActionLabel(decision.audioAction || "pending"));
  setText("forecastSubtitles", playbackActionLabel(decision.subtitleAction || "none"));
  setText("forecastServer", playbackEffortLabel(decision));
  setText("forecastLoad", loadLabel(decision));
}

function setText(id, value) {
  const target = document.getElementById(id);
  if (target) target.textContent = value ?? "";
}

function serverImpact(decision = {}) {
  if (decision.mode === "Subtitle Burn" || decision.mode === "Video Transcode") return "Video conversion needed";
  if (decision.mode === "Audio Transcode") return "Audio conversion needed";
  if (decision.mode === "Remux") return "Live repackage";
  if (decision.mode === "Direct Play") return "Low impact route";
  return "Decision pending";
}

function playbackReadinessLabel(decision = {}) {
  const mode = String(decision.mode || "").toLowerCase();
  if (mode === "direct play") return "Ready to play";
  if (mode === "remux") return "Repackage while playing";
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
  if (mode === "audio transcode") return "Light PC load";
  if (mode === "video transcode" || mode === "subtitle burn") {
    if (hardware.unlockState === "unlocked" && hardware.configured) return "GPU conversion";
    if (hardware.available) return "GPU locked";
    return "High CPU load";
  }
  if (mode === "decision deferred") return "Waiting for file check";
  return serverImpact(decision);
}

function playbackReason(decision = {}, hardware = {}) {
  const reason = String(decision.reason || "");
  const mode = String(decision.mode || "").toLowerCase();
  if (!reason) return "Vyrden is checking this file before choosing the best playback path.";
  if (reason.includes("has not been probed")) return "Vyrden needs to check this file once before it can confirm the best playback path.";
  if (reason.includes("Container and video codec match")) return "This player can stream the file directly. Vyrden should not need extra CPU, GPU, or temporary disk work.";
  if (reason.includes("Video can direct play")) return "The video can play as-is, but Vyrden may convert the audio track for this player. Expect a light CPU load.";
  if (reason.includes("video codec is compatible")) return "The video can stay untouched, but Vyrden may need to repackage the file while playing for this device. No permanent file is created.";
  if (reason.includes("subtitle track is image-based")) return "This file can still play, but image subtitles may need to be burned into the video. That is a heavy path and can use significant CPU/GPU.";
  if (reason.includes("not safely direct-playable")) {
    if (mode === "video transcode") return hardware.configured && hardware.unlockState === "unlocked"
      ? "This file can still play, but this player profile needs video conversion. GPU acceleration is unlocked and is the right path for keeping CPU load low."
      : "This file can still play, but this player profile needs video conversion. Without hardware acceleration, expect high CPU use, more power draw, and more heat.";
    if (mode === "subtitle burn") return "This file can still play, but subtitles may need to be burned into the video for this player. This is one of the heaviest playback paths.";
    if (mode === "audio transcode") return "The video can stay intact, but Vyrden may convert audio for this player. This is usually a light PC load.";
    if (mode === "remux") return "The streams can stay intact, but Vyrden may repackage the file while playing for this device. This is temporary and usually low impact.";
    return "This file can still play, but Vyrden may need to prepare it before or during playback for this player profile.";
  }
  return reason;
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
    transcode: "Convert",
    burn_in: "Burn subtitles",
    selected_source: "Selected source",
  };
  return labels[String(value).toLowerCase()] || value;
}

function friendlyFix(value = "") {
  if (String(value).toLowerCase().includes("ffprobe")) return "Run file check for this file";
  return value;
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
        <button onclick="refreshMetadataBatch('series')">Refresh Metadata</button>
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
  state.currentSeries = series;
  const seasonSections = await Promise.all(series.seasons.map(async season => `<section class="card"><h2>Season ${season.seasonNumber}</h2>${await episodeList(series.id, season.episodes)}</section>`));
  viewTitle.textContent = series.title;
  view.innerHTML = `
    <div class="detail-shell">
      <section class="detail-hero">
        <img alt="" src="${artworkURL("series", series.id, "backdrop")}">
        <div class="detail-poster"><img alt="" src="${artworkURL("series", series.id)}"></div>
        <div class="detail-copy">
          <button class="ghost" onclick="navigate('tv')">Back to TV</button>
          <h2>${escapeHTML(series.title)}</h2>
          <p>${series.metadata?.overview ? escapeHTML(series.metadata.overview) : `${series.seasonCount} seasons - ${series.episodeCount} episodes`}</p>
          <div class="meta-line">
            <span class="badge">${series.seasonCount} seasons</span>
            <span class="badge">${series.episodeCount} episodes</span>
            ${metadataBadges(series.metadata)}
          </div>
          ${ratingsRail(series.metadata)}
          <div class="actions">
            <button class="primary" onclick="refreshMetadata('series','${series.id}','${escapeAttr(series.title)}',0)">Fetch Ratings</button>
            <button onclick="refreshMetadataBatch('series')">Refresh TV Metadata</button>
          </div>
        </div>
      </section>
      <div class="stack">
        ${seasonSections.join("")}
      </div>
    </div>`;
}

async function episodeList(seriesId, episodes) {
  if (!episodes.length) return empty("No episodes in this season.");
  const rows = await Promise.all(episodes.map(async episode => {
    const version = episode.versions && episode.versions[0];
    const state = version ? await api(`/api/playback/state/${version.mediaSourceId}`).catch(() => ({})) : {};
    const playLabel = state.progressSeconds > 5 && !state.watched ? "Resume" : "Play";
    const watchedLabel = state.watched ? "Watched" : state.progressSeconds > 5 ? `${Math.round((state.percent || 0) * 100)}%` : `${episode.versionCount || 0} version${episode.versionCount === 1 ? "" : "s"}`;
    const play = version ? `<a class="button primary" href="/play/${version.mediaSourceId}" target="_blank">${playLabel}</a><a class="button" href="/play/${version.mediaSourceId}?start=0" target="_blank">Start Over</a><button onclick="markWatched('${version.mediaSourceId}', true)">Watched</button><button onclick="markWatched('${version.mediaSourceId}', false)">Unwatched</button>` : `<span class="muted">No source</span>`;
    const label = episode.episodeEnd && episode.episodeEnd !== episode.episodeNumber ? `E${episode.episodeNumber}-E${episode.episodeEnd}` : `E${episode.episodeNumber}`;
    return `<div class="episode-row">
      <button class="episode-jump" onclick="showEpisode('${seriesId}','${episode.id}')">${label}</button>
      <button class="episode-title-button" onclick="showEpisode('${seriesId}','${episode.id}')">${escapeHTML(episode.title || "Episode")}</button>
      <small>${escapeHTML(watchedLabel)}</small>
      <div class="inline-actions">${play}${episode.needsReview ? `<button onclick="openMetadataFix('episode','${episode.id}','${escapeAttr(episode.title || "Episode")}',0)">Fix</button>` : ""}</div>
    </div>`;
  }));
  return rows.join("");
}

async function showEpisode(seriesId, episodeId) {
  const series = state.currentSeries?.id === seriesId ? state.currentSeries : await api(`/api/series/${seriesId}`);
  state.currentSeries = series;
  const season = (series.seasons || []).find(item => (item.episodes || []).some(episode => episode.id === episodeId));
  const episode = (season?.episodes || []).find(item => item.id === episodeId);
  if (!episode) {
    view.innerHTML = `<div class="card error">Episode not found.</div>`;
    return;
  }
  const versionModels = await Promise.all((episode.versions || []).map(loadVersionModel));
  const selected = versionModels[0];
  const tracks = selected?.tracks || { audioTracks: [], subtitleTracks: [] };
  if (selected) seedPlaybackSelection(selected.mediaSourceId, tracks);
  const episodeCode = episode.episodeEnd && episode.episodeEnd !== episode.episodeNumber ? `S${season.seasonNumber} E${episode.episodeNumber}-E${episode.episodeEnd}` : `S${season.seasonNumber} E${episode.episodeNumber}`;
  viewTitle.textContent = `${series.title} - ${episodeCode}`;
  view.innerHTML = `
    <div class="detail-command">
      <div class="detail-backdrop"><img alt="" src="${artworkURL("series", series.id, "backdrop")}"></div>
      <section class="detail-main">
        <button class="back-button" onclick="showSeries('${series.id}')">Back to ${escapeHTML(series.title)}</button>
        <div class="eyebrow">Episode detail</div>
        <h1>${escapeHTML(episode.title || episodeCode)}</h1>
        <div class="meta-line">
          <span class="badge">${escapeHTML(series.title)}</span>
          <span class="badge">${escapeHTML(episodeCode)}</span>
          <span class="badge">${episode.versionCount || 0} version${episode.versionCount === 1 ? "" : "s"}</span>
          <span class="badge ${episode.needsReview ? "warn" : "good"}">${episode.needsReview ? "Needs Review" : "Matched"}</span>
        </div>
        <p class="lead">${escapeHTML(series.metadata?.overview || "Local episode source with playback readiness, audio, subtitle, and metadata details in one place.")}</p>
        ${ratingsRail(series.metadata)}
        <div class="actions detail-hero-actions">
          ${selected ? `<a class="button primary" href="/play/${selected.mediaSourceId}" target="_blank">Play</a><a class="button" href="/play/${selected.mediaSourceId}?start=0" target="_blank">Start Over</a>` : ""}
          ${selected ? `<button onclick="markWatched('${selected.mediaSourceId}', true)">Watched</button><button onclick="markWatched('${selected.mediaSourceId}', false)">Unwatched</button>` : ""}
          <button onclick="refreshMetadata('series','${series.id}','${escapeAttr(series.title)}',0)">Fetch Metadata</button>
          ${episode.needsReview ? `<button onclick="openMetadataFix('episode','${episode.id}','${escapeAttr(episode.title || "Episode")}',0)">Fix Episode</button>` : ""}
        </div>
        <section class="section">
          <div class="section-head"><div class="section-title">Source file</div></div>
          <div class="detail-source-workbench">
            <div class="version-grid">${versionModels.map((model, index) => versionCard(model, index === 0)).join("") || empty("No playable versions found.")}</div>
            ${episodeDetailRail(series, episode, selected)}
          </div>
        </section>
        <section class="section">
          <div class="section-head"><div class="section-title">Audio & Subtitles</div></div>
          <div class="track-grid">
            <div class="track-panel"><strong>Audio</strong>${audioTrackRows(tracks.audioTracks, selected?.mediaSourceId)}</div>
            <div class="track-panel"><strong>Subtitles</strong>${subtitleTrackRows(tracks.subtitleTracks, selected?.mediaSourceId)}</div>
          </div>
        </section>
      </section>
    </div>`;
  if (selected) updatePlaybackForecast(selected.mediaSourceId);
}

function episodeDetailRail(series = {}, episode = {}, selected = null) {
  return `
    <aside class="panel pad source-control-panel">
      ${sourceControlHeader(selected)}
      ${sourceControlActions(selected)}
    </aside>`;
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
        <div class="signal-stack" data-live="activity-signals">
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
        <div class="card"><h2>Active Sessions</h2><div data-live="activity-sessions">${sessionCards(sessions.sessions)}</div></div>
        <div class="card"><h2>Work</h2><div data-live="activity-work">${jobCards(work.work)}</div></div>
      </div>
      <div class="grid two">
        <div class="card"><h2>Scans</h2><div data-live="activity-scans">${jobCards(scans.scans)}</div></div>
        <div class="card"><h2>Probes</h2><div data-live="activity-probes">${jobCards(probes.probes)}</div></div>
      </div>
      <div class="card"><h2>Downloads</h2><div data-live="activity-downloads">${downloadCards(downloads.downloads)}</div></div>
    </div>`;
}

async function renderHealth() {
  const [health, review, versions] = await Promise.all([api("/api/catalog/health"), api("/api/review"), api("/api/versions")]);
  const workbench = await reviewWorkbench(review.items || []);
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
      <div class="toolbar">
        <button onclick="refreshMetadataBatch('movie')">Refresh Movie Metadata</button>
        <button onclick="refreshMetadataBatch('series')">Refresh TV Metadata</button>
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
        <div class="card"><h2>Metadata Workbench</h2>${workbench}</div>
        <div class="card"><h2>Version Groups</h2>${versionGroups(versions.versions)}</div>
      </div>
    </div>`;
}

async function renderPlaybackLab() {
  const [payload, profiles, downloads] = await Promise.all([api("/api/media-sources?limit=200"), api("/api/devices/profiles"), api("/api/downloads")]);
  view.innerHTML = `<div class="stack">
    <div class="hero-strip">
      <div class="hero-copy"><span>Source Inspector</span><strong>Advanced diagnostics now live on each movie and episode source.</strong><p>This screen remains as an internal safety net, but normal users should inspect playback routes from the version cards inside Movies and TV Shows.</p></div>
      <div class="signal-stack">${signalPill("Sources", payload.mediaSources.length)}${signalPill("Profiles", profiles.profiles.length)}${signalPill("Downloads", downloads.downloads.length)}</div>
    </div>
    <div class="card"><h2>Client Profiles</h2>${profileCards(profiles.profiles)}</div>
    <div class="card"><h2>Source Inspector</h2>${playbackCards(payload.mediaSources)}</div>
    <div class="card"><h2>Offline Downloads</h2><div data-live="playback-downloads">${downloadCards(downloads.downloads)}</div></div>
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
    <div class="card"><h2>Playback Policy</h2>${playbackPolicyForm(settings.config.playbackPolicy || "original_only")}</div>
    <div class="card"><h2>Library Automation</h2>${automationSettingsForm(settings.config)}</div>
    <div class="card"><h2>Hardware Acceleration</h2>${hardwareAccelerationPanel(performance.hardwareAcceleration || {}, settings.config)}</div>
    <div class="card"><h2>Metadata Providers</h2>${providerSettingsForm(settings.config.metadataProviders || {})}</div>
    <div class="card"><h2>Folder Capacity</h2>${runtimeFolders(system.disks || [])}</div>
    <div class="card"><h2>Server Config</h2>${settingsGrid(settings.config)}</div>
    <div class="card"><h2>Performance</h2><pre>${escapeHTML(JSON.stringify(performance, null, 2))}</pre></div>
  </div>`;
}

function playbackPolicyForm(current = "original_only") {
  const policies = [
    ["original_only", "Original Only", "Plays the original file only. If this device needs help, Vyrden shows fallback options instead of changing anything automatically."],
    ["light", "Light Compatibility", "Allows live repackage and audio conversion. Video stays untouched, so there is no video quality loss."],
    ["full", "Full Compatibility", "Allows temporary video conversion when a player needs it. No permanent optimized version is created."],
    ["cinema", "Cinema Server", "Allows heavier compatibility work and future scheduled optimization controls for power users."],
  ];
  return `<form class="policy-form" onsubmit="savePlaybackPolicy(event)">
    <div class="policy-grid">
      ${policies.map(([value, label, detail]) => `<label class="policy-option ${current === value ? "selected" : ""}">
        <input type="radio" name="playbackPolicy" value="${escapeAttr(value)}" ${current === value ? "checked" : ""}>
        <strong>${escapeHTML(label)}</strong>
        <span>${escapeHTML(detail)}</span>
      </label>`).join("")}
    </div>
    <div class="inline-actions"><button class="primary" type="submit">Save playback policy</button><span class="muted">This controls automatic live playback work only. Stored optimized versions remain explicit user actions.</span></div>
  </form>`;
}

function hardwareAccelerationPanel(hardware = {}, config = {}) {
  const encoders = Array.isArray(hardware.encoders) ? hardware.encoders : [];
  const available = !!hardware.available;
  const status = available ? "FFmpeg support found" : hardware.status === "unknown" ? "Check failed" : "Not detected";
  const unlock = hardware.unlockState === "unlocked" ? "Unlocked" : "Paid feature locked";
  const workers = Number(config.gpuWorkers ?? hardware.gpuWorkers ?? 0);
  const encoderList = encoders.length
    ? `<div class="encoder-grid">${encoders.map(encoder => `<div class="encoder-card">
        <span>${escapeHTML(encoder.vendor || "Hardware")}</span>
        <strong>${escapeHTML(encoder.codec || encoder.label || "Encoder")}</strong>
        <small>${escapeHTML(encoder.id || "")}</small>
      </div>`).join("")}</div>`
    : `<div class="hardware-empty">No hardware encoder support was detected from FFmpeg yet. Direct play still works normally, but heavy video conversion will use the CPU until GPU support is available.</div>`;
  return `<div class="hardware-accel-panel">
    <div class="hardware-accel-copy">
      <div><span>${escapeHTML(status)}</span><strong>GPU-assisted conversion is the premium playback path.</strong></div>
      <p>When a player needs video conversion or subtitle burn-in, hardware acceleration can move most of the work from CPU to Intel Quick Sync, NVIDIA NVENC, AMD AMF, VAAPI, or VideoToolbox. Vyrden detects FFmpeg support here; actual runtime use will be verified before a paid hardware path starts.</p>
    </div>
    <div class="settings-grid hardware-accel-facts">
      <div><span>Unlock state</span><strong>${escapeHTML(unlock)}</strong></div>
      <div><span>GPU worker slots</span><strong>${escapeHTML(workers)}</strong></div>
      <div><span>Supported encoders</span><strong>${escapeHTML(encoders.length)}</strong></div>
    </div>
    ${encoderList}
    <div class="hardware-note">${escapeHTML(hardware.recommendation || "Hardware acceleration will be used only for routes that need video conversion.")}</div>
    ${hardware.error ? `<div class="hardware-empty warn">${escapeHTML(hardware.error)}</div>` : ""}
    <div id="hardwareTestResult" class="hardware-test-result"></div>
    <form class="path-form hardware-form" onsubmit="saveHardwareSettings(event)">
      <label><span>GPU worker slots</span><input name="gpuWorkers" value="${escapeAttr(workers)}" inputmode="numeric" autocomplete="off"></label>
      <div class="inline-actions">
        <button class="primary" type="submit">Save hardware settings</button>
        <button type="button" onclick="testHardwareAcceleration(event)">Test hardware acceleration</button>
        <span class="muted">This reserves GPU conversion capacity. Direct play and normal library browsing are unaffected.</span>
      </div>
    </form>
  </div>`;
}

function automationSettingsForm(config = {}) {
  const mode = config.librarySyncMode || "daily";
  const interval = Number(config.syncIntervalMins || 1440);
  const probeLimit = Number(config.probeBatchLimit || 50);
  return `<form class="path-form automation-form" onsubmit="saveAutomationSettings(event)">
    <label><span>Library sync</span><select name="librarySyncMode">
      ${option("manual", "Manual only", mode)}
      ${option("hourly", "Hourly", mode)}
      ${option("daily", "Daily", mode)}
      ${option("custom", "Custom interval", mode)}
    </select></label>
    <label><span>Custom interval</span><input name="syncIntervalMins" value="${escapeAttr(interval)}" inputmode="numeric"></label>
    <label><span>Probe batch size</span><input name="probeBatchLimit" value="${escapeAttr(probeLimit)}" inputmode="numeric"></label>
    <div class="inline-actions"><button class="primary" type="submit">Save automation</button><span class="muted">Scans stay quick. Media checks run in batches so NAS, USB, and local disks are not hammered.</span></div>
  </form>`;
}

function option(value, label, selected) {
  return `<option value="${escapeAttr(value)}" ${value === selected ? "selected" : ""}>${escapeHTML(label)}</option>`;
}

function runtimePathForm(paths = {}) {
  const fields = [
    ["data", "Database and settings"],
    ["transcode", "Transcode temp"],
    ["downloads", "Optimized versions"],
    ["metadata", "Metadata and artwork"],
    ["cache", "Cache"],
    ["temp", "Scratch temp"],
  ];
  return `<form class="path-form" onsubmit="saveRuntimePaths(event)">
    ${fields.map(([key, label]) => `<label><span>${escapeHTML(label)}</span><input name="${key}" value="${escapeAttr(paths[key] || "")}" autocomplete="off"></label>`).join("")}
    <div class="inline-actions"><button class="primary" type="submit">Save folders</button><span class="muted">Applies fully after restart so active jobs do not lose files.</span></div>
  </form>`;
}

function providerSettingsForm(providers = {}) {
  return `<form class="path-form" onsubmit="saveProviderSettings(event)">
    <div class="settings-grid provider-grid">
      ${providerCard("OMDb", "omdbApiKey", providers.omdb, "IMDb, Rotten Tomatoes, Metacritic")}
      ${providerCard("TMDB", "tmdbApiKey", providers.tmdb, "TMDB community ratings and IDs")}
    </div>
    <div class="inline-actions"><button class="primary" type="submit">Save providers</button><span class="muted">Keys are stored locally in Vyrden config and used only by your server.</span></div>
  </form>`;
}

function providerCard(name, inputName, provider = {}, description = "") {
  const configured = provider.configured ? "Configured" : "Not configured";
  return `<label class="provider-card">
    <span>${escapeHTML(name)} <em>${escapeHTML(configured)}</em></span>
    <small>${escapeHTML(description)}</small>
    <input type="password" name="${inputName}" value="" placeholder="${provider.configured ? "Leave blank to keep existing key" : "Paste API key"}" autocomplete="off">
  </label>`;
}

function settingsGrid(config) {
  return `<div class="settings-grid">${Object.entries(config).filter(([key]) => key !== "metadataProviders").map(([key, value]) => `<div><span>${escapeHTML(key)}</span><strong>${escapeHTML(value)}</strong></div>`).join("")}</div>`;
}

async function saveProviderSettings(event) {
  event.preventDefault();
  const data = new FormData(event.currentTarget);
  const payload = {
    omdbApiKey: data.get("omdbApiKey"),
    tmdbApiKey: data.get("tmdbApiKey"),
  };
  const result = await send("/api/settings", payload, "PUT");
  alert(result.restartRequired ? "Saved. Restart Vyrden for provider services to reload the new keys." : "Saved.");
  navigate("settings");
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

async function saveAutomationSettings(event) {
  event.preventDefault();
  const data = new FormData(event.currentTarget);
  const mode = String(data.get("librarySyncMode") || "daily");
  const intervalByMode = { manual: 1440, hourly: 60, daily: 1440 };
  const interval = mode === "custom" ? Number(data.get("syncIntervalMins") || 1440) : intervalByMode[mode] || 1440;
  const probeLimit = Number(data.get("probeBatchLimit") || 50);
  const result = await send("/api/settings", {
    librarySyncMode: mode,
    syncIntervalMins: Math.max(15, interval),
    probeBatchLimit: Math.max(1, Math.min(500, probeLimit)),
  }, "PUT");
  alert(result.restartRequired ? "Saved. Restart Vyrden for the scheduler to use the new cadence." : "Saved.");
  navigate("settings");
}

async function savePlaybackPolicy(event) {
  event.preventDefault();
  const data = new FormData(event.currentTarget);
  const result = await send("/api/settings", { playbackPolicy: data.get("playbackPolicy") || "original_only" }, "PUT");
  state.performanceSnapshot = null;
  alert(result.restartRequired ? "Saved. Restart Vyrden if an active player is already running." : "Saved.");
  navigate("settings");
}

async function saveHardwareSettings(event) {
  event.preventDefault();
  const data = new FormData(event.currentTarget);
  const gpuWorkers = Math.max(0, Math.min(8, Number(data.get("gpuWorkers") || 0)));
  const result = await send("/api/settings", { gpuWorkers }, "PUT");
  alert(result.restartRequired ? "Saved. Restart Vyrden for hardware worker limits to take effect." : "Saved.");
  navigate("settings");
}

async function testHardwareAcceleration(event) {
  const button = event?.currentTarget;
  const target = document.getElementById("hardwareTestResult");
  const original = button?.textContent;
  if (button) {
    button.disabled = true;
    button.textContent = "Testing...";
  }
  if (target) target.innerHTML = `<div class="hardware-empty">Running a short FFmpeg encode test for each supported hardware path. This may take a few seconds.</div>`;
  try {
    const result = await send("/api/settings/hardware/test", {});
    const tests = Array.isArray(result.tests) ? result.tests : [];
    if (target) {
      target.innerHTML = `<div class="hardware-test-summary">
        <strong>${escapeHTML(result.working || 0)} of ${escapeHTML(result.tested || tests.length)} hardware paths passed runtime test</strong>
        <span>${result.working ? "At least one GPU path can encode on this machine. Paid hardware conversion can use a tested route when unlocked." : "FFmpeg lists hardware support, but no tested GPU encode path succeeded yet. Vyrden should keep using CPU conversion until drivers or FFmpeg are fixed."}</span>
      </div>
      <div class="encoder-grid">${tests.map(test => `<div class="encoder-card ${test.ok ? "ok" : "fail"}">
        <span>${escapeHTML(test.vendor || "Hardware")}</span>
        <strong>${escapeHTML(test.codec || test.id || "Encoder")}</strong>
        <small>${test.ok ? `Passed in ${escapeHTML(test.durationMs || 0)} ms` : escapeHTML(test.error || "Failed")}</small>
      </div>`).join("")}</div>`;
    }
  } catch (error) {
    if (target) target.innerHTML = `<div class="hardware-empty warn">${escapeHTML(error.message)}</div>`;
  } finally {
    if (button) {
      button.disabled = false;
      button.textContent = original;
    }
  }
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
  navigate("activity");
}

async function startScan(path) {
  await send(path, { sampleLimit: 50 });
  navigate("activity");
}

async function startProbe() {
  await send("/api/probes", { limit: 50 });
  navigate("activity");
}

async function probeSource(mediaSourceId) {
  const button = event?.currentTarget;
  const original = button?.textContent;
  if (button) {
    button.disabled = true;
    button.textContent = "Probing...";
  }
  try {
    await send(`/api/media-sources/${mediaSourceId}/probe`, {});
    closeSourceInspector();
    navigate(state.activeView);
  } catch (error) {
    alert(error.message);
  } finally {
    if (button) {
      button.disabled = false;
      button.textContent = original;
    }
  }
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

async function refreshMetadataBatch(kind) {
  const button = event?.currentTarget;
  const original = button?.textContent;
  if (button) {
    button.disabled = true;
    button.textContent = "Refreshing...";
  }
  try {
    const result = await send("/api/metadata/refresh-batch", { kind, limit: 25 });
    if (result.warnings?.length) alert(result.warnings.slice(0, 5).join("\n"));
    refreshLiveViews();
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

function automationStatus(config = {}, scans = [], probes = []) {
  const mode = config.librarySyncMode || "daily";
  const activeScan = scans.find(isActiveJob);
  const activeProbe = probes.find(isActiveJob);
  const modeLabel = { manual: "Manual", hourly: "Hourly", daily: "Daily", custom: "Custom" }[mode] || mode;
  const interval = Number(config.syncIntervalMins || 1440);
  return `<div class="automation-status">
    <div class="automation-primary"><strong>${escapeHTML(modeLabel)} library care</strong><span>${mode === "manual" ? "Automatic scans are off." : `Runs every ${formatInterval(interval)} after server start.`}</span></div>
    <div class="quality-grid">
      ${qualityTile("Scan", activeScan ? "Running" : "Ready", activeScan?.lastPath || "Fast library discovery")}
      ${qualityTile("Media checks", activeProbe ? "Running" : "Batched", activeProbe?.lastPath || `${config.probeBatchLimit || 50} files per pass`)}
      ${qualityTile("Storage impact", "Low", "Separate scan and probe queues")}
    </div>
  </div>`;
}

function formatInterval(minutes) {
  const value = Number(minutes || 0);
  if (value >= 1440 && value % 1440 === 0) return `${value / 1440} day${value === 1440 ? "" : "s"}`;
  if (value >= 60 && value % 60 === 0) return `${value / 60} hour${value === 60 ? "" : "s"}`;
  return `${value} minutes`;
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
  return `<div class="session-list">${items.slice(0, 4).map(session => sessionMonitorCard(session)).join("")}</div>`;
}

function sessionMonitorCard(session = {}) {
  const percent = sessionPercent(session);
  const title = session.title || session.sourceName || session.mediaSourceId || "Active playback";
  const sourceLine = sessionSourceLine(session);
  return `<article class="session-card rich-session">
    <div class="session-art" ${session.artworkUrl ? `style="background-image:url('${escapeAttr(session.artworkUrl)}')"` : ""} data-initial="${escapeAttr(initials(title))}"></div>
    <div class="session-main">
      <strong>${escapeHTML(title)}</strong>
      <span>${escapeHTML(sourceLine)}</span>
      <small>${escapeHTML(sessionRouteLine(session))}</small>
    </div>
    <div class="session-side">
      <b>${percent}%</b>
      <small>${escapeHTML(session.serverImpact || "Route pending")}</small>
    </div>
    <i style="--progress:${Math.max(3, percent)}%"></i>
    <small class="session-time">${formatDuration(session.progressSeconds)} / ${formatDuration(session.durationSeconds)}</small>
  </article>`;
}

function sessionPercent(session = {}) {
  return session.durationSeconds ? Math.min(100, Math.round((session.progressSeconds / session.durationSeconds) * 100)) : 0;
}

function sessionSourceLine(session = {}) {
  const parts = [
    session.qualityLabel,
    session.videoCodec ? String(session.videoCodec).toUpperCase() : "",
    session.container ? String(session.container).toUpperCase() : "",
    session.bitrate ? formatBitrate(session.bitrate) : "",
  ].filter(Boolean);
  return parts.length ? parts.join(" - ") : session.sourceName || "Source details pending";
}

function sessionRouteLine(session = {}) {
  const route = session.route || session.mode || "route pending";
  const client = session.clientProfile || session.deviceId || "client";
  const status = session.status || "playing";
  return `${route} - ${client} - ${status}`;
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
  return { data: "Database", transcode: "Transcode temp", downloads: "Optimized versions", metadata: "Metadata cache", cache: "App cache", temp: "Scratch temp" }[value] || value;
}

function movieOverview(movie = {}) {
  if (movie.metadata?.overview) return escapeHTML(movie.metadata.overview);
  if (movie.needsReview) return "This item needs metadata review before Vyrden can fully trust its match.";
  return "Local file ready in your library. Vyrden will show richer cast, artwork, ratings, and playback details as metadata and file checks complete.";
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

function formatBitrate(value) {
  const bitrate = Number(value || 0);
  if (!bitrate) return "";
  if (bitrate >= 1000000) return `${Math.round(bitrate / 1000000)} Mbps`;
  if (bitrate >= 1000) return `${Math.round(bitrate / 1000)} Kbps`;
  return `${Math.round(bitrate)} bps`;
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
  return `<button class="shelf-poster" type="button" onclick="navigate('${index < 1 ? "movies" : index < 2 ? "tv" : "health"}')" data-initial="${escapeAttr(initials(label))}" style="--tone:${index}">
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

async function reviewWorkbench(items = []) {
  if (!items.length) return empty("No review items.");
  const rows = await Promise.all(items.slice(0, 40).map(async item => {
    const metadata = await api(`/api/metadata/${item.kind}/${item.id}`).catch(() => ({ best: null, records: [] }));
    const best = metadata.best || null;
    return `<article class="review-workbench-row">
      <div>
        <strong>${escapeHTML(item.title || "Untitled")}</strong>
        <span>${escapeHTML(item.kind)} - ${escapeHTML(item.reviewReason || "Needs review")}</span>
        <small>Current: ${escapeHTML(best?.title || "No selected metadata")} ${best?.year ? `(${best.year})` : ""} - ${escapeHTML(providerLabel(best?.provider || "none"))} - ${metadataConfidenceLabel(best)}</small>
      </div>
      <form onsubmit="quickFixMetadata(event,'${item.kind}','${item.id}')">
        <input name="title" value="${escapeAttr(best?.title || item.title || "")}" placeholder="Correct title" autocomplete="off">
        ${item.kind === "movie" ? `<input name="year" value="${escapeAttr(best?.year || "")}" placeholder="Year" inputmode="numeric">` : ""}
        <button class="primary" type="submit">Apply</button>
      </form>
      <div class="inline-actions">
        <button onclick="refreshMetadata('${item.kind}','${item.id}','${escapeAttr(best?.title || item.title || "")}',${Number(best?.year || 0)})">Fetch</button>
        <button onclick="openMetadataFix('${item.kind}','${item.id}','${escapeAttr(best?.title || item.title || "")}',${Number(best?.year || 0)})">Manual</button>
      </div>
    </article>`;
  }));
  return `<div class="review-workbench">${rows.join("")}</div>`;
}

async function quickFixMetadata(event, kind, id) {
  event.preventDefault();
  const data = new FormData(event.currentTarget);
  const title = String(data.get("title") || "").trim();
  if (!title) return;
  const year = kind === "movie" ? Number(data.get("year") || 0) || 0 : 0;
  await send("/api/metadata/match", { kind, id, title, year, review: false }, "PUT");
  navigate("health");
}

function versionGroups(items = [], totalTitles = 0) {
  if (!items.length) return empty(totalTitles ? "No duplicate version groups found." : "No version groups yet. Scan libraries to build movie and TV source intelligence.");
  return `<div class="version-grid">${items.map(item => `<article class="version-card">
    <div><strong>${escapeHTML(item.title)}</strong><small>${escapeHTML(item.kind)} - ${item.versionCount} versions</small></div>
  </article>`).join("")}</div>`;
}

function sessionCards(items = []) {
  if (!items.length) return empty("No active sessions.");
  return `<div class="version-grid">${items.map(item => `<article class="version-card session-summary">
    <div><strong>${escapeHTML(item.title || item.sourceName || item.deviceId || "Active playback")}</strong><small>${escapeHTML(sessionRouteLine(item))}</small></div>
    <small>${escapeHTML(sessionSourceLine(item))}</small>
    <i style="--progress:${Math.max(3, sessionPercent(item))}%"></i>
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
      <button onclick="openSourceInspector('${item.id}')">Inspect</button>
      ${item.probed ? "" : `<button onclick="probeSource('${item.id}')">Probe</button>`}
      <button onclick="startWork('${item.id}','remux')">Repackage while playing</button>
      <button onclick="startWork('${item.id}','transcode')">Convert while playing</button>
      <button onclick="startDownload('${item.id}','balanced')">Create remote optimized version</button>
      <a class="button primary" href="/play/${item.id}" target="_blank">Play</a>
    </div>
  </article>`).join("")}</div>`;
}

function downloadCards(items = []) {
  if (!items.length) return empty("No prepared downloads yet. Start one from a movie or episode source.");
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

function artworkURL(kind, id, type = "poster") {
  return `/api/artwork/${encodeURIComponent(kind)}/${encodeURIComponent(id)}?style=neutral&type=${encodeURIComponent(type)}`;
}

function initials(value) {
  const words = String(value || "V").trim().split(/\s+/).filter(Boolean);
  return words.slice(0, 2).map(word => word[0]).join("").toUpperCase() || "V";
}

function escapeHTML(value) {
  return String(value ?? "").replace(/[&<>"']/g, char => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "\"": "&quot;", "'": "&#039;" }[char]));
}

let dashboardRefreshTimer = 0;
let livePatchTimer = 0;
const livePatchEvents = new Set([
  "scan.progress", "probe.running", "transcode.running", "download.running",
  "session.started", "session.updated", "session.stopped", "playback.state.updated"
]);
const dashboardStructuralEvents = new Set([
  "scan.queued", "scan.started", "scan.completed", "scan.failed",
  "probe.queued", "probe.started", "probe.completed", "probe.failed",
  "transcode.queued", "transcode.started", "transcode.completed", "transcode.failed",
  "download.queued", "download.started", "download.completed", "download.failed",
  "metadata.updated", "metadata.ratings.updated", "metadata.batch.completed", "metadata.batch.failed",
  "settings.updated", "library.updated", "library.deleted"
]);
function refreshLiveViews(eventName = "") {
  if (!["dashboard", "activity", "health", "playback"].includes(state.activeView)) return;
  if (state.activeView === "activity") {
    patchLiveView(eventName);
    return;
  }
  if (state.activeView === "playback" && eventName.startsWith("download.")) {
    patchLiveView(eventName);
    return;
  }
  if (livePatchEvents.has(eventName)) {
    patchLiveView(eventName);
    return;
  }
  if (state.activeView === "dashboard" && eventName && !dashboardStructuralEvents.has(eventName)) return;
  clearTimeout(dashboardRefreshTimer);
  dashboardRefreshTimer = setTimeout(() => navigate(state.activeView), 180);
}

function patchLiveView(eventName = "") {
  clearTimeout(livePatchTimer);
  livePatchTimer = setTimeout(async () => {
    if (state.activeView === "dashboard") await patchDashboardLive(eventName);
    if (state.activeView === "activity") await patchActivityLive();
    if (state.activeView === "playback") await patchPlaybackLive();
  }, 180);
}

async function patchDashboardLive(eventName = "") {
  const [sessions, scans, probes, work, downloads, system, summary, health, sources] = await Promise.all([
    api("/api/sessions"),
    api("/api/scans"),
    api("/api/probes"),
    api("/api/work"),
    api("/api/downloads"),
    api("/api/system/status"),
    api("/api/catalog/summary"),
    api("/api/catalog/health"),
    api("/api/media-sources?limit=200"),
  ]);
  const activeSessions = sessions.sessions || [];
  const scanJobs = scans.scans || [];
  const probeJobs = probes.probes || [];
  const workJobs = work.work || [];
  const downloadJobs = downloads.downloads || [];
  const mediaSources = sources.mediaSources || [];
  const quality = sourceQuality(mediaSources);
  const activeJobs = [...scanJobs, ...probeJobs, ...workJobs, ...downloadJobs].filter(isActiveJob);
  const directPlayable = summary.mediaSources ? Math.max(0, Math.round(((summary.mediaSources - (health.unsupported || 0)) / summary.mediaSources) * 100)) : 0;
  setLiveHTML("dashboard-title", dashboardCommandTitle(activeSessions, summary));
  setLiveHTML("dashboard-copy", dashboardCommandCopy(activeSessions, summary, health, quality));
  setLiveHTML("playing-now", playingNow(activeSessions));
  setLiveHTML("playing-now-status", activeSessions.length ? "Live" : "Idle");
  setLiveHTML("operations", operationsPanel(scanJobs, probeJobs, workJobs, downloadJobs));
  setLiveHTML("operations-status", activeJobs.length ? `${activeJobs.length} running` : "Clear");
  setLiveTone("operations-status", activeJobs.length ? "warn" : "good");
  setLiveHTML("hardware", hardwarePanel(system));
  setLiveHTML("hardware-cores", `${system.cpu?.cores || 0} cores`);
  setLiveHTML("recent-scans", jobCards(scanJobs));
  setLiveHTML("dashboard-snapshot", dashboardSnapshot({ activeSessions, activeJobs, summary, health, quality, system }));
  setLiveHTML("live-signals", `${signalPill("Sessions", activeSessions.length)}${signalPill("Jobs", activeJobs.length)}${signalPill("Downloads", downloadJobs.length)}${signalPill("Review", health.needsReview || 0)}${signalPill("CPU", `${Math.round(system.cpu?.percent || 0)}%`)}`);
  setLiveHTML("updated-at", `Updated ${new Date().toLocaleTimeString()}`);
  const playableBadge = document.querySelector(".feature-content .meta-line .badge:nth-child(4)");
  if (playableBadge) {
    playableBadge.textContent = `${directPlayable}% playable`;
    playableBadge.classList.toggle("good", directPlayable >= 90);
    playableBadge.classList.toggle("warn", directPlayable < 90);
  }
}

async function patchActivityLive() {
  const [scans, probes, work, downloads, sessions] = await Promise.all([api("/api/scans"), api("/api/probes"), api("/api/work"), api("/api/downloads"), api("/api/sessions")]);
  setLiveHTML("activity-signals", `${signalPill("Sessions", sessions.sessions.length)}${signalPill("Scans", scans.scans.length)}${signalPill("Work jobs", work.work.length)}${signalPill("Downloads", downloads.downloads.length)}`);
  setLiveHTML("activity-sessions", sessionCards(sessions.sessions));
  setLiveHTML("activity-work", jobCards(work.work));
  setLiveHTML("activity-scans", jobCards(scans.scans));
  setLiveHTML("activity-probes", jobCards(probes.probes));
  setLiveHTML("activity-downloads", downloadCards(downloads.downloads));
}

async function patchPlaybackLive() {
  const downloads = await api("/api/downloads");
  setLiveHTML("playback-downloads", downloadCards(downloads.downloads));
}

function setLiveHTML(name, html) {
  const target = document.querySelector(`[data-live="${name}"]`);
  if (target) target.innerHTML = html;
}

function setLiveTone(name, tone) {
  const target = document.querySelector(`[data-live="${name}"]`);
  if (!target) return;
  target.classList.toggle("warn", tone === "warn");
  target.classList.toggle("good", tone === "good");
}

const liveEventNames = [
  "scan.queued", "scan.started", "scan.progress", "scan.completed", "scan.failed",
  "probe.queued", "probe.started", "probe.running", "probe.completed", "probe.failed",
  "transcode.queued", "transcode.started", "transcode.running", "transcode.completed", "transcode.failed",
  "download.queued", "download.started", "download.running", "download.completed", "download.failed",
  "session.started", "session.updated", "session.stopped",
  "playback.state.updated", "metadata.updated", "metadata.ratings.updated", "metadata.batch.completed", "metadata.batch.failed",
  "settings.updated", "library.updated", "library.deleted"
];
const events = new EventSource("/api/events");
for (const name of liveEventNames) {
  events.addEventListener(name, () => {
    refreshLiveViews(name);
  });
}
setInterval(() => {
  if (state.activeView === "dashboard") patchLiveView("live.tick");
  if (state.activeView === "activity") patchLiveView("live.tick");
}, 2500);

refreshShell().catch(() => {});
navigate("dashboard");
