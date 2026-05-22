const path = require("path");
const fs = require("fs");
const { app, BrowserWindow, Menu, Tray, shell, dialog, ipcMain } = require("electron");
const { spawn } = require("child_process");

const LOCAL_APP_URL = "http://127.0.0.1:8097";
const SETTINGS_FILE = () => path.join(app.getPath("userData"), "xuva-settings.json");
const serverCwdDefault = path.resolve(__dirname, "..", "..", "server");

let mainWindow = null;
let setupWindow = null;
let tray = null;
let serverProcess = null;
let serverRestarting = false;
let discoveredServers = [];
let bonjourBrowser = null;
let currentMode = "local";  // "local" | "remote"
let currentRemoteUrl = "";

// ── Settings persistence ───────────────────────────────────────────────────

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

// ── Logging ────────────────────────────────────────────────────────────────

function logEvent(event, fields = {}) {
  const payload = { ts: new Date().toISOString(), component: "desktop-shell", event, ...fields };
  try {
    console.log(JSON.stringify(payload));
  } catch {
    console.log(`[desktop-shell] ${event}`);
  }
}

// ── Bonjour discovery ──────────────────────────────────────────────────────

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
      if (existing >= 0) {
        discoveredServers[existing] = entry;
      } else {
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
  const all = BrowserWindow.getAllWindows();
  for (const win of all) {
    try {
      win.webContents.send("xuva:servers-updated", discoveredServers);
    } catch {
      // window may be closing
    }
  }
}

// ── Local server process ───────────────────────────────────────────────────

function serverCommand() {
  const overrideCmd = String(process.env.XUVA_SERVER_CMD || "").trim();
  const overrideArgs = String(process.env.XUVA_SERVER_ARGS || "").trim();
  const cwd = String(process.env.XUVA_SERVER_CWD || "").trim() || serverCwdDefault;
  if (overrideCmd) {
    return { cmd: overrideCmd, args: overrideArgs ? overrideArgs.split(" ").filter(Boolean) : [], cwd };
  }
  return { cmd: "go", args: ["run", "./cmd/Xuva"], cwd };
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
    env: process.env,
  });
  serverProcess.on("exit", () => {
    logEvent("server.exit", { restarting: serverRestarting, quitting: Boolean(app.isQuiting) });
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

// ── Windows ────────────────────────────────────────────────────────────────

function activeAppURL() {
  return currentMode === "remote" && currentRemoteUrl ? currentRemoteUrl : LOCAL_APP_URL;
}

function createMainWindow() {
  mainWindow = new BrowserWindow({
    width: 1280,
    height: 820,
    show: false,
    autoHideMenuBar: true,
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
  void mainWindow.loadURL(activeAppURL());
}

function createSetupWindow() {
  if (setupWindow) {
    setupWindow.focus();
    return;
  }
  setupWindow = new BrowserWindow({
    width: 640,
    height: 540,
    resizable: false,
    show: false,
    autoHideMenuBar: true,
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

function trayIconPath() {
  const base = path.resolve(__dirname, "..", "web", "svelte", "static", "icons", "tray");
  return path.join(base, "tray-24.png");
}

function showWindow() {
  if (!mainWindow) return;
  if (mainWindow.isMinimized()) mainWindow.restore();
  mainWindow.show();
  mainWindow.focus();
}

function createTray() {
  tray = new Tray(trayIconPath());
  tray.setToolTip("Xuva");
  tray.on("double-click", showWindow);
  refreshTrayMenu();
}

function refreshTrayMenu() {
  const modeLabel = currentMode === "remote"
    ? `Server: ${currentRemoteUrl || "remote"}`
    : "Server: local";
  const menu = Menu.buildFromTemplate([
    { label: "Open Xuva", click: showWindow },
    { label: "Open in Browser", click: () => shell.openExternal(activeAppURL()) },
    { type: "separator" },
    { label: modeLabel, enabled: false },
    { label: "Change Server…", click: () => createSetupWindow() },
    { type: "separator" },
    ...(currentMode === "local"
      ? [{ label: "Restart Xuva Server", click: async () => { await restartServer(); } }]
      : []),
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

// ── IPC handlers ───────────────────────────────────────────────────────────

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

ipcMain.handle("xuva:get-settings", () => {
  return { mode: currentMode, remoteUrl: currentRemoteUrl };
});

ipcMain.handle("xuva:get-discovered-servers", () => {
  return discoveredServers;
});

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

  // Close setup window and reload / open main window
  if (setupWindow) {
    setupWindow.close();
    setupWindow = null;
  }

  if (mainWindow) {
    void mainWindow.loadURL(activeAppURL());
    mainWindow.show();
    mainWindow.focus();
  } else {
    createMainWindow();
  }

  refreshTrayMenu();
  return { ok: true };
});

// ── App lifecycle ──────────────────────────────────────────────────────────

app.whenReady().then(() => {
  const saved = readSettings();
  applySettings(saved);

  startDiscovery();

  if (saved) {
    // Settings already exist — start normally.
    if (currentMode === "local") startServer();
    createMainWindow();
  } else {
    // First launch — show the setup / server picker.
    createSetupWindow();
  }

  createTray();

  app.on("activate", () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      if (readSettings()) createMainWindow();
      else createSetupWindow();
    } else {
      showWindow();
    }
  });
});

app.on("before-quit", () => {
  app.isQuiting = true;
});

app.on("window-all-closed", () => {
  // Keep tray running.
});
