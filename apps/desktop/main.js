const path = require("path");
const fs = require("fs");
const http = require("http");
const { app, BrowserWindow, Menu, Tray, shell, dialog, ipcMain } = require("electron");
const { spawn } = require("child_process");

const LOCAL_APP_URL = process.env.XUVA_DESKTOP_LOCAL_URL || "http://127.0.0.1:8097";
const SETTINGS_FILE = () => path.join(app.getPath("userData"), "xuva-settings.json");
const serverCwdDefault = path.resolve(__dirname, "..", "..", "server");

let mainWindow = null;
let setupWindow = null;
let tray = null;
let serverProcess = null;
let serverRestarting = false;
let discoveredServers = [];
let bonjourBrowser = null;
let currentMode = "local";
let currentRemoteUrl = "";
let currentRuntimeHome = "";

function logEvent(event, fields = {}) {
  const payload = { ts: new Date().toISOString(), component: "desktop-shell", event, ...fields };
  try {
    console.log(JSON.stringify(payload));
  } catch {
    console.log(`[desktop-shell] ${event}`);
  }
}

function readSettings() {
  try {
    const raw = fs.readFileSync(SETTINGS_FILE(), "utf8");
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

function writeSettings(settings) {
  try {
    fs.writeFileSync(SETTINGS_FILE(), JSON.stringify(settings, null, 2), "utf8");
  } catch (err) {
    logEvent("settings.write_error", { error: String(err) });
  }
}

function applySettings(settings) {
  currentMode = settings && settings.mode === "remote" ? "remote" : "local";
  currentRemoteUrl = (settings && settings.mode === "remote" && settings.remoteUrl) || "";
}

function runtimeRoot() {
  if (app.isPackaged) {
    return path.join(process.resourcesPath, "runtime");
  }
  return path.resolve(__dirname, "runtime");
}

function packagedServerPath() {
  return path.join(runtimeRoot(), process.platform === "win32" ? "xuva-server.exe" : "xuva-server");
}

function ensureDir(dir) {
  try {
    fs.mkdirSync(dir, { recursive: true });
    return true;
  } catch (err) {
    logEvent("directory.create_error", { dir, error: String(err) });
    return false;
  }
}

function canWriteDir(dir) {
  if (!ensureDir(dir)) return false;
  const probe = path.join(dir, ".xuva-write-test");
  try {
    fs.writeFileSync(probe, String(Date.now()), "utf8");
    fs.unlinkSync(probe);
    return true;
  } catch (err) {
    logEvent("directory.write_probe_failed", { dir, error: String(err) });
    return false;
  }
}

function defaultLocalAppData(env) {
  return env.LOCALAPPDATA || app.getPath("appData");
}

function defaultProgramData(env) {
  if (env.PROGRAMDATA) return env.PROGRAMDATA;
  return process.platform === "win32" ? "C:\\ProgramData" : defaultLocalAppData(env);
}

function resolveRuntimeHome(env) {
  const explicit = String(env.XUVA_RUNTIME_HOME || "").trim();
  if (explicit) {
    const root = path.resolve(explicit);
    if (!canWriteDir(root)) {
      logEvent("runtime_home.explicit_unavailable", { root });
      currentRuntimeHome = root;
      return { root, scope: "explicit-unavailable" };
    }
    currentRuntimeHome = root;
    return { root, scope: "explicit" };
  }
  const candidates = [
    { root: path.join(defaultProgramData(env), "Xuva"), scope: "machine" },
    { root: path.join(defaultLocalAppData(env), "Xuva"), scope: "user" },
  ];

  for (const candidate of candidates) {
    if (canWriteDir(candidate.root)) {
      currentRuntimeHome = candidate.root;
      return candidate;
    }
  }

  const fallback = path.join(defaultLocalAppData(env), "Xuva");
  ensureDir(fallback);
  currentRuntimeHome = fallback;
  return { root: fallback, scope: "user-fallback" };
}

function serverEnv() {
  const env = { ...process.env };
  const root = runtimeRoot();
  const runtime = resolveRuntimeHome(env);
  const xuvaRoot = runtime.root;
  env.XUVA_RUNTIME_HOME = xuvaRoot;
  env.XUVA_RUNTIME_SCOPE = runtime.scope;
  if (!env.XUVA_DATA_DIR) {
    env.XUVA_DATA_DIR = path.join(xuvaRoot, "data");
  }
  if (!env.XUVA_LOG_DIR) env.XUVA_LOG_DIR = path.join(xuvaRoot, "logs");
  if (!env.XUVA_TRANSCODE_DIR) env.XUVA_TRANSCODE_DIR = path.join(xuvaRoot, "transcode");
  if (!env.XUVA_DOWNLOADS_DIR) env.XUVA_DOWNLOADS_DIR = path.join(xuvaRoot, "downloads");
  if (!env.XUVA_METADATA_DIR) env.XUVA_METADATA_DIR = path.join(xuvaRoot, "metadata");
  if (!env.XUVA_CACHE_DIR) env.XUVA_CACHE_DIR = path.join(xuvaRoot, "cache");
  if (!env.XUVA_TEMP_DIR) env.XUVA_TEMP_DIR = path.join(xuvaRoot, "temp");
  if (!env.XUVA_TRAILERS_DIR) env.XUVA_TRAILERS_DIR = path.join(xuvaRoot, "trailers");
  if (!env.XUVA_HTTP_ADDR) {
    env.XUVA_HTTP_ADDR = "0.0.0.0:8097";
  }
  const ffmpegPath = path.join(root, "bin", process.platform === "win32" ? "ffmpeg.exe" : "ffmpeg");
  const ffprobePath = path.join(root, "bin", process.platform === "win32" ? "ffprobe.exe" : "ffprobe");
  if (!env.XUVA_FFMPEG_PATH && fs.existsSync(ffmpegPath)) {
    env.XUVA_FFMPEG_PATH = ffmpegPath;
  }
  if (!env.XUVA_FFPROBE_PATH && fs.existsSync(ffprobePath)) {
    env.XUVA_FFPROBE_PATH = ffprobePath;
  }
  for (const dir of [
    env.XUVA_DATA_DIR,
    env.XUVA_LOG_DIR,
    env.XUVA_TRANSCODE_DIR,
    env.XUVA_DOWNLOADS_DIR,
    env.XUVA_METADATA_DIR,
    env.XUVA_CACHE_DIR,
    env.XUVA_TEMP_DIR,
    env.XUVA_TRAILERS_DIR,
  ]) {
    ensureDir(dir);
  }
  return env;
}

function serverCommand() {
  const overrideCmd = String(process.env.XUVA_SERVER_CMD || "").trim();
  const overrideArgs = String(process.env.XUVA_SERVER_ARGS || "").trim();
  const cwd = String(process.env.XUVA_SERVER_CWD || "").trim() || serverCwdDefault;
  if (overrideCmd) {
    return { cmd: overrideCmd, args: overrideArgs ? overrideArgs.split(" ").filter(Boolean) : [], cwd, env: serverEnv() };
  }

  const packaged = packagedServerPath();
  if (fs.existsSync(packaged)) {
    return { cmd: packaged, args: [], cwd: runtimeRoot(), env: serverEnv() };
  }

  return { cmd: "go", args: ["run", "./cmd/Xuva"], cwd, env: process.env };
}

function startServer() {
  if (serverProcess || currentMode !== "local") return;
  const cfg = serverCommand();
  logEvent("server.start.requested", { cmd: cfg.cmd, args: cfg.args, cwd: cfg.cwd });
  serverProcess = spawn(cfg.cmd, cfg.args, {
    cwd: cfg.cwd,
    stdio: "inherit",
    windowsHide: true,
    shell: false,
    env: cfg.env,
  });
  serverProcess.on("exit", (code, signal) => {
    logEvent("server.exit", { code, signal, restarting: serverRestarting, quitting: Boolean(app.isQuiting) });
    serverProcess = null;
    if (!serverRestarting && !app.isQuiting && currentMode === "local") {
      setTimeout(() => startServer(), 1200);
    }
  });
}

function stopServer() {
  return new Promise((resolve) => {
    if (!serverProcess) return resolve();
    logEvent("server.stop.requested");
    const current = serverProcess;
    serverProcess = null;
    current.once("exit", () => resolve());
    current.kill();
  });
}

async function restartServer() {
  serverRestarting = true;
  try {
    logEvent("server.restart.requested");
    await stopServer();
    startServer();
    logEvent("server.restart.completed");
    return { ok: true };
  } finally {
    serverRestarting = false;
  }
}

function activeAppURL() {
  return currentMode === "remote" && currentRemoteUrl ? currentRemoteUrl : LOCAL_APP_URL;
}

function waitForLocalServer(timeoutMs = 30000) {
  if (currentMode !== "local") return Promise.resolve();
  const deadline = Date.now() + timeoutMs;
  const healthURL = new URL("/api/health", LOCAL_APP_URL);

  return new Promise((resolve) => {
    const attempt = () => {
      const req = http.get(healthURL, (res) => {
        res.resume();
        if (res.statusCode && res.statusCode >= 200 && res.statusCode < 500) {
          resolve();
          return;
        }
        retry();
      });
      req.on("error", retry);
      req.setTimeout(1200, () => {
        req.destroy();
        retry();
      });
    };
    const retry = () => {
      if (Date.now() >= deadline) {
        logEvent("server.wait.timeout", { url: healthURL.toString() });
        resolve();
        return;
      }
      setTimeout(attempt, 500);
    };
    attempt();
  });
}

async function loadAppWindow(win) {
  await waitForLocalServer();
  await win.loadURL(activeAppURL());
}

function appIconPath() {
  return "";
}

function createMainWindow() {
  const icon = appIconPath();
  mainWindow = new BrowserWindow({
    width: 1280,
    height: 820,
    show: false,
    autoHideMenuBar: true,
    icon: icon || undefined,
    webPreferences: {
      preload: path.join(__dirname, "preload.js"),
      contextIsolation: true,
      nodeIntegration: false,
    },
  });
  mainWindow.on("ready-to-show", () => mainWindow.show());
  mainWindow.on("close", (event) => {
    if (!app.isQuiting) {
      event.preventDefault();
      mainWindow.hide();
    }
  });
  loadAppWindow(mainWindow).catch((err) => logEvent("window.load_error", { error: String(err) }));
}

function createSetupWindow() {
  if (setupWindow) {
    setupWindow.focus();
    return;
  }
  const icon = appIconPath();
  setupWindow = new BrowserWindow({
    width: 640,
    height: 540,
    resizable: false,
    show: false,
    autoHideMenuBar: true,
    icon: icon || undefined,
    webPreferences: {
      preload: path.join(__dirname, "preload.js"),
      contextIsolation: true,
      nodeIntegration: false,
    },
  });
  setupWindow.on("ready-to-show", () => setupWindow.show());
  setupWindow.on("closed", () => {
    setupWindow = null;
  });
  void setupWindow.loadFile(path.join(__dirname, "setup.html"));
}

function showWindow() {
  if (!mainWindow) {
    createMainWindow();
    return;
  }
  if (mainWindow.isMinimized()) mainWindow.restore();
  mainWindow.show();
  mainWindow.focus();
}

function createTray() {
  const icon = appIconPath();
  if (!icon) {
    logEvent("tray.icon_missing");
    return;
  }
  tray = new Tray(icon);
  tray.setToolTip("Xuva");
  tray.on("double-click", showWindow);
  refreshTrayMenu();
}

function refreshTrayMenu() {
  if (!tray) return;
  const modeLabel = currentMode === "remote" ? `Server: ${currentRemoteUrl || "remote"}` : "Server: local";
  const menu = Menu.buildFromTemplate([
    { label: "Open Xuva", click: showWindow },
    { label: "Open in Browser", click: () => shell.openExternal(activeAppURL()) },
    { label: "Open Runtime Folder", click: () => currentRuntimeHome && shell.openPath(currentRuntimeHome) },
    { type: "separator" },
    { label: modeLabel, enabled: false },
    { label: "Change Server...", click: () => createSetupWindow() },
    { type: "separator" },
    ...(currentMode === "local" ? [{ label: "Restart Xuva Server", click: async () => { await restartServer(); } }] : []),
    { type: "separator" },
    {
      label: "Quit",
      click: async () => {
        app.isQuiting = true;
        await stopServer();
        app.quit();
      },
    },
  ]);
  tray.setContextMenu(menu);
}

function startDiscovery() {
  let Bonjour;
  try {
    Bonjour = require("bonjour-service").Bonjour;
  } catch {
    logEvent("discovery.bonjour_unavailable", { hint: "run npm install in apps/desktop to enable LAN discovery" });
    return;
  }
  try {
    const bonjour = new Bonjour();
    bonjourBrowser = bonjour.find({ type: "xuva" }, (service) => {
      const host = service.addresses?.[0] || service.host || "unknown";
      const port = service.port || 8097;
      const baseURL = `http://${host}:${port}`;
      const entry = {
        id: service.fqdn || baseURL,
        name: service.txt?.serverName || service.name || "Xuva Server",
        baseURL,
        host,
        port,
      };
      const existing = discoveredServers.findIndex((s) => s.id === entry.id);
      if (existing >= 0) discoveredServers[existing] = entry;
      else {
        discoveredServers.push(entry);
        logEvent("discovery.server_found", { name: entry.name, baseURL });
      }
      broadcastDiscovery();
    });
    bonjourBrowser.on("down", (service) => {
      const id = service.fqdn || "";
      discoveredServers = discoveredServers.filter((s) => s.id !== id);
      logEvent("discovery.server_lost", { fqdn: id });
      broadcastDiscovery();
    });
    logEvent("discovery.started");
  } catch (err) {
    logEvent("discovery.start_error", { error: String(err) });
  }
}

function broadcastDiscovery() {
  for (const win of BrowserWindow.getAllWindows()) {
    try {
      win.webContents.send("xuva:servers-updated", discoveredServers);
    } catch {
      // The window may be closing.
    }
  }
}

ipcMain.handle("xuva:pick-folder", async (_event, request = {}) => {
  logEvent("bridge.pick_folder.requested", { purpose: request.purpose || "folder" });
  const chosen = await dialog.showOpenDialog(mainWindow || undefined, {
    title: request.title || "Select folder",
    defaultPath: request.currentPath || undefined,
    properties: ["openDirectory", "dontAddToRecent"],
  });
  if (chosen.canceled || !chosen.filePaths || chosen.filePaths.length === 0) {
    return { path: "" };
  }
  return { path: chosen.filePaths[0] || "" };
});

ipcMain.handle("xuva:restart-server", async () => {
  logEvent("bridge.restart.requested");
  return restartServer();
});

ipcMain.handle("xuva:get-settings", () => ({ mode: currentMode, remoteUrl: currentRemoteUrl }));
ipcMain.handle("xuva:get-discovered-servers", () => discoveredServers);

ipcMain.handle("xuva:save-settings", async (_event, settings = {}) => {
  const mode = settings.mode === "remote" ? "remote" : "local";
  const remoteUrl = mode === "remote" ? String(settings.remoteUrl || "").trim() : "";
  writeSettings({ mode, remoteUrl });
  applySettings({ mode, remoteUrl });
  logEvent("settings.saved", { mode, remoteUrl });

  if (mode === "remote" && serverProcess) {
    await stopServer();
  } else if (mode === "local" && !serverProcess) {
    startServer();
  }

  if (setupWindow) {
    setupWindow.close();
    setupWindow = null;
  }

  if (mainWindow) {
    loadAppWindow(mainWindow).catch((err) => logEvent("window.reload_error", { error: String(err) }));
    mainWindow.show();
    mainWindow.focus();
  } else {
    createMainWindow();
  }

  refreshTrayMenu();
  return { ok: true };
});

app.whenReady().then(() => {
  const saved = readSettings();
  applySettings(saved);
  startDiscovery();

  if (saved) {
    if (currentMode === "local") startServer();
    createMainWindow();
  } else {
    const defaultSettings = { mode: "local", remoteUrl: "" };
    writeSettings(defaultSettings);
    applySettings(defaultSettings);
    startServer();
    createMainWindow();
  }

  createTray();

  app.on("activate", () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createMainWindow();
    } else {
      showWindow();
    }
  });
});

app.on("before-quit", () => {
  app.isQuiting = true;
  if (bonjourBrowser && typeof bonjourBrowser.stop === "function") {
    try {
      bonjourBrowser.stop();
    } catch {
      // Ignore discovery shutdown races.
    }
  }
});

app.on("window-all-closed", () => {
  // Keep the tray process alive until the operator explicitly quits.
});
