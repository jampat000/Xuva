#!/usr/bin/env node

// Release tag guard. Enforces strict semver `vMAJOR.MINOR.PATCH` so the tag,
// the buildinfo Version string, the MSI ProductVersion, and the Docker tag
// all stay in lock-step. Optional `-rc.N` / `-beta.N` / `-alpha.N` suffixes
// for pre-release tags (e.g. `v1.1.0-rc.1`) — these never become `:latest`
// (release.yml gates the `:latest` push on a stable tag).

const tag = String(process.argv[2] || process.env.GITHUB_REF_NAME || "").trim();

if (!tag) {
  console.error("release tag is required");
  process.exit(1);
}

// MAJOR.MINOR.PATCH with an optional pre-release suffix (semver).
const STABLE = /^v(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)$/;
const PRERELEASE = /^v(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)-(?:alpha|beta|rc)\.(?:0|[1-9]\d*)$/;

if (!STABLE.test(tag) && !PRERELEASE.test(tag)) {
  console.error(`invalid release tag: ${tag}`);
  console.error("release tags must be strict semver: vMAJOR.MINOR.PATCH or vMAJOR.MINOR.PATCH-(alpha|beta|rc).N");
  console.error("examples: v1.0.0  v1.2.3  v2.0.0-rc.1  v1.0.0-beta.2");
  console.error("see docs/release-versioning.md");
  process.exit(1);
}

console.log(`release tag ok: ${tag}`);
