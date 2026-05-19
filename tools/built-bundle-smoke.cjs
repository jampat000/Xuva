#!/usr/bin/env node
"use strict";

// Built-bundle smoke check.
//
// Issue #67: the CSP-blocks-artwork bug only manifested on the production-style
// build (xuva.exe serving the embedded SPA from server/internal/webapp/static-next),
// not on Vite dev. CI runs Go tests against source but never against the *built*
// bundle. This script closes that gap.
//
// v1 scope (minimum viable):
//   1. Build the Svelte frontend (publishes into server/internal/webapp/static-next).
//   2. Build xuva (Go) into a temp executable.
//   3. Boot xuva against a temp XUVA_DATA_DIR + non-default port.
//   4. Poll /api/health until 200 (up to 30s).
//   5. Curl / and assert Content-Security-Policy img-src permits https:
//      (or at minimum the known artwork hosts).
//   6. Clean up: kill process, remove temp dir + temp binary.
//
// Future iterations could add: headless browser, per-image CSP cross-check,
// service-worker validation, base-path/MIME checks.

const { spawn, spawnSync } = require("node:child_process");
const fs = require("node:fs");
const http = require("node:http");
const os = require("node:os");
const path = require("node:path");

const repoRoot = path.resolve(__dirname, "..");
const svelteAppDir = path.join(repoRoot, "apps", "web", "svelte");
const serverDir = path.join(repoRoot, "server");
const isWindows = process.platform === "win32";
const exeName = isWindows ? "xuva-smoke.exe" : "xuva-smoke";
// --skip-frontend-build is a local-dev escape hatch: on Windows the rename in
// publish-go-static.mjs fails (EPERM) when an air-watched xuva.exe holds the
// embedded static-next directory open. CI always runs the full pipeline.
const skipFrontendBuild = process.argv.includes("--skip-frontend-build");
const tempRoot = fs.mkdtempSync(path.join(os.tmpdir(), "xuva-smoke-"));
const tempDataDir = path.join(tempRoot, "data");
const tempExePath = path.join(tempRoot, exeName);
const httpHost = "127.0.0.1";
// Non-default port (default is 8097) to avoid colliding with a dev server.
const httpPort = process.env.XUVA_SMOKE_PORT ? Number(process.env.XUVA_SMOKE_PORT) : 18097;
const httpAddr = `${httpHost}:${httpPort}`;

// Known artwork hosts the production app needs to load posters/backdrops from.
// If CSP doesn't either allow https: globally OR enumerate these, image fetches
// from the SPA will be blocked at runtime.
const REQUIRED_ARTWORK_HOSTS = [
  "image.tmdb.org",
  "assets.fanart.tv",
  "m.media-amazon.com",
  "artworks.thetvdb.com",
  "upload.wikimedia.org",
  "i.ytimg.com",
];

let serverProcess = null;
let exitCode = 0;

function log(msg) {
  process.stdout.write(`[smoke] ${msg}\n`);
}

function fail(msg) {
  process.stderr.write(`[smoke] FAIL: ${msg}\n`);
  exitCode = 1;
}

function runSync(label, command, args, options) {
  log(`${label}: ${command} ${args.join(" ")}`);
  const result = spawnSync(command, args, {
    stdio: "inherit",
    shell: isWindows, // npm.cmd on Windows
    ...options,
  });
  if (result.status !== 0) {
    throw new Error(`${label} exited with status ${result.status}`);
  }
}

function buildFrontend() {
  // publish:go-static runs `npm run build` and copies output into
  // server/internal/webapp/static-next, which is what gets embedded.
  runSync("frontend build", "npm", ["run", "publish:go-static"], {
    cwd: svelteAppDir,
  });
}

function buildServer() {
  runSync("go build", "go", ["build", "-o", tempExePath, "./cmd/Xuva"], {
    cwd: serverDir,
  });
}

function startServer() {
  log(`starting xuva: ${tempExePath} (addr=${httpAddr}, dataDir=${tempDataDir})`);
  fs.mkdirSync(tempDataDir, { recursive: true });
  serverProcess = spawn(tempExePath, [], {
    cwd: tempRoot,
    env: {
      ...process.env,
      XUVA_DATA_DIR: tempDataDir,
      XUVA_HTTP_ADDR: httpAddr,
      XUVA_DISCOVERY_ENABLED: "false",
      XUVA_AUTH_DISABLED: "true",
    },
    stdio: ["ignore", "pipe", "pipe"],
  });
  serverProcess.stdout.on("data", (chunk) => {
    process.stdout.write(`[xuva] ${chunk}`);
  });
  serverProcess.stderr.on("data", (chunk) => {
    process.stderr.write(`[xuva] ${chunk}`);
  });
  serverProcess.on("exit", (code, signal) => {
    log(`xuva exited (code=${code} signal=${signal})`);
  });
}

function httpGet(pathname) {
  return new Promise((resolve, reject) => {
    const req = http.request(
      {
        host: httpHost,
        port: httpPort,
        path: pathname,
        method: "GET",
        timeout: 5000,
      },
      (res) => {
        const chunks = [];
        res.on("data", (c) => chunks.push(c));
        res.on("end", () => {
          resolve({
            status: res.statusCode,
            headers: res.headers,
            body: Buffer.concat(chunks).toString("utf8"),
          });
        });
      },
    );
    req.on("error", reject);
    req.on("timeout", () => {
      req.destroy(new Error("request timeout"));
    });
    req.end();
  });
}

async function waitForHealth(timeoutMs) {
  const deadline = Date.now() + timeoutMs;
  let lastErr = null;
  while (Date.now() < deadline) {
    try {
      const res = await httpGet("/api/health");
      if (res.status === 200) {
        log(`/api/health returned 200 after ${timeoutMs - (deadline - Date.now())}ms`);
        return res;
      }
      lastErr = new Error(`/api/health returned ${res.status}`);
    } catch (err) {
      lastErr = err;
    }
    await new Promise((r) => setTimeout(r, 250));
  }
  throw new Error(`server did not become healthy within ${timeoutMs}ms: ${lastErr ? lastErr.message : "no response"}`);
}

// Parse a CSP header value into a { directive: [tokens...] } map.
// CSP is semicolon-separated, each directive is "name token1 token2 ...".
function parseCSP(value) {
  const map = {};
  for (const rawSegment of value.split(";")) {
    const segment = rawSegment.trim();
    if (!segment) continue;
    const parts = segment.split(/\s+/);
    const name = parts.shift().toLowerCase();
    map[name] = parts;
  }
  return map;
}

function assertCSPAllowsArtwork(cspHeader) {
  if (!cspHeader) {
    fail("missing Content-Security-Policy header on /");
    return;
  }
  log(`CSP: ${cspHeader}`);
  const parsed = parseCSP(cspHeader);
  const imgSrc = parsed["img-src"] || parsed["default-src"] || [];
  if (imgSrc.length === 0) {
    fail("CSP has neither img-src nor default-src directive");
    return;
  }
  // If https: is present, all https hosts are permitted — done.
  if (imgSrc.includes("https:")) {
    log("img-src permits https: (covers all artwork hosts)");
    return;
  }
  // Otherwise, every required artwork host must be explicitly enumerated.
  const tokens = imgSrc.map((t) => t.toLowerCase());
  const missing = REQUIRED_ARTWORK_HOSTS.filter((host) => {
    const lower = host.toLowerCase();
    return !tokens.some((tok) => tok === lower || tok === `https://${lower}` || tok.endsWith(`//${lower}`));
  });
  if (missing.length > 0) {
    fail(
      `CSP img-src does not permit https: and is missing required artwork hosts: ${missing.join(", ")}. ` +
        `img-src tokens: [${imgSrc.join(", ")}]. ` +
        `This is the class of regression #67 was opened to catch.`,
    );
    return;
  }
  log(`img-src enumerates all required artwork hosts (${REQUIRED_ARTWORK_HOSTS.length} hosts)`);
}

async function runChecks() {
  const health = await waitForHealth(30_000);
  log(`health body (truncated): ${health.body.slice(0, 200)}`);

  const root = await httpGet("/");
  if (root.status !== 200) {
    fail(`GET / returned status ${root.status} (expected 200)`);
  }
  const csp = root.headers["content-security-policy"];
  assertCSPAllowsArtwork(csp);
}

function stopServer() {
  if (!serverProcess || serverProcess.exitCode !== null) return;
  log("stopping xuva");
  try {
    if (isWindows) {
      spawnSync("taskkill", ["/PID", String(serverProcess.pid), "/T", "/F"], { stdio: "ignore" });
    } else {
      serverProcess.kill("SIGTERM");
    }
  } catch (err) {
    log(`stop error (ignored): ${err.message}`);
  }
}

function cleanup() {
  stopServer();
  try {
    fs.rmSync(tempRoot, { recursive: true, force: true });
    log(`removed ${tempRoot}`);
  } catch (err) {
    log(`cleanup error (ignored): ${err.message}`);
  }
}

process.on("SIGINT", () => {
  cleanup();
  process.exit(130);
});

(async () => {
  try {
    if (skipFrontendBuild) {
      log("--skip-frontend-build set: skipping `npm run publish:go-static`");
    } else {
      buildFrontend();
    }
    buildServer();
    startServer();
    await runChecks();
  } catch (err) {
    fail(err && err.message ? err.message : String(err));
  } finally {
    cleanup();
  }
  if (exitCode === 0) {
    log("OK — built bundle smoke check passed");
  }
  process.exit(exitCode);
})();
