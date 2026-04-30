(function (root, factory) {
  const api = factory();
  if (typeof module === "object" && module.exports) module.exports = api;
  root.VyrdenApi = api;
})(typeof globalThis !== "undefined" ? globalThis : window, function () {
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
    if (status === 401 || status === 403) return "This action needs permission. Sign in again or use an account with access.";
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

  async function requestOnce(path, options, timeoutMs, fetchImpl) {
    const { signal, cancel } = withTimeout(options.signal, options.timeoutMs || timeoutMs);
    try {
      const response = await fetchImpl(path, { ...options, signal });
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

  return { ApiError, createApiClient, normalizeErrorMessage, validateArrayPayload };
});
