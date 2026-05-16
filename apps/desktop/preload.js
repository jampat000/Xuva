const { contextBridge, ipcRenderer } = require("electron");

contextBridge.exposeInMainWorld("xuvaDesktop", {
  async pickFolder(request = {}) {
    return ipcRenderer.invoke("xuva:pick-folder", request);
  },
  async restartServer() {
    return ipcRenderer.invoke("xuva:restart-server");
  },
});
