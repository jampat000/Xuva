(function (root, factory) {
  const api = factory();
  if (typeof module === "object" && module.exports) module.exports = api;
  root.VyrdenDomains = api;
})(typeof globalThis !== "undefined" ? globalThis : window, function () {
  const domains = {
    dashboard: {
      owns: ["library summary", "live sessions", "system status", "operations snapshot"],
      endpoints: ["/api/catalog/summary", "/api/sessions", "/api/system/status", "/api/scans", "/api/probes", "/api/work", "/api/downloads"],
    },
    playback: {
      owns: ["source detail", "playback readiness", "device impact", "source inspector"],
      endpoints: ["/api/playback/decision", "/api/media-sources", "/api/sessions", "/api/playback/state"],
    },
    metadata: {
      owns: ["ratings", "provider match", "review workbench"],
      endpoints: ["/api/metadata", "/api/review", "/api/artwork"],
    },
    settings: {
      owns: ["runtime folders", "metadata providers", "sync policy", "hardware unlock state"],
      endpoints: ["/api/settings", "/api/system/status", "/api/libraries"],
    },
    activity: {
      owns: ["jobs", "sessions", "scan/probe/download queues"],
      endpoints: ["/api/scans", "/api/probes", "/api/work", "/api/downloads", "/api/sessions"],
    },
  };

  return { domains };
});
