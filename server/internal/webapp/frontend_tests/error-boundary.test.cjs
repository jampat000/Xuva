const test = require("node:test");
const assert = require("node:assert/strict");

const errors = require("../static/modules/error-boundary.js");

test("error boundary renders retry action and escapes messages", () => {
  const html = errors.renderErrorBoundary({ userMessage: "Bad <thing>", requestId: "req-2" }, {
    title: "Movies could not load",
    retryHandler: "navigate('movies')",
  });

  assert.match(html, /Movies could not load/);
  assert.match(html, /Bad &lt;thing&gt;/);
  assert.match(html, /navigate\(&#039;movies&#039;\)/);
  assert.match(html, /req-2/);
});

test("inline live refresh error uses a stable fallback label", () => {
  const html = errors.renderInlineError({ userMessage: "Network down" }, "patchLiveView('retry')");

  assert.match(html, /Could not refresh live data/);
  assert.match(html, /Network down/);
  assert.match(html, /Retry/);
});
