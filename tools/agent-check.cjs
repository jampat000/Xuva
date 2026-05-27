#!/usr/bin/env node
"use strict";

const fs = require("node:fs");
const path = require("node:path");

const root = path.resolve(__dirname, "..");

const requiredFiles = [
  "AGENTS.md",
  "docs/index.md",
  "docs/agent-harness.md",
  "docs/quality-score.md",
  "docs/tech-debt-tracker.md",
  "docs/plans/README.md",
  "docs/plans/active/apple-tv-alpha.md",
  "docs/plans/active/desktop-alpha.md",
  "docs/apple-tv-alpha.md",
  "docs/alpha-desktop.md",
];

const agentLinkTargets = [
  "docs/index.md",
  "docs/architecture.md",
  "docs/product-principles.md",
  "docs/plans/active/",
  "docs/quality-score.md",
];

function read(relativePath) {
  return fs.readFileSync(path.join(root, relativePath), "utf8");
}

function exists(relativePath) {
  return fs.existsSync(path.join(root, relativePath));
}

function checkRequiredFiles(errors) {
  for (const item of requiredFiles) {
    if (!exists(item)) errors.push(`missing required harness file: ${item}`);
  }
}

function checkAgentsMap(errors) {
  const text = exists("AGENTS.md") ? read("AGENTS.md") : "";
  const lines = text.split(/\r?\n/).filter(line => line.trim());
  if (lines.length > 140) errors.push("AGENTS.md should stay short; move detailed guidance into docs/");
  for (const target of agentLinkTargets) {
    if (!text.includes(target)) errors.push(`AGENTS.md does not point to ${target}`);
  }
}

function matchAll(text, regex) {
  return Array.from(text.matchAll(regex), item => item[1]);
}

function checkRoutePolicy(errors) {
  if (!exists("server/internal/api/router.go") || !exists("server/internal/api/authz.go") || !exists("docs/route-policy.md")) return;
  const router = read("server/internal/api/router.go");
  const authz = read("server/internal/api/authz.go");
  const docs = read("docs/route-policy.md");
  const protectedRoutes = new Set(matchAll(router, /handleProtected(?:CSRF)?\(mux,\s*deps,\s*"([^"]+)"/g));
  const policyRoutes = new Set(matchAll(authz, /"([A-Z]+ \/api\/[^"]+|GET \/play\/\{id\})"\s*:\s*route/g));
  for (const route of [...protectedRoutes].sort()) {
    if (!policyRoutes.has(route)) errors.push(`protected route missing authz policy: ${route}`);
    if (!docs.includes(`\`${route}\``)) errors.push(`protected route missing docs/route-policy.md entry: ${route}`);
  }
  for (const route of [...policyRoutes].sort()) {
    if (!protectedRoutes.has(route)) errors.push(`authz policy has no protected route registration: ${route}`);
  }
}

function checkPlanStructure(errors) {
  if (!exists("docs/plans/active")) errors.push("missing docs/plans/active");
  if (!exists("docs/plans/completed")) errors.push("missing docs/plans/completed");
  const activeDir = path.join(root, "docs/plans/active");
  if (fs.existsSync(activeDir) && !fs.readdirSync(activeDir).some(name => name.endsWith(".md"))) {
    errors.push("docs/plans/active has no active plan markdown files");
  }
}

function checkGithubActionRuntimePins(errors) {
  const deprecatedActionPins = [
    "actions/checkout@v4",
    "actions/setup-go@v5",
    "actions/setup-node@v4",
    "actions/upload-artifact@v4",
    "actions/download-artifact@v4",
    "docker/setup-buildx-action@v3",
    "docker/login-action@v3",
    "docker/build-push-action@v6",
    "softprops/action-gh-release@v2",
  ];
  const workflowsDir = path.join(root, ".github", "workflows");
  if (!fs.existsSync(workflowsDir)) return;
  for (const file of fs.readdirSync(workflowsDir).filter(name => /\.ya?ml$/i.test(name))) {
    const relativePath = path.join(".github", "workflows", file);
    const text = read(relativePath);
    for (const pin of deprecatedActionPins) {
      if (text.includes(pin)) {
        errors.push(`${relativePath} uses deprecated Node 20 action pin: ${pin}`);
      }
    }
    if (/runs-on:\s*windows-latest\b/.test(text)) {
      errors.push(`${relativePath} uses implicit windows-latest; pin a concrete Windows runner image`);
    }
  }
}

const errors = [];
checkRequiredFiles(errors);
checkAgentsMap(errors);
checkPlanStructure(errors);
checkRoutePolicy(errors);
checkGithubActionRuntimePins(errors);

if (errors.length) {
  for (const error of errors) console.error(`agent-check: ${error}`);
  process.exit(1);
}

console.log("agent-check: ok");
