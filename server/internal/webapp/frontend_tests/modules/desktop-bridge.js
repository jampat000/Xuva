(function (root, factory) {
  const api = factory(root);
  if (typeof module === "object" && module.exports) module.exports = api;
  root.XuvaDesktopBridge = api;
})(typeof globalThis !== "undefined" ? globalThis : window, function (root) {
  function desktopBridge() {
    return root.xuvaDesktop || root.XuvaDesktop || null;
  }

  function capabilities() {
    const bridge = desktopBridge();
    return {
      available: Boolean(bridge),
      canPickFolder: Boolean(bridge && typeof bridge.pickFolder === "function"),
      canRestart: Boolean(bridge && typeof bridge.restartServer === "function"),
    };
  }

  async function restartServer() {
    const bridge = desktopBridge();
    if (!bridge || typeof bridge.restartServer !== "function") {
      return { supported: false };
    }
    await bridge.restartServer();
    return { supported: true };
  }

  return { desktopBridge, capabilities, restartServer };
});
