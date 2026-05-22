const { contextBridge, ipcRenderer } = require("electron");

contextBridge.exposeInMainWorld("xuvaDesktop", {
  async pickFolder(request = {}) {
    return ipcRenderer.invoke("xuva:pick-folder", request);
  },
  async restartServer() {
    return ipcRenderer.invoke("xuva:restart-server");
  },
  async getSettings() {
    return ipcRenderer.invoke("xuva:get-settings");
  },
  async getDiscoveredServers() {
    return ipcRenderer.invoke("xuva:get-discovered-servers");
  },
  async saveSettings(settings = {}) {
    return ipcRenderer.invoke("xuva:save-settings", settings);
  },
  onServersUpdated(callback) {
    ipcRenderer.on("xuva:servers-updated", (_event, servers) => callback(servers));
  },
});
