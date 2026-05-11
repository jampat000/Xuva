(function (root, factory) {
  const api = factory();
  if (typeof module === "object" && module.exports) module.exports = api;
  root.VyrdenApi = api;
})(typeof globalThis !== "undefined" ? globalThis : window, function () {
  const AUTH_TOKEN_KEY = "vyrden-auth-token";
  const AUTH_TOKEN_WINDOW_NAME_KEY = "vyrdenAuthToken";
  let memoryAuthToken = "";

  function readStorage(storage) {
    try {
      if (!storage) return "";
      return String(storage.getItem(AUTH_TOKEN_KEY) || "").trim();
    } catch {
      return "";
    }
  }

  function writeStorage(storage, value) {
    try {
      if (!storage) return;
      if (!value) storage.removeItem(AUTH_TOKEN_KEY);
      else storage.setItem(AUTH_TOKEN_KEY, value);
    } catch {
      // ignore storage failures
    }
  }

  function readWindowNameToken() {
    try {
      if (typeof window === "undefined") return "";
      const raw = String(window.name || "").trim();
      if (!raw) return "";
      const parts = raw.split(";");
      const pair = parts.find(item => item.startsWith(`${AUTH_TOKEN_WINDOW_NAME_KEY}=`));
      if (!pair) return "";
      return decodeURIComponent(pair.slice(`${AUTH_TOKEN_WINDOW_NAME_KEY}=`.length)).trim();
    } catch {
      return "";
    }
  }

  function writeWindowNameToken(value) {
    try {
      if (typeof window === "undefined") return;
      const raw = String(window.name || "");
      const parts = raw
        .split(";")
        .map(item => item.trim())
        .filter(Boolean)
        .filter(item => !item.startsWith(`${AUTH_TOKEN_WINDOW_NAME_KEY}=`));
      if (value) parts.push(`${AUTH_TOKEN_WINDOW_NAME_KEY}=${encodeURIComponent(value)}`);
      window.name = parts.join(";");
    } catch {
      // ignore window.name failures
    }
  }

  function readAuthToken() {
    if (memoryAuthToken) return memoryAuthToken;
    const localValue = typeof localStorage !== "undefined" ? readStorage(localStorage) : "";
    if (localValue) {
      memoryAuthToken = localValue;
      return localValue;
    }
    const sessionValue = typeof sessionStorage !== "undefined" ? readStorage(sessionStorage) : "";
    if (sessionValue) {
      memoryAuthToken = sessionValue;
      return sessionValue;
    }
    const windowValue = readWindowNameToken();
    if (windowValue) memoryAuthToken = windowValue;
    return windowValue;
  }

  function writeAuthToken(token) {
    const value = String(token || "").trim();
    memoryAuthToken = value;
    if (typeof localStorage !== "undefined") writeStorage(localStorage, value);
    if (typeof sessionStorage !== "undefined") writeStorage(sessionStorage, value);
    writeWindowNameToken(value);
  }

  class ApiError extends Error {
    constructor(message, options = {}) {
      super(message || "Request failed");
      this.name = "ApiError";
      this.status = options.status || 0;
      this.path = options.path || "";
      this.requestId = options.requestId || "";
      this.payload = options.payload || null;
      this.userMessage = options.userMessage || normalizeErrorMessage(this.status, message);
    }
  }

  function normalizeErrorMessage(status, message = "") {
    if (status === 0) return "Vyrden could not reach the local server. Check that the server is still running, then try again.";
    if (status === 401) return "Your session is no longer active. Sign in again to continue.";
    if (status === 403) return "This account cannot perform that action. Use an admin account for server changes.";
    if (status === 404) return "Vyrden could not find that item. It may have moved, been deleted, or not finished scanning.";
    if (status === 409) return "That action conflicts with the current server state. Refresh and try again.";
    if (status >= 500) return "The server hit a problem while handling this action. Retry once, then check Activity if it keeps failing.";
    return message || "Something went wrong. Try again.";
  }

  function parsePayload(text, path) {
    if (!text) return {};
    try {
      return JSON.parse(text);
    } catch (error) {
      throw new ApiError("Server returned unreadable data.", { path, userMessage: "Vyrden received a response it could not read. Refresh and try again." });
    }
  }

  function withTimeout(signal, timeoutMs) {
    if (!timeoutMs || typeof AbortController === "undefined") return { signal, cancel: () => {} };
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), timeoutMs);
    if (signal) {
      if (signal.aborted) controller.abort();
      signal.addEventListener("abort", () => controller.abort(), { once: true });
    }
    return { signal: controller.signal, cancel: () => clearTimeout(timeout) };
  }

  function shouldRetry(error, attempt, retries) {
    if (attempt >= retries) return false;
    return error.status === 0 || error.status === 408 || error.status === 429 || error.status >= 500;
  }

  function createApiClient(fetchImpl = globalThis.fetch, defaults = {}) {
    if (!fetchImpl) throw new Error("fetch is required");
    const timeoutMs = defaults.timeoutMs || 15000;
    const retries = defaults.retries || 0;

    async function request(path, options = {}) {
      const retryCount = Number.isFinite(options.retries) ? options.retries : retries;
      let attempt = 0;
      for (;;) {
        try {
          return await requestOnce(path, options, timeoutMs, fetchImpl);
        } catch (error) {
          const apiError = error instanceof ApiError ? error : new ApiError(error.message, { path, status: 0 });
          if (!shouldRetry(apiError, attempt, retryCount)) throw apiError;
          attempt += 1;
          await new Promise(resolve => setTimeout(resolve, 120 * attempt));
        }
      }
    }

    function send(path, body, method = "POST", options = {}) {
      return request(path, {
        ...options,
        method,
        headers: { "Content-Type": "application/json", ...(options.headers || {}) },
        body: JSON.stringify(body || {}),
      });
    }

    return { request, send };
  }

  function cookieValue(name) {
    if (typeof document === "undefined" || !document.cookie) return "";
    const escaped = name.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
    const match = document.cookie.match(new RegExp(`(?:^|; )${escaped}=([^;]*)`));
    return match ? decodeURIComponent(match[1]) : "";
  }

  function withCSRFHeaders(options = {}) {
    const method = String(options.method || "GET").toUpperCase();
    const headers = { ...(options.headers || {}) };
    if (!headers["X-Auth-Token"]) {
      const authToken = readAuthToken();
      if (authToken) headers["X-Auth-Token"] = authToken;
    }
    if (method === "GET" || method === "HEAD" || method === "OPTIONS") return { ...options, headers };
    if (!headers["X-CSRF-Token"]) {
      const token = cookieValue("vyrden_csrf");
      if (token) headers["X-CSRF-Token"] = token;
    }
    return { ...options, headers };
  }

  async function requestOnce(path, options, timeoutMs, fetchImpl) {
    const securedOptions = withCSRFHeaders(options);
    const { signal, cancel } = withTimeout(securedOptions.signal, securedOptions.timeoutMs || timeoutMs);
    try {
      const response = await fetchImpl(path, { credentials: "include", ...securedOptions, signal });
      const rotatedToken = response.headers?.get?.("X-Auth-Token") || "";
      if (rotatedToken) writeAuthToken(rotatedToken);
      const text = await response.text();
      const payload = parsePayload(text, path);
      if (!response.ok) {
        throw new ApiError(payload.error || response.statusText, {
          status: response.status,
          path,
          payload,
          requestId: response.headers?.get?.("X-Request-ID") || payload.requestId || "",
        });
      }
      return payload;
    } catch (error) {
      if (error instanceof ApiError) throw error;
      const aborted = error.name === "AbortError";
      throw new ApiError(aborted ? "Request timed out." : error.message, {
        status: 0,
        path,
        userMessage: aborted
          ? "The server took too long to answer. Retry the action or check Activity for a busy job."
          : normalizeErrorMessage(0, error.message),
      });
    } finally {
      cancel();
    }
  }

  function validateArrayPayload(payload, key) {
    if (!payload || !Array.isArray(payload[key])) {
      throw new ApiError(`Expected ${key} list from server.`, { userMessage: "The server response shape changed. Refresh the app; if it continues, check the latest update." });
    }
    return payload[key];
  }

  return { ApiError, createApiClient, normalizeErrorMessage, validateArrayPayload, readAuthToken, writeAuthToken, AUTH_TOKEN_KEY };
});
