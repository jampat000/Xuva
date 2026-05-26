#!/usr/bin/env node

const tag = String(process.argv[2] || process.env.GITHUB_REF_NAME || "").trim();

if (!tag) {
  console.error("release tag is required");
  process.exit(1);
}

if (!/^v0\.0\.(?:[1-9]\d*)$/.test(tag)) {
  console.error(`invalid release tag: ${tag}`);
  console.error("early Xuva releases must use v0.0.x tags, for example v0.0.1");
  console.error("update docs/release-versioning.md and this guard only when promoting to the 1.x track");
  process.exit(1);
}

console.log(`release tag ok: ${tag}`);
