import { readFile } from 'node:fs/promises';
import path from 'node:path';
import process from 'node:process';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const appDir = path.resolve(scriptDir, '..');
const staticNextDir = path.resolve(appDir, '../../../server/internal/webapp/static-next');
const localBuildInfoPath = path.join(staticNextDir, 'build-info.json');

const rootOrigin = String(
	process.argv[2] || process.env.LORIVO_ROOT_ORIGIN || process.env.LORIVO_ORIGIN || 'http://127.0.0.1:18100'
)
	.trim()
	.replace(/\/+$/, '');

function log(line = '') {
	process.stdout.write(`${line}\n`);
}

function asText(value) {
	return String(value ?? '').trim();
}

async function fetchJSON(url) {
	const response = await fetch(url, { redirect: 'follow' });
	const text = await response.text();
	if (!response.ok) {
		throw new Error(`HTTP ${response.status} for ${url}: ${text.slice(0, 180)}`);
	}
	try {
		return JSON.parse(text);
	} catch {
		throw new Error(`Invalid JSON at ${url}`);
	}
}

async function tryDiscoverDynamicRoute(kind) {
	const endpoint = kind === 'movie' ? '/api/movies?limit=1' : '/api/series?limit=1';
	try {
		const response = await fetch(`${rootOrigin}${endpoint}`, { redirect: 'follow' });
		const text = await response.text();
		if (!response.ok) return { id: '', reason: `API ${response.status}` };
		const payload = JSON.parse(text);
		const list = Array.isArray(payload?.movies)
			? payload.movies
			: Array.isArray(payload?.series)
				? payload.series
				: [];
		const id = asText(list?.[0]?.id);
		return id ? { id, reason: '' } : { id: '', reason: 'no items' };
	} catch (error) {
		return { id: '', reason: error instanceof Error ? error.message : 'request failed' };
	}
}

async function checkHTMLRoute(route) {
	const url = `${rootOrigin}${route}`;
	const response = await fetch(url, { redirect: 'follow' });
	const body = await response.text();
	const contentType = asText(response.headers.get('content-type')).toLowerCase();
	const hasSvelteBootstrap = body.includes('__sveltekit_') && body.includes('/_app/immutable/');
	const hasFallback404 = body.includes('<h1>404</h1>') && body.includes('<p>Not Found</p>');
	const ok = response.status === 200 && contentType.includes('text/html') && hasSvelteBootstrap && !hasFallback404;
	return {
		route,
		status: response.status,
		contentType,
		ok,
		note: ok ? 'ok' : 'unexpected response content'
	};
}

async function checkNotSupportedRoute(route) {
	const url = `${rootOrigin}${route}`;
	const response = await fetch(url, { redirect: 'manual' });
	const status = response.status;
	const ok = status === 404 || status === 405;
	return {
		route,
		status,
		ok,
		note: ok ? 'not-supported-as-expected' : 'unexpectedly reachable'
	};
}

const localBuildInfoRaw = await readFile(localBuildInfoPath, 'utf8');
const localBuildInfo = JSON.parse(localBuildInfoRaw);

const remoteBuildInfo = await fetchJSON(`${rootOrigin}/build-info.json?ts=${Date.now()}`);
if (asText(localBuildInfo.buildID) !== asText(remoteBuildInfo.buildID)) {
	log('FAIL: active server is serving a stale embedded Svelte bundle.');
	log(`Local buildID : ${asText(localBuildInfo.buildID)}`);
	log(`Served buildID: ${asText(remoteBuildInfo.buildID)}`);
	log('Action: restart/rebuild the Go server, then re-run this check.');
	process.exit(1);
}

const movieDynamic = await tryDiscoverDynamicRoute('movie');
const tvDynamic = await tryDiscoverDynamicRoute('tv');
const dynamicNotes = [];
const dynamicRoutes = [];

if (movieDynamic.id) {
	dynamicRoutes.push(`/movies/${encodeURIComponent(movieDynamic.id)}`);
} else {
	dynamicNotes.push(`movie details skipped: ${movieDynamic.reason || 'no id'}`);
}
if (tvDynamic.id) {
	dynamicRoutes.push(`/tv/${encodeURIComponent(tvDynamic.id)}`);
} else {
	dynamicNotes.push(`tv details skipped: ${tvDynamic.reason || 'no id'}`);
}

const routesToCheck = [
	'/',
	'/signin',
	'/setup',
	'/movies',
	'/tv',
	...dynamicRoutes,
	'/settings',
	'/admin',
	'/collections',
	'/watchlist',
	'/continue-watching',
	'/recently-added'
];
const notSupportedRoutes = ['/legacy', '/legacy/settings', '/next', '/next/movies', '/next/settings'];

const results = [];
for (const route of routesToCheck) {
	results.push(await checkHTMLRoute(route));
}
const unsupportedResults = [];
for (const route of notSupportedRoutes) {
	unsupportedResults.push(await checkNotSupportedRoute(route));
}

log(`Root origin: ${rootOrigin}`);
log(`Build ID: ${asText(remoteBuildInfo.buildID)}`);
log('');
log('Root route smoke results');
log('------------------------');
for (const result of results) {
	log(`${result.ok ? 'PASS' : 'FAIL'}  ${result.status}  ${result.route}`);
}
log('');
log('Unsupported route checks');
log('------------------------');
for (const result of unsupportedResults) {
	log(`${result.ok ? 'PASS' : 'FAIL'}  ${result.status}  ${result.route}`);
}

if (dynamicNotes.length > 0) {
	log('');
	log('Dynamic route discovery notes');
	log('-----------------------------');
	for (const note of dynamicNotes) log(`- ${note}`);
}

const failures = results.filter((result) => !result.ok);
const unsupportedFailures = unsupportedResults.filter((result) => !result.ok);
if (failures.length > 0 || unsupportedFailures.length > 0) {
	process.exit(1);
}
