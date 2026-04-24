const state = { activeView: "dashboard", activeSession: "" };

const view = document.getElementById("view");
const viewTitle = document.getElementById("viewTitle");
const serverStatus = document.getElementById("serverStatus");
const serverDot = document.getElementById("serverDot");

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
  return { dashboard: "Dashboard", movies: "Movies", tv: "TV", activity: "Activity", health: "Health", playback: "Playback Lab", settings: "Settings" }[name] || name;
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
  if (name === "activity") return renderActivity();
  if (name === "health") return renderHealth();
  if (name === "playback") return renderPlaybackLab();
  if (name === "settings") return renderSettings();
}

async function renderDashboard() {
  const [summary, libraries, scans, probes, recent, sessions] = await Promise.all([
    api("/api/catalog/summary"),
    api("/api/libraries"),
    api("/api/scans"),
    api("/api/probes"),
    api("/api/playback/recent"),
    api("/api/sessions"),
  ]);
  view.innerHTML = `
    <div class="stack">
      <div class="grid">
        ${metric("Libraries", summary.libraries)}
        ${metric("Media", summary.mediaSources)}
        ${metric("Movies", summary.movies)}
        ${metric("Series", summary.series)}
        ${metric("Episodes", summary.episodes)}
        ${metric("Active", sessions.sessions.length)}
      </div>
      <div class="hero-strip">
        <div><span>Vyrden Signal</span><strong>${summary.unprobed || 0} sources need probing</strong></div>
        <button class="primary" onclick="navigate('movies')">Browse Library</button>
      </div>
      <div class="grid two">
        <div class="card"><h2>Resume</h2>${mediaCards(recent.recent, "recent")}</div>
        <div class="card"><h2>Libraries</h2>${table(["Name", "Path", "Storage"], libraries.libraries.map(item => [item.name, item.path, item.storageType]))}</div>
      </div>
      <div class="grid two">
        <div class="card"><h2>Recent Scans</h2>${jobTable(scans.scans)}</div>
        <div class="card"><h2>Recent Probes</h2>${jobTable(probes.probes)}</div>
      </div>
    </div>`;
}

async function renderMovies() {
  const payload = await api("/api/movies?limit=500");
  view.innerHTML = `
    <div class="stack">
      <div class="toolbar">
        <button class="primary" onclick="startScan('/api/libraries/movies/scan')">Scan Movies</button>
        <input class="search" id="movieFilter" placeholder="Filter movies" oninput="filterCards(this.value)">
      </div>
      <div class="poster-grid">${payload.movies.map(movieCard).join("") || empty("No movies yet. Run a scan first.")}</div>
    </div>`;
}

function movieCard(item) {
  return `<article class="poster-card" data-filter="${escapeAttr(item.title)} ${item.year || ""}">
    <button class="poster-action" onclick="showMovie('${item.id}')">
      <span class="poster-title">${escapeHTML(item.title)}</span>
      <small>${item.year || "Unknown year"} · ${item.versionCount || 0} version${item.versionCount === 1 ? "" : "s"}</small>
      ${item.needsReview ? `<em>Review</em>` : ""}
    </button>
  </article>`;
}

async function showMovie(id) {
  const movie = await api(`/api/movies/${id}`);
  const rows = await Promise.all(movie.versions.map(async version => {
    const state = await api(`/api/playback/state/${version.mediaSourceId}`);
    return versionRow(version, state);
  }));
  viewTitle.textContent = movie.title;
  view.innerHTML = `
    <div class="detail-shell">
      <section class="detail-hero">
        <button onclick="navigate('movies')">Back</button>
        <h2>${escapeHTML(movie.title)}</h2>
        <p>${movie.year || "Unknown year"} · ${movie.versionCount} version${movie.versionCount === 1 ? "" : "s"}</p>
        ${movie.needsReview ? `<button onclick="openMetadataFix('movie','${movie.id}','${escapeAttr(movie.title)}',${movie.year || 0})">Fix Match</button>` : ""}
      </section>
      <section class="card"><h2>Versions</h2>${table(["Quality", "File", "State", "Controls"], rows)}</section>
    </div>`;
}

function versionRow(version, state) {
  const watched = state.watched ? "Watched" : state.progressSeconds > 0 ? `Resume ${Math.round(state.percent * 100)}%` : "Unplayed";
  return [
    version.qualityLabel || "Source",
    version.relPath,
    watched,
    `<div class="inline-actions">
      <a class="button primary" href="/play/${version.mediaSourceId}" target="_blank">Play</a>
      <button onclick="markWatched('${version.mediaSourceId}', true)">Watched</button>
      <button onclick="markWatched('${version.mediaSourceId}', false)">Unwatched</button>
    </div>`,
  ];
}

async function renderTV() {
  const payload = await api("/api/series?limit=500");
  view.innerHTML = `
    <div class="stack">
      <div class="toolbar">
        <button class="primary" onclick="startScan('/api/libraries/tv/scan')">Scan TV</button>
        <input class="search" placeholder="Filter shows" oninput="filterCards(this.value)">
      </div>
      <div class="poster-grid">${payload.series.map(seriesCard).join("") || empty("No TV yet. Run a scan first.")}</div>
    </div>`;
}

function seriesCard(item) {
  return `<article class="poster-card" data-filter="${escapeAttr(item.title)}">
    <button class="poster-action" onclick="showSeries('${item.id}')">
      <span class="poster-title">${escapeHTML(item.title)}</span>
      <small>${item.seasonCount || 0} seasons · ${item.episodeCount || 0} episodes</small>
    </button>
  </article>`;
}

async function showSeries(id) {
  const series = await api(`/api/series/${id}`);
  viewTitle.textContent = series.title;
  view.innerHTML = `
    <div class="detail-shell">
      <section class="detail-hero">
        <button onclick="navigate('tv')">Back</button>
        <h2>${escapeHTML(series.title)}</h2>
        <p>${series.seasonCount} seasons · ${series.episodeCount} episodes</p>
      </section>
      <div class="stack">
        ${series.seasons.map(season => `<section class="card"><h2>Season ${season.seasonNumber}</h2>${episodeList(season.episodes)}</section>`).join("")}
      </div>
    </div>`;
}

function episodeList(episodes) {
  if (!episodes.length) return empty("No episodes in this season.");
  return episodes.map(episode => {
    const version = episode.versions && episode.versions[0];
    const play = version ? `<a class="button primary" href="/play/${version.mediaSourceId}" target="_blank">Play</a><button onclick="markWatched('${version.mediaSourceId}', true)">Watched</button>` : `<span class="muted">No source</span>`;
    const label = episode.episodeEnd && episode.episodeEnd !== episode.episodeNumber ? `E${episode.episodeNumber}-E${episode.episodeEnd}` : `E${episode.episodeNumber}`;
    return `<div class="episode-row">
      <strong>${label}</strong>
      <span>${escapeHTML(episode.title || "Episode")}</span>
      <small>${episode.versionCount || 0} version${episode.versionCount === 1 ? "" : "s"}</small>
      <div class="inline-actions">${play}${episode.needsReview ? `<button onclick="openMetadataFix('episode','${episode.id}','${escapeAttr(episode.title || "Episode")}',0)">Fix</button>` : ""}</div>
    </div>`;
  }).join("");
}

async function renderActivity() {
  const [scans, probes, work, sessions] = await Promise.all([api("/api/scans"), api("/api/probes"), api("/api/work"), api("/api/sessions")]);
  view.innerHTML = `
    <div class="stack">
      <div class="toolbar">
        <button class="primary" onclick="startScan('/api/libraries/scan')">Scan Movies + TV</button>
        <button onclick="startProbe()">Probe Unprobed</button>
      </div>
      <div class="grid two">
        <div class="card"><h2>Active Sessions</h2>${sessionTable(sessions.sessions)}</div>
        <div class="card"><h2>Work</h2>${jobTable(work.work)}</div>
      </div>
      <div class="grid two">
        <div class="card"><h2>Scans</h2>${jobTable(scans.scans)}</div>
        <div class="card"><h2>Probes</h2>${jobTable(probes.probes)}</div>
      </div>
    </div>`;
}

async function renderHealth() {
  const [health, review, versions] = await Promise.all([api("/api/catalog/health"), api("/api/review"), api("/api/versions")]);
  view.innerHTML = `
    <div class="stack">
      <div class="grid">
        ${metric("Needs Review", health.needsReview)}
        ${metric("Unprobed", health.unprobed)}
        ${metric("Unsupported", health.unsupported)}
        ${metric("High Bitrate", health.highBitrate)}
        ${metric("Subtitles", health.withSubtitles)}
        ${metric("Versions", versions.versions.length)}
      </div>
      <div class="grid two">
        <div class="card"><h2>Review Queue</h2>${reviewTable(review.items)}</div>
        <div class="card"><h2>Version Groups</h2>${table(["Kind", "Title", "Versions"], versions.versions.map(item => [item.kind, item.title, item.versionCount]))}</div>
      </div>
    </div>`;
}

async function renderPlaybackLab() {
  const [payload, profiles] = await Promise.all([api("/api/media-sources?limit=200"), api("/api/devices/profiles")]);
  view.innerHTML = `<div class="stack">
    <div class="card"><h2>Client Profiles</h2>${table(["Profile", "Containers", "Video"], profiles.profiles.map(item => [item.name, item.containers.join(", "), item.videoCodecs.join(", ")]))}</div>
    <div class="card"><h2>Playback Lab</h2>${table(["Name", "Kind", "Probed", "Decision", "Work", "Player"], payload.mediaSources.map(item => [
      item.name,
      item.kind,
      item.probed ? "Yes" : "No",
      link(`/api/playback/decision?mediaSourceId=${item.id}&clientProfile=web`, "Decision"),
      `<button onclick="startWork('${item.id}','remux')">Remux</button> <button onclick="startWork('${item.id}','transcode')">Transcode</button>`,
      link(`/play/${item.id}`, "Open"),
    ]))}</div>
  </div>`;
}

async function renderSettings() {
  const payload = await api("/api/settings/performance");
  view.innerHTML = `<div class="stack">
    <div class="card"><h2>Performance</h2><pre>${escapeHTML(JSON.stringify(payload, null, 2))}</pre></div>
    <div class="card"><h2>Runtime Direction</h2><p class="muted">Windows tray app first, service mode later for advanced installs. Local, USB, NAS, and mapped storage stay first-class.</p></div>
  </div>`;
}

async function markWatched(mediaSourceId, watched) {
  await send(`/api/playback/state/${mediaSourceId}`, { watched, progressSeconds: 0 }, "PUT");
  navigate(state.activeView);
}

async function startWork(mediaSourceId, mode) {
  await send("/api/work", { mediaSourceId, mode });
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

function openMetadataFix(kind, id, currentTitle, year) {
  const title = prompt("Correct title", currentTitle);
  if (!title) return;
  const parsedYear = kind === "movie" ? Number(prompt("Year", year || "")) || 0 : 0;
  send("/api/metadata/match", { kind, id, title, year: parsedYear, review: false }, "PUT").then(() => navigate(state.activeView));
}

function filterCards(value) {
  const needle = value.trim().toLowerCase();
  document.querySelectorAll("[data-filter]").forEach(card => {
    card.hidden = needle && !card.dataset.filter.toLowerCase().includes(needle);
  });
}

function metric(label, value) {
  return `<div class="card metric"><span>${escapeHTML(label)}</span><strong>${value ?? 0}</strong></div>`;
}

function mediaCards(items = []) {
  if (!items.length) return empty("Nothing to resume yet.");
  return `<div class="mini-list">${items.slice(0, 6).map(item => `<a href="/play/${item.mediaSourceId}" target="_blank"><strong>${escapeHTML(item.name)}</strong><span>${Math.round((item.percent || 0) * 100)}% · ${escapeHTML(item.kind)}</span></a>`).join("")}</div>`;
}

function reviewTable(items = []) {
  return table(["Kind", "Title", "Reason", "Fix"], items.map(item => [
    item.kind,
    item.title,
    item.reviewReason,
    `<button onclick="openMetadataFix('${item.kind}','${item.id}','${escapeAttr(item.title)}',0)">Fix Match</button>`,
  ]));
}

function sessionTable(items = []) {
  return table(["Device", "Media", "Status", "Progress"], items.map(item => [item.deviceId, item.mediaSourceId, item.status, Math.round(item.progressSeconds || 0) + "s"]));
}

function jobTable(items = []) {
  return table(["ID", "Status", "Done", "Last"], items.slice(0, 8).map(item => [item.id || "", item.status || "", item.completed ?? item.mediaFiles ?? "", item.lastPath || item.outputPath || ""]));
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

function escapeHTML(value) {
  return String(value ?? "").replace(/[&<>"']/g, char => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "\"": "&quot;", "'": "&#039;" }[char]));
}

const events = new EventSource("/api/events");
for (const name of ["scan.completed", "probe.completed", "transcode.completed", "session.started", "session.updated", "session.stopped", "playback.state.updated", "metadata.updated"]) {
  events.addEventListener(name, () => {
    if (["dashboard", "activity", "health"].includes(state.activeView)) navigate(state.activeView);
  });
}

refreshShell().catch(() => {});
navigate("dashboard");
