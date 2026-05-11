const test = require("node:test");
const assert = require("node:assert/strict");

const { ApiError, createApiClient, validateArrayPayload, readAuthToken, writeAuthToken } = require("./modules/api-client.js");

function response(status, body, headers = {}) {
  return {
    ok: status >= 200 && status < 300,
    status,
    statusText: status === 200 ? "OK" : "Failed",
    headers: { get: name => headers[name] || headers[name.toLowerCase()] || "" },
    text: async () => body,
  };
}

test("api client returns parsed JSON for successful responses", async () => {
  const client = createApiClient(async () => response(200, JSON.stringify({ ok: true })));
  assert.deepEqual(await client.request("/api/health"), { ok: true });
});

test("api client normalizes server errors with request id", async () => {
  const client = createApiClient(async () => response(500, JSON.stringify({ error: "database unavailable" }), { "X-Request-ID": "req-1" }));

  await assert.rejects(
    () => client.request("/api/catalog/summary"),
    error => {
      assert.equal(error instanceof ApiError, true);
      assert.equal(error.status, 500);
      assert.equal(error.requestId, "req-1");
      assert.match(error.userMessage, /server hit a problem/i);
      return true;
    },
  );
});

test("api client retries transient failures", async () => {
  let calls = 0;
  const client = createApiClient(async () => {
    calls += 1;
    return calls === 1 ? response(503, JSON.stringify({ error: "busy" })) : response(200, JSON.stringify({ status: "ok" }));
  }, { retries: 1 });

  assert.deepEqual(await client.request("/api/ready"), { status: "ok" });
  assert.equal(calls, 2);
});

test("array payload validator catches contract drift", () => {
  assert.deepEqual(validateArrayPayload({ sessions: [] }, "sessions"), []);
  assert.throws(() => validateArrayPayload({ sessions: {} }, "sessions"), ApiError);
});

test("auth token falls back when localStorage is unavailable", () => {
  const originalWindow = global.window;
  const originalLocalStorage = global.localStorage;
  const originalSessionStorage = global.sessionStorage;
  try {
    const sessionStore = new Map();
    global.window = { name: "" };
    global.localStorage = {
      getItem() {
        throw new Error("localStorage blocked");
      },
      setItem() {
        throw new Error("localStorage blocked");
      },
      removeItem() {
        throw new Error("localStorage blocked");
      },
    };
    global.sessionStorage = {
      getItem(key) {
        return sessionStore.get(key) || "";
      },
      setItem(key, value) {
        sessionStore.set(key, String(value));
      },
      removeItem(key) {
        sessionStore.delete(key);
      },
    };

    writeAuthToken("token-123");
    assert.equal(sessionStore.get("vyrden-auth-token"), "token-123");
    assert.match(global.window.name, /vyrdenAuthToken=token-123/);
    assert.equal(readAuthToken(), "token-123");
  } finally {
    global.window = originalWindow;
    global.localStorage = originalLocalStorage;
    global.sessionStorage = originalSessionStorage;
    writeAuthToken("");
  }
});
