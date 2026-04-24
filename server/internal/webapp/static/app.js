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
  return { dashboard: "Home", movies: "Movies", tv: "TV", libraries: "Libraries", activity: "Activity", health: "Review", playback: "Playback Lab", remote: "Remote Access", settings: "Settings" }[name] || name;
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
      <div class="hero-strip">
        <div class="hero-copy">
          <span>Vyrden Signal</span>
          <strong>Your library is local, fast, and under control.</strong>
          <p>${summary.movies || 0} movies, ${summary.series || 0} series, ${summary.episodes || 0} episodes. ${summary.unprobed || 0} sources are waiting for probe analysis.</p>
          <div class="detail-actions">
            <button class="primary" onclick="navigate('movies')">Browse Movies</button>
            <button class="ghost" onclick="navigate('tv')">Browse TV</button>
          </div>
        </div>
        <div class="signal-stack">
          ${signalPill("Libraries", summary.libraries)}
          ${signalPill("Active sessions", sessions.sessions.length)}
          ${signalPill("Media sources", summary.mediaSources)}
        </div>
      </div>
      <div class="grid two">
        <div class="card"><h2>Resume</h2>${mediaCards(recent.recent, "recent")}</div>
        <div class="card"><h2>Storage</h2>${libraryCards(libraries.libraries)}</div>
      </div>
      <div class="grid three">
        ${metric("Movies", summary.movies)}
        ${metric("Series", summary.series)}
        ${metric("Episodes", summary.episodes)}
      </div>
      <div class="grid two">
        <div class="card"><h2>Recent Scans</h2>${jobCards(scans.scans)}</div>
        <div class="card"><h2>Recent Probes</h2>${jobCards(probes.probes)}</div>
      </div>
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
  return `<article class="poster-card" data-filter="${escapeAttr(item.title)} ${item.year || ""}">
    <img alt="" src="/api/artwork/movie/${item.id}" loading="lazy">
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
  viewTitle.textContent = movie.title;
  view.innerHTML = `
    <div class="detail-shell">
      <section class="detail-hero">
        <img alt="" src="/api/artwork/movie/${movie.id}">
        <div class="detail-poster"><img alt="" src="/api/artwork/movie/${movie.id}"></div>
        <div class="detail-copy">
          <button class="ghost" onclick="navigate('movies')">Back to Movies</button>
          <h2>${escapeHTML(movie.title)}</h2>
          <p>${movie.year || "Unknown year"} - ${movie.versionCount} version${movie.versionCount === 1 ? "" : "s"} - ${movie.needsReview ? "Needs metadata review" : "Ready"}</p>
          <div class="detail-actions">
            ${movie.versions[0] ? `<a class="button primary" href="/play/${movie.versions[0].mediaSourceId}" target="_blank">Play</a>` : ""}
            ${movie.versions[0] ? `<button onclick="markWatched('${movie.versions[0].mediaSourceId}', true)">Mark Watched</button>` : ""}
            ${movie.needsReview ? `<button onclick="openMetadataFix('movie','${movie.id}','${escapeAttr(movie.title)}',${movie.year || 0})">Fix Match</button>` : ""}
          </div>
        </div>
      </section>
      <section class="card"><h2>Available Versions</h2><div class="version-grid">${rows.join("") || empty("No playable versions found.")}</div></section>
    </div>`;
}

function versionCard(version, state, selected) {
  const watched = state.watched ? "Watched" : state.progressSeconds > 0 ? `Resume ${Math.round(state.percent * 100)}%` : "Unplayed";
  return `<article class="version-card">
    <div>
      <strong>${escapeHTML(version.qualityLabel || "Source")}${selected ? " - selected" : ""}</strong>
      <small>${escapeHTML(version.relPath)} - ${watched}</small>
    </div>
    <div class="inline-actions">
      <a class="button primary" href="/play/${version.mediaSourceId}" target="_blank">Play</a>
      <button onclick="markWatched('${version.mediaSourceId}', true)">Watched</button>
      <button onclick="markWatched('${version.mediaSourceId}', false)">Unwatched</button>
    </div>
  </article>`;
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
  return `<article class="poster-card" data-filter="${escapeAttr(item.title)}">
    <img alt="" src="/api/artwork/series/${item.id}" loading="lazy">
    <button class="poster-action" onclick="showSeries('${item.id}')">
      <span class="poster-title">${escapeHTML(item.title)}</span>
      <small>${item.seasonCount || 0} seasons - ${item.episodeCount || 0} episodes</small>
    </button>
  </article>`;
}

async function showSeries(id) {
  const series = await api(`/api/series/${id}`);
  viewTitle.textContent = series.title;
  view.innerHTML = `
    <div class="detail-shell">
      <section class="detail-hero">
        <img alt="" src="/api/artwork/series/${series.id}">
        <div class="detail-poster"><img alt="" src="/api/artwork/series/${series.id}"></div>
        <div class="detail-copy">
          <button class="ghost" onclick="navigate('tv')">Back to TV</button>
          <h2>${escapeHTML(series.title)}</h2>
          <p>${series.seasonCount} seasons - ${series.episodeCount} episodes</p>
        </div>
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
    <div><strong>${escapeHTML(item.name)}</strong><span>${escapeHTML(item.kind)} - ${escapeHTML(item.storageType)}</span></div>
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
  const [scans, probes, work, sessions] = await Promise.all([api("/api/scans"), api("/api/probes"), api("/api/work"), api("/api/sessions")]);
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
  const [payload, profiles] = await Promise.all([api("/api/media-sources?limit=200"), api("/api/devices/profiles")]);
  view.innerHTML = `<div class="stack">
    <div class="card"><h2>Client Profiles</h2>${profileCards(profiles.profiles)}</div>
    <div class="card"><h2>Playback Lab</h2>${playbackCards(payload.mediaSources)}</div>
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
  const [settings, performance] = await Promise.all([api("/api/settings"), api("/api/settings/performance")]);
  view.innerHTML = `<div class="stack">
    <div class="hero-strip">
      <div class="hero-copy"><span>Runtime</span><strong>Local-first settings before tray and installer.</strong><p>These values are persisted locally and give the future tray app a real product runtime to control.</p></div>
      <div class="signal-stack">${signalPill("Profile", performance.profile)}${signalPill("Queues", performance.queues.length)}${signalPill("Libraries", settings.libraries.length)}</div>
    </div>
    <div class="card"><h2>Server Config</h2>${settingsGrid(settings.config)}</div>
    <div class="card"><h2>Performance</h2><pre>${escapeHTML(JSON.stringify(performance, null, 2))}</pre></div>
  </div>`;
}

function settingsGrid(config) {
  return `<div class="settings-grid">${Object.entries(config).map(([key, value]) => `<div><span>${escapeHTML(key)}</span><strong>${escapeHTML(value)}</strong></div>`).join("")}</div>`;
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

function signalPill(label, value) {
  return `<div class="signal-pill"><small>${escapeHTML(label)}</small><b>${escapeHTML(value ?? 0)}</b></div>`;
}

function mediaCards(items = []) {
  if (!items.length) return empty("Nothing to resume yet.");
  return `<div class="mini-list">${items.slice(0, 6).map(item => `<a href="/play/${item.mediaSourceId}" target="_blank"><strong>${escapeHTML(item.name)}</strong><span>${Math.round((item.percent || 0) * 100)}% - ${escapeHTML(item.kind)}</span></a>`).join("")}</div>`;
}

function reviewCards(items = []) {
  if (!items.length) return empty("No review items.");
  return `<div class="version-grid">${items.map(item => `<article class="version-card">
    <div><strong>${escapeHTML(item.title)}</strong><small>${escapeHTML(item.kind)} - ${escapeHTML(item.reviewReason)}</small></div>
    <button onclick="openMetadataFix('${item.kind}','${item.id}','${escapeAttr(item.title)}',0)">Fix Match</button>
  </article>`).join("")}</div>`;
}

function versionGroups(items = []) {
  if (!items.length) return empty("No duplicate version groups.");
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
  if (!items || !items.length) return empty("No jobs yet.");
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
      <a class="button primary" href="/play/${item.id}" target="_blank">Play</a>
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
