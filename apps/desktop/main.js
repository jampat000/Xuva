const path = require("path");
const { app, BrowserWindow, Menu, Tray, shell, dialog, ipcMain } = require("electron");
const { spawn } = require("child_process");

const APP_URL = process.env.XUVA_APP_URL || "http://127.0.0.1:8097";
const serverCwdDefault = path.resolve(__dirname, "..", "..", "server");

let mainWindow = null;
let tray = null;
let serverProcess = null;
let serverRestarting = false;

function logEvent(event, fields = {}) {
  const payload = {
    ts: new Date().toISOString(),
    component: "desktop-shell",
    event,
    ...fields,
  };
  try {
    console.log(JSON.stringify(payload));
  } catch {
    console.log(`[desktop-shell] ${event}`);
  }
}

function createWindow() {
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

  void mainWindow.loadURL(APP_URL);
}

function trayIconPath() {
  const base = path.resolve(__dirname, "..", "web", "svelte", "static", "icons", "tray");
  return path.join(base, "tray-24.png");
}

function createTray() {
  tray = new Tray(trayIconPath());
  tray.setToolTip("Xuva");
  tray.on("double-click", showWindow);
  refreshTrayMenu();
}

function refreshTrayMenu() {
  const menu = Menu.buildFromTemplate([
    { label: "Open Xuva", click: showWindow },
    { label: "Open in Browser", click: () => shell.openExternal(APP_URL) },
    { type: "separator" },
    {
      label: "Restart Xuva",
      click: async () => {
        await restartServer();
      },
    },
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

function showWindow() {
  if (!mainWindow) return;
  if (mainWindow.isMinimized()) mainWindow.restore();
  mainWindow.show();
  mainWindow.focus();
}

function serverCommand() {
  const overrideCmd = String(process.env.XUVA_SERVER_CMD || "").trim();
  const overrideArgs = String(process.env.XUVA_SERVER_ARGS || "").trim();
  const cwd = String(process.env.XUVA_SERVER_CWD || "").trim() || serverCwdDefault;

  if (overrideCmd) {
    return {
      cmd: overrideCmd,
      args: overrideArgs ? overrideArgs.split(" ").filter(Boolean) : [],
      cwd,
    };
  }

  return {
    cmd: "go",
    args: ["run", "./cmd/Xuva"],
    cwd,
  };
}

function startServer() {
  if (serverProcess) return;
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
    if (!serverRestarting && !app.isQuiting) {
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

ipcMain.handle("xuva:pick-folder", async (_event, request = {}) => {
  logEvent("bridge.pick_folder.requested", { purpose: request.purpose || "folder" });
  const chosen = await dialog.showOpenDialog(mainWindow || undefined, {
    title: request.title || "Select folder",
    defaultPath: request.currentPath || undefined,
    properties: ["openDirectory", "dontAddToRecent"],
  });
  if (chosen.canceled || !chosen.filePaths || chosen.filePaths.length === 0) {
    logEvent("bridge.pick_folder.completed", { canceled: true });
    return { path: "" };
  }
  logEvent("bridge.pick_folder.completed", { canceled: false });
  return { path: chosen.filePaths[0] || "" };
});

ipcMain.handle("xuva:restart-server", async () => {
  logEvent("bridge.restart.requested");
  return restartServer();
});

app.whenReady().then(() => {
  startServer();
  createWindow();
  createTray();
  app.on("activate", () => {
    if (BrowserWindow.getAllWindows().length === 0) createWindow();
    showWindow();
  });
});

app.on("before-quit", () => {
  app.isQuiting = true;
});

app.on("window-all-closed", () => {
  // Keep tray app running on Windows.
});
