(function (root, factory) {
  const api = factory();
  if (typeof module === "object" && module.exports) module.exports = api;
  root.VyrdenErrors = api;
})(typeof globalThis !== "undefined" ? globalThis : window, function () {
  function escapeHTML(value) {
    return String(value ?? "").replace(/[&<>"']/g, char => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "\"": "&quot;", "'": "&#039;" }[char]));
  }

  function userMessage(error) {
    return error?.userMessage || error?.message || "Something went wrong. Try again.";
  }

  function renderErrorBoundary(error, action = {}) {
    const title = action.title || "Something needs attention";
    const retry = action.retryLabel || "Try again";
    const detail = userMessage(error);
    return `<div class="error-boundary card" role="alert">
      <div>
        <p class="eyebrow">Action could not complete</p>
        <h2>${escapeHTML(title)}</h2>
        <p>${escapeHTML(detail)}</p>
        ${error?.requestId ? `<small>Request ${escapeHTML(error.requestId)}</small>` : ""}
      </div>
      ${action.retryHandler ? `<button class="primary" type="button" onclick="${escapeHTML(action.retryHandler)}">${escapeHTML(retry)}</button>` : ""}
    </div>`;
  }

  function renderInlineError(error, retryHandler = "") {
    return `<div class="inline-error" role="status">
      <strong>Could not refresh live data</strong>
      <span>${escapeHTML(userMessage(error))}</span>
      ${retryHandler ? `<button type="button" onclick="${escapeHTML(retryHandler)}">Retry</button>` : ""}
    </div>`;
  }

  return { renderErrorBoundary, renderInlineError, userMessage };
});
