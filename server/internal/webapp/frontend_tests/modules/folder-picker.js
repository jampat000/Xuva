(function (root, factory) {
  const api = factory(root);
  if (typeof module === "object" && module.exports) module.exports = api;
  root.XuvaFolderPicker = api;
})(typeof globalThis !== "undefined" ? globalThis : window, function (root) {
  function desktopBridge() {
    const bridge = root.xuvaDesktop || root.XuvaDesktop;
    if (!bridge || typeof bridge.pickFolder !== "function") return null;
    return bridge;
  }

  function normalizePickedPath(result) {
    if (!result) return "";
    if (typeof result === "string") return result.trim();
    if (typeof result.path === "string") return result.path.trim();
    return "";
  }

  async function pickFolder(options = {}) {
    const bridge = desktopBridge();
    if (!bridge) return { supported: false, path: "" };
    const result = await bridge.pickFolder({
      title: options.title || "Choose folder",
      currentPath: options.currentPath || "",
      purpose: options.purpose || "folder",
    });
    return { supported: true, path: normalizePickedPath(result) };
  }

  return { desktopBridge, normalizePickedPath, pickFolder };
});
