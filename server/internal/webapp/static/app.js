const state = {
  activeView: "dashboard",
  activeScanId: "",
};

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
  return {
    dashboard: "Dashboard",
    movies: "Movies",
    tv: "TV",
    activity: "Activity",
    health: "Health",
    playback: "Playback Lab",
    settings: "Settings",
  }[name] || name;
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
  const [summary, libraries, scans, probes] = await Promise.all([
    api("/api/catalog/summary"),
    api("/api/libraries"),
    api("/api/scans"),
    api("/api/probes"),
  ]);
  view.innerHTML = `
    <div class="stack">
      <div class="grid">
        ${metric("Libraries", summary.libraries)}
        ${metric("Media", summary.mediaSources)}
        ${metric("Movies", summary.movies)}
        ${metric("Series", summary.series)}
        ${metric("Episodes", summary.episodes)}
        ${metric("Unprobed", summary.unprobed)}
      </div>
      <div class="card">
        <h2>Libraries</h2>
        ${table(["Name", "Path", "Storage"], libraries.libraries.map(item => [item.name, item.path, item.storageType]))}
      </div>
      <div class="grid two">
        <div class="card"><h2>Recent Scans</h2>${jobTable(scans.scans)}</div>
        <div class="card"><h2>Recent Probes</h2>${jobTable(probes.probes)}</div>
      </div>
    </div>`;
}

async function renderMovies() {
  const payload = await api("/api/movies?limit=200");
  view.innerHTML = `<div class="card"><h2>Movies</h2>${table(["Title", "Year", "Versions", "Review"], payload.movies.map(item => [
    link(`/api/movies/${item.id}`, item.title),
    item.year || "",
    item.versionCount || 0,
    item.needsReview ? "Yes" : "No",
  ]))}</div>`;
}

async function renderTV() {
  const payload = await api("/api/series?limit=200");
  view.innerHTML = `<div class="card"><h2>Series</h2>${table(["Series", "Seasons", "Episodes"], payload.series.map(item => [
    link(`/api/series/${item.id}`, item.title),
    item.seasonCount || 0,
    item.episodeCount || 0,
  ]))}</div>`;
}

async function renderActivity() {
  const [scans, probes, work] = await Promise.all([api("/api/scans"), api("/api/probes"), api("/api/work")]);
  view.innerHTML = `
    <div class="stack">
      <div class="toolbar">
        <button class="primary" onclick="startScan('/api/libraries/scan')">Scan Movies + TV</button>
        <button onclick="startProbe()">Probe Unprobed</button>
      </div>
      <div class="grid three">
        <div class="card"><h2>Scans</h2>${jobTable(scans.scans)}</div>
        <div class="card"><h2>Probes</h2>${jobTable(probes.probes)}</div>
        <div class="card"><h2>Work</h2>${jobTable(work.work)}</div>
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
        <div class="card"><h2>Review Queue</h2>${table(["Kind", "Title", "Reason"], review.items.map(item => [item.kind, item.title, item.reviewReason]))}</div>
        <div class="card"><h2>Version Groups</h2>${table(["Kind", "Title", "Versions"], versions.versions.map(item => [item.kind, item.title, item.versionCount]))}</div>
      </div>
    </div>`;
}

async function renderPlaybackLab() {
  const payload = await api("/api/media-sources?limit=100");
  view.innerHTML = `<div class="card"><h2>Playback Lab</h2>${table(["Name", "Kind", "Probed", "Decision", "Player"], payload.mediaSources.map(item => [
    item.name,
    item.kind,
    item.probed ? "Yes" : "No",
    link(`/api/playback/decision?mediaSourceId=${item.id}&clientProfile=web`, "Decision"),
    link(`/play/${item.id}`, "Open"),
  ]))}</div>`;
}

async function renderSettings() {
  const payload = await api("/api/settings/performance");
  view.innerHTML = `<div class="stack">
    <div class="card"><h2>Performance</h2><pre>${escapeHTML(JSON.stringify(payload, null, 2))}</pre></div>
    <div class="card"><h2>Runtime</h2><pre>Config file and installer-managed settings are the next packaging step.</pre></div>
  </div>`;
}

async function startScan(path) {
  await api(path, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ sampleLimit: 50 }) });
  navigate("activity");
}

async function startProbe() {
  await api("/api/probes", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ limit: 50 }) });
  navigate("activity");
}

function metric(label, value) {
  return `<div class="card metric"><span>${escapeHTML(label)}</span><strong>${value ?? 0}</strong></div>`;
}

function jobTable(items = []) {
  return table(["ID", "Status", "Done", "Last"], items.slice(0, 8).map(item => [
    item.id || "",
    item.status || "",
    item.completed ?? item.mediaFiles ?? "",
    item.lastPath || "",
  ]));
}

function table(headers, rows) {
  if (!rows || rows.length === 0) return `<p class="muted">No items yet.</p>`;
  return `<table><thead><tr>${headers.map(header => `<th>${escapeHTML(header)}</th>`).join("")}</tr></thead><tbody>${rows.map(row => `<tr>${row.map(cell => `<td>${cell}</td>`).join("")}</tr>`).join("")}</tbody></table>`;
}

function link(href, text) {
  return `<a href="${href}">${escapeHTML(text)}</a>`;
}

function escapeHTML(value) {
  return String(value ?? "").replace(/[&<>"']/g, char => ({
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    "\"": "&quot;",
    "'": "&#039;",
  }[char]));
}

const events = new EventSource("/api/events");
for (const name of ["scan.completed", "probe.completed", "transcode.completed"]) {
  events.addEventListener(name, () => {
    if (["dashboard", "activity", "health"].includes(state.activeView)) navigate(state.activeView);
  });
}

refreshShell().catch(() => {});
navigate("dashboard");
