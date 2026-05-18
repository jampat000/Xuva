#!/usr/bin/env node
/**
 * Xuva Library Audit Script
 * ─────────────────────────
 * Walks every media source in your library and asks the server "how would
 * you play this file?" — without streaming anything. Outputs a JSON data
 * file and a human-readable text report.
 *
 * Usage:
 *   node audit-library.mjs [options]
 *
 * Options:
 *   --url          Server base URL (default: http://127.0.0.1:8097)
 *   --username     Username — prompted interactively if omitted
 *   --password     Password — prompted interactively if omitted
 *   --token        Skip login entirely, use this auth token directly
 *   --concurrency  Parallel requests (default: 4, max: 10)
 *   --out          Output directory (default: ./audit-results)
 *   --limit        Max files to audit, useful for a quick spot-check (default: all)
 *   --client-profile  Client profile string (default: web)
 *
 * Examples:
 *   node audit-library.mjs
 *   node audit-library.mjs --username admin --password mysecret
 *   node audit-library.mjs --token abc123 --concurrency 6
 *   node audit-library.mjs --limit 50
 */

import { writeFileSync, mkdirSync } from 'fs';
import { join, resolve } from 'path';
import { performance } from 'perf_hooks';
import { createInterface } from 'readline';

// ─── CLI arg parsing ─────────────────────────────────────────────────────────
const args = parseArgs(process.argv.slice(2));
const BASE_URL    = args['url']            ?? 'http://127.0.0.1:8097';
const TOKEN_ARG   = args['token']          ?? '';
const CONCURRENCY = Math.min(10, Math.max(1, parseInt(args['concurrency'] ?? '4', 10)));
const OUT_DIR     = resolve(args['out']    ?? './audit-results');
const LIMIT       = args['limit']          ? parseInt(args['limit'], 10) : Infinity;
const CLIENT_PROF = args['client-profile'] ?? 'web';

function parseArgs(argv) {
  const result = {};
  for (let i = 0; i < argv.length; i++) {
    const key = argv[i];
    if (key.startsWith('--')) {
      result[key.slice(2)] = argv[i + 1] ?? true;
      i++;
    }
  }
  return result;
}

// ─── Interactive credential prompts ──────────────────────────────────────────
function prompt(question) {
  const rl = createInterface({ input: process.stdin, output: process.stdout });
  return new Promise(resolve => rl.question(question, answer => { rl.close(); resolve(answer); }));
}

function promptPassword(question) {
  return new Promise(resolve => {
    const rl = createInterface({ input: process.stdin, output: process.stdout });
    // Hide input on terminals that support it
    if (process.stdout.isTTY) {
      process.stdout.write(question);
      process.stdin.setRawMode(true);
      let password = '';
      process.stdin.resume();
      process.stdin.setEncoding('utf8');
      process.stdin.on('data', function handler(ch) {
        if (ch === '\n' || ch === '\r' || ch === '') {
          process.stdin.setRawMode(false);
          process.stdin.pause();
          process.stdin.removeListener('data', handler);
          process.stdout.write('\n');
          rl.close();
          resolve(password);
        } else if (ch === '') {
          process.stdout.write('\n');
          process.exit(0);
        } else if (ch === '') {
          password = password.slice(0, -1);
        } else {
          password += ch;
          process.stdout.write('*');
        }
      });
    } else {
      // Non-TTY (piped input) — just read normally
      rl.question(question, answer => { rl.close(); resolve(answer); });
    }
  });
}

// ─── Helpers ──────────────────────────────────────────────────────────────────
function fmtBytes(b) {
  if (!b) return '—';
  if (b >= 1_073_741_824) return `${(b / 1_073_741_824).toFixed(2)} GB`;
  if (b >= 1_048_576)     return `${(b / 1_048_576).toFixed(1)} MB`;
  return `${Math.round(b / 1024)} KB`;
}
function fmtBitrate(bps) {
  if (!bps) return '—';
  if (bps >= 1_000_000) return `${(bps / 1_000_000).toFixed(1)} Mbps`;
  return `${Math.round(bps / 1000)} kbps`;
}
function fmtMs(ms) {
  if (ms < 1000) return `${Math.round(ms)}ms`;
  return `${(ms / 1000).toFixed(1)}s`;
}
function pad(str, len) {
  return String(str).padEnd(len).slice(0, len);
}
function progress(done, total, label) {
  const pct = total > 0 ? Math.round((done / total) * 100) : 0;
  const bar  = '█'.repeat(Math.floor(pct / 5)) + '░'.repeat(20 - Math.floor(pct / 5));
  process.stdout.write(`\r  [${bar}] ${pct}%  ${done}/${total}  ${label}`.padEnd(80));
}

// ─── HTTP helpers ─────────────────────────────────────────────────────────────
async function apiGet(path, token) {
  const res = await fetch(`${BASE_URL}${path}`, {
    headers: { 'X-Auth-Token': token, 'Accept': 'application/json' }
  });
  if (!res.ok) throw new Error(`GET ${path} → HTTP ${res.status}`);
  return res.json();
}

async function apiPost(path, body, token) {
  const res = await fetch(`${BASE_URL}${path}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { 'X-Auth-Token': token } : {}),
    },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const text = await res.text().catch(() => '');
    throw new Error(`POST ${path} → HTTP ${res.status}: ${text.slice(0, 120)}`);
  }
  return res.json();
}

// ─── Concurrency pool ─────────────────────────────────────────────────────────
async function pool(tasks, concurrency, fn) {
  const results = [];
  let i = 0;
  async function worker() {
    while (i < tasks.length) {
      const idx = i++;
      results[idx] = await fn(tasks[idx], idx);
    }
  }
  const workers = Array.from({ length: Math.min(concurrency, tasks.length) }, worker);
  await Promise.all(workers);
  return results;
}

// ─── Main ─────────────────────────────────────────────────────────────────────
async function main() {
  const startTime = performance.now();
  mkdirSync(OUT_DIR, { recursive: true });

  console.log('\n  ╔══════════════════════════════════════════╗');
  console.log('  ║       Xuva Library Audit Script          ║');
  console.log('  ╚══════════════════════════════════════════╝\n');
  console.log(`  Server     : ${BASE_URL}`);
  console.log(`  Profile    : ${CLIENT_PROF}`);
  console.log(`  Concurrency: ${CONCURRENCY}`);
  console.log(`  Output     : ${OUT_DIR}\n`);

  // ── Step 1: Authenticate ────────────────────────────────────────────────────
  let token = TOKEN_ARG;
  if (!token) {
    // First check if auth is even enabled
    let authDisabledOnServer = false;
    try {
      const probe = await apiPost('/api/auth/login', { username: '', password: '' }).catch(() => null);
      if (probe?.authDisabled) authDisabledOnServer = true;
    } catch { /* ignore */ }

    if (authDisabledOnServer) {
      token = '__auth_disabled__';
      console.log('  Auth is disabled on this server — no credentials needed.\n');
    } else {
      // Prompt for credentials interactively
      const usernameArg = args['username'] ?? '';
      const passwordArg = args['password'] ?? '';

      const username = usernameArg || await prompt('  Username: ');
      const password = passwordArg || await promptPassword('  Password: ');

      process.stdout.write('\n  Authenticating…');
      try {
        const resp = await apiPost('/api/auth/login', {
          username: username.trim(),
          password: password.trim(),
        });
        if (resp.authDisabled) {
          token = '__auth_disabled__';
          console.log(' auth disabled, proceeding without token');
        } else {
          token = resp.sessionToken;
          if (!token) throw new Error('No sessionToken in login response');
          console.log(` ✓ signed in as ${resp.user?.username ?? username}`);
        }
      } catch (e) {
        console.error(`\n  ✗ Login failed: ${e.message}`);
        process.exit(1);
      }
    }
  } else {
    console.log('  Using supplied --token');
  }
  const authHeader = token === '__auth_disabled__' ? '' : token;

  // ── Step 2: Fetch all media sources ─────────────────────────────────────────
  process.stdout.write('\n  Loading media sources…');
  let allSources = [];
  try {
    const resp = await apiGet(`/api/media-sources?limit=5000`, authHeader);
    allSources = resp.mediaSources ?? [];
    console.log(` ✓ ${allSources.length} files found`);
  } catch (e) {
    console.error(`\n  ✗ Failed to load media sources: ${e.message}`);
    process.exit(1);
  }

  if (allSources.length === 0) {
    console.log('\n  Library is empty. Nothing to audit.');
    process.exit(0);
  }

  const sources = allSources.slice(0, LIMIT);
  if (sources.length < allSources.length) {
    console.log(`  (limited to first ${sources.length} of ${allSources.length})`);
  }

  // ── Step 3: Probe playback route for each file ───────────────────────────────
  console.log(`\n  Auditing ${sources.length} files (concurrency: ${CONCURRENCY})…\n`);

  let done = 0;
  const results = await pool(sources, CONCURRENCY, async (src) => {
    const id = src.id ?? '';
    const t0 = performance.now();
    let route = null;
    let routeError = null;

    try {
      const params = new URLSearchParams({
        mediaSourceId: id,
        clientProfile: CLIENT_PROF,
        supportsAdaptive: 'true',
      });
      route = await apiGet(`/api/playback/route?${params}`, authHeader);
    } catch (e) {
      routeError = e.message;
    }

    done++;
    const name = (src.name ?? src.relPath ?? id).slice(-45);
    progress(done, sources.length, name);

    return {
      id,
      name: src.name ?? '',
      relPath: src.relPath ?? src.path ?? '',
      container: src.container ?? '',
      videoCodec: src.videoCodec ?? '',
      width: src.width ?? 0,
      height: src.height ?? 0,
      bitrate: src.bitrate ?? 0,
      sizeBytes: src.sizeBytes ?? 0,
      audioStreams: src.audioStreams ?? 0,
      subtitleStreams: src.subtitleStreams ?? 0,
      durationSeconds: src.durationSeconds ?? 0,
      // Decision fields
      mode: route?.decision?.mode ?? (routeError ? 'error' : 'unknown'),
      containerAction: route?.decision?.containerAction ?? '',
      videoAction: route?.decision?.videoAction ?? '',
      audioAction: route?.decision?.audioAction ?? '',
      subtitleAction: route?.decision?.subtitleAction ?? '',
      reasonCode: route?.decision?.reasonCode ?? '',
      reasonText: route?.decision?.reasonText ?? '',
      estimatedCpuCost: route?.decision?.estimatedCpuCost ?? '',
      estimatedNetworkBitrate: route?.decision?.estimatedNetworkBitrate ?? 0,
      suggestedFixes: route?.decision?.suggestedFixes ?? [],
      protocol: route?.protocol ?? '',
      routeStatus: route?.status ?? '',
      routeError,
      probeMs: Math.round(performance.now() - t0),
    };
  });

  console.log('\n');

  // ── Step 4: Build aggregate stats ────────────────────────────────────────────
  const byMode = {};
  const byVideoCodec = {};
  const byContainer = {};
  const errors = [];
  const withFixes = [];
  const noRoute = [];

  for (const r of results) {
    // Mode bucket
    const mode = r.mode?.toLowerCase() ?? 'unknown';
    if (!byMode[mode]) byMode[mode] = [];
    byMode[mode].push(r);

    // Video codec bucket
    const codec = (r.videoCodec || 'unknown').toLowerCase();
    if (!byVideoCodec[codec]) byVideoCodec[codec] = [];
    byVideoCodec[codec].push(r);

    // Container bucket
    const container = (r.container || 'unknown').toLowerCase();
    if (!byContainer[container]) byContainer[container] = [];
    byContainer[container].push(r);

    if (r.routeError) errors.push(r);
    if (r.suggestedFixes?.length > 0) withFixes.push(r);
    if (!r.mode || r.mode === 'unknown' || r.mode === 'error') noRoute.push(r);
  }

  const totalDurationHours = results.reduce((s, r) => s + r.durationSeconds, 0) / 3600;
  const totalSizeGB = results.reduce((s, r) => s + r.sizeBytes, 0) / 1_073_741_824;
  const probeAvgMs = results.reduce((s, r) => s + r.probeMs, 0) / results.length;
  const wallMs = performance.now() - startTime;

  // ── Step 5: Write JSON data ──────────────────────────────────────────────────
  const jsonPath = join(OUT_DIR, 'audit.json');
  writeFileSync(jsonPath, JSON.stringify({
    meta: {
      generatedAt: new Date().toISOString(),
      serverUrl: BASE_URL,
      clientProfile: CLIENT_PROF,
      totalFiles: results.length,
      totalDurationHours: +totalDurationHours.toFixed(2),
      totalSizeGB: +totalSizeGB.toFixed(3),
      wallTimeMs: Math.round(wallMs),
      probeAvgMs: Math.round(probeAvgMs),
    },
    results,
  }, null, 2));

  // ── Step 6: Write human-readable report ──────────────────────────────────────
  const lines = [];
  const hr = (char = '─', len = 68) => char.repeat(len);

  lines.push(hr('═'));
  lines.push('  XUVA LIBRARY PLAYBACK AUDIT');
  lines.push(`  Generated ${new Date().toLocaleString()}`);
  lines.push(hr('═'));
  lines.push('');

  // Summary
  lines.push('SUMMARY');
  lines.push(hr());
  lines.push(`  Files audited       : ${results.length.toLocaleString()}`);
  lines.push(`  Total library size  : ${totalSizeGB.toFixed(2)} GB`);
  lines.push(`  Total duration      : ${totalDurationHours.toFixed(1)} hours`);
  lines.push(`  Audit wall time     : ${fmtMs(wallMs)}`);
  lines.push(`  Avg route probe     : ${fmtMs(probeAvgMs)} / file`);
  lines.push('');

  // Playback mode breakdown
  lines.push('PLAYBACK MODE BREAKDOWN');
  lines.push(hr());
  const modeOrder = ['direct', 'direct_play', 'directplay', 'remux', 'audio_transcode', 'adaptive', 'transcode', 'video_transcode', 'error', 'unknown'];
  const modeLabels = {
    direct: '● Direct Play', direct_play: '● Direct Play', directplay: '● Direct Play',
    remux: '◎ Remux', audio_transcode: '◈ Audio Transcode',
    adaptive: '◉ Adaptive HLS', transcode: '▲ Video Transcode',
    video_transcode: '▲ Video Transcode', error: '✗ Error', unknown: '? Unknown',
  };
  const modeSeen = new Set();
  const modesSorted = [
    ...modeOrder.filter(m => byMode[m]?.length > 0),
    ...Object.keys(byMode).filter(m => !modeOrder.includes(m)),
  ];
  for (const mode of modesSorted) {
    if (!byMode[mode]) continue;
    const label = modeLabels[mode] ?? `  ${mode}`;
    const count = byMode[mode].length;
    const pct   = ((count / results.length) * 100).toFixed(1);
    const bar   = '█'.repeat(Math.round(parseFloat(pct) / 5));
    lines.push(`  ${pad(label, 22)}  ${pad(String(count), 5)}  (${pad(pct + '%', 6)})  ${bar}`);
  }
  lines.push('');

  // Video codec breakdown
  lines.push('VIDEO CODEC BREAKDOWN');
  lines.push(hr());
  const codecsSorted = Object.entries(byVideoCodec).sort((a, b) => b[1].length - a[1].length);
  for (const [codec, items] of codecsSorted) {
    const count = items.length;
    const pct   = ((count / results.length) * 100).toFixed(1);
    const avgBr = fmtBitrate(items.reduce((s, i) => s + i.bitrate, 0) / items.length);
    const heights = [...new Set(items.filter(i => i.height).map(i => `${i.height}p`))].sort().join(', ');
    lines.push(`  ${pad(codec.toUpperCase(), 12)}  ${pad(count + ' files', 12)}  ${pad(pct + '%', 7)}  avg ${pad(avgBr, 12)}  ${heights}`);
  }
  lines.push('');

  // Container breakdown
  lines.push('CONTAINER BREAKDOWN');
  lines.push(hr());
  const containersSorted = Object.entries(byContainer).sort((a, b) => b[1].length - a[1].length);
  for (const [container, items] of containersSorted) {
    const count = items.length;
    const pct   = ((count / results.length) * 100).toFixed(1);
    lines.push(`  ${pad(container.toUpperCase(), 10)}  ${pad(count + ' files', 12)}  ${pct}%`);
  }
  lines.push('');

  // Files needing attention
  if (withFixes.length > 0) {
    lines.push('FILES WITH SERVER SUGGESTIONS');
    lines.push(hr());
    lines.push(`  ${withFixes.length} file(s) flagged with suggested fixes:\n`);
    for (const r of withFixes.slice(0, 50)) {
      lines.push(`  ${r.relPath || r.name}`);
      lines.push(`    Mode: ${r.mode}  Codec: ${r.videoCodec}  ${r.width}×${r.height}`);
      for (const fix of r.suggestedFixes) {
        lines.push(`    → ${fix}`);
      }
      lines.push('');
    }
    if (withFixes.length > 50) lines.push(`  … and ${withFixes.length - 50} more (see audit.json)`);
    lines.push('');
  }

  // Errors
  if (errors.length > 0) {
    lines.push('ERRORS — FILES WITH NO PLAYBACK ROUTE');
    lines.push(hr());
    lines.push(`  ${errors.length} file(s) could not be routed:\n`);
    for (const r of errors.slice(0, 30)) {
      lines.push(`  ${r.relPath || r.name}`);
      lines.push(`    ${r.routeError}`);
    }
    if (errors.length > 30) lines.push(`  … and ${errors.length - 30} more (see audit.json)`);
    lines.push('');
  }

  // Transcode candidates — likely to hit CPU
  const transcodeFiles = results.filter(r =>
    ['transcode', 'video_transcode'].includes(r.mode?.toLowerCase())
  );
  if (transcodeFiles.length > 0) {
    lines.push('TRANSCODING CANDIDATES (will use server CPU)');
    lines.push(hr());
    lines.push(`  ${transcodeFiles.length} file(s) require video transcoding.\n`);
    // Group by reason code
    const byReason = {};
    for (const r of transcodeFiles) {
      const key = r.reasonCode || r.reasonText || 'unspecified';
      if (!byReason[key]) byReason[key] = [];
      byReason[key].push(r);
    }
    for (const [reason, items] of Object.entries(byReason)) {
      lines.push(`  Reason: ${reason} (${items.length} files)`);
      for (const item of items.slice(0, 5)) {
        lines.push(`    • ${item.videoCodec?.toUpperCase() ?? '?'} ${item.width}×${item.height}  ${fmtBitrate(item.bitrate)}  ${item.relPath?.split('/').pop() ?? item.name}`);
      }
      if (items.length > 5) lines.push(`    … and ${items.length - 5} more`);
      lines.push('');
    }
  }

  // Top 10 largest files
  lines.push('TOP 10 LARGEST FILES');
  lines.push(hr());
  const top10 = [...results].sort((a, b) => b.sizeBytes - a.sizeBytes).slice(0, 10);
  for (const r of top10) {
    const name = (r.relPath || r.name).split('/').pop() ?? '';
    lines.push(`  ${pad(fmtBytes(r.sizeBytes), 10)}  ${pad(r.mode ?? '?', 16)}  ${pad(r.videoCodec?.toUpperCase() ?? '?', 8)}  ${r.width}×${r.height}  ${name}`);
  }
  lines.push('');

  // Resolution distribution
  lines.push('RESOLUTION DISTRIBUTION');
  lines.push(hr());
  const resBuckets = { '4K (2160p+)': 0, '1440p': 0, 'Full HD (1080p)': 0, 'HD (720p)': 0, 'SD (<720p)': 0, 'Unknown': 0 };
  for (const r of results) {
    const h = r.height ?? 0;
    if (h >= 2160)      resBuckets['4K (2160p+)']++;
    else if (h >= 1440) resBuckets['1440p']++;
    else if (h >= 1080) resBuckets['Full HD (1080p)']++;
    else if (h >= 720)  resBuckets['HD (720p)']++;
    else if (h > 0)     resBuckets['SD (<720p)']++;
    else                resBuckets['Unknown']++;
  }
  for (const [label, count] of Object.entries(resBuckets)) {
    if (count === 0) continue;
    const pct = ((count / results.length) * 100).toFixed(1);
    const bar = '█'.repeat(Math.round(parseFloat(pct) / 5));
    lines.push(`  ${pad(label, 20)}  ${pad(String(count), 5)}  (${pad(pct + '%', 6)})  ${bar}`);
  }
  lines.push('');

  // Recommended manual test matrix (one sample per category)
  lines.push('RECOMMENDED MANUAL PLAYBACK TEST MATRIX');
  lines.push(hr());
  lines.push('  Pick one file from each category to test in the player:\n');

  function pickSample(filterFn, label) {
    const match = results.find(filterFn);
    if (match) {
      const name = (match.relPath || match.name).split('/').pop() ?? match.id;
      lines.push(`  ${pad(label, 30)}  ${match.id}`);
      lines.push(`    ${name}`);
      lines.push(`    ${match.videoCodec?.toUpperCase() ?? '?'} · ${match.width}×${match.height} · ${fmtBitrate(match.bitrate)} · ${match.mode}`);
    } else {
      lines.push(`  ${pad(label, 30)}  (no matching file in library)`);
    }
    lines.push('');
  }

  pickSample(r => r.mode === 'direct' || r.mode === 'direct_play' || r.mode === 'directplay', 'Direct Play');
  pickSample(r => r.mode === 'remux', 'Remux (container only)');
  pickSample(r => r.mode === 'audio_transcode', 'Audio Transcode');
  pickSample(r => r.mode === 'adaptive', 'Adaptive HLS');
  pickSample(r => ['transcode', 'video_transcode'].includes(r.mode), 'Video Transcode');
  pickSample(r => r.height >= 2160, '4K file');
  pickSample(r => r.audioStreams >= 3, 'Multi-audio (3+ tracks)');
  pickSample(r => r.subtitleStreams >= 2, 'Embedded subtitles');
  pickSample(r => r.videoCodec?.toLowerCase().includes('hevc') || r.videoCodec?.toLowerCase().includes('h265') || r.videoCodec?.toLowerCase() === 'hevc', 'H.265/HEVC');
  pickSample(r => r.videoCodec?.toLowerCase() === 'av1', 'AV1');
  pickSample(r => r.sizeBytes > 20_000_000_000, 'Extra-large file (>20GB)');

  lines.push(hr('═'));
  lines.push(`  Full data: ${jsonPath}`);
  lines.push(hr('═'));
  lines.push('');

  const reportText = lines.join('\n');
  const reportPath = join(OUT_DIR, 'audit-report.txt');
  writeFileSync(reportPath, reportText);

  // Print the report
  console.log(reportText);
  console.log(`\n  ✓ JSON data written to : ${jsonPath}`);
  console.log(`  ✓ Text report written to: ${reportPath}\n`);
}

main().catch(e => {
  console.error(`\n  Fatal: ${e.message}`);
  process.exit(1);
});
