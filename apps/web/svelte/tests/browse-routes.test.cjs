const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

const appRoot = path.resolve(__dirname, '..');

function read(file) {
	return fs.readFileSync(path.join(appRoot, file), 'utf8');
}

test('movies route is wired to browse API and controls', () => {
	const source = read('src/routes/movies/+page.svelte');
	assert.match(source, /getMovies\(apiClient,\s*500\)/);
	assert.match(source, /scanMovies\(apiClient,\s*50\)/);
	assert.match(source, /refreshMetadataBatch\('movie'/);
	assert.match(source, /<BrowseHeader title="Movies"/);
	assert.match(source, /Browse your movie library\./);
	assert.match(source, /Needs Review/);
	assert.match(source, /Metadata Pending/);
	assert.match(source, /<MediaShell active="movies"/);
	assert.match(source, /resolvePreviewMode\(new URL\(window\.location\.href\)\.searchParams\)/);
	assert.match(source, /if \((previewMode|activePreviewMode) && movieRows\.length === 0\)/);
});

test('tv route is wired to browse API and controls', () => {
	const source = read('src/routes/tv/+page.svelte');
	assert.match(source, /getSeries\(apiClient,\s*500\)/);
	assert.match(source, /scanTV\(apiClient,\s*50\)/);
	assert.match(source, /refreshMetadataBatch\('series'/);
	assert.match(source, /<BrowseHeader title="TV Shows"/);
	assert.match(source, /Browse your TV library\./);
	assert.match(source, /Multi-Season/);
	assert.match(source, /With Episodes/);
	assert.match(source, /Unknown Year/);
	assert.match(source, /Title/);
	assert.match(source, /Year/);
	assert.match(source, /Seasons/);
	assert.match(source, /Episodes/);
	assert.match(source, /<MediaShell active="tv"/);
	assert.match(source, /resolvePreviewMode\(new URL\(window\.location\.href\)\.searchParams\)/);
	assert.match(source, /if \((previewMode|activePreviewMode) && seriesRows\.length === 0\)/);
});

test('home route remains media-first and excludes operator telemetry labels', () => {
	const routeSource = read('src/routes/+page.svelte');
	const componentSource = read('src/lib/components/home/LorivoMediaHome.svelte');
	assert.match(routeSource, /resolvePreviewMode\(page\.url\.searchParams\)/);
	assert.match(routeSource, /<LorivoMediaHome/);
	assert.match(componentSource, /Continue Watching/);
	assert.match(componentSource, /Recently Added Movies/);
	assert.match(componentSource, /Recently Added TV/);
	assert.doesNotMatch(componentSource, /Server Status/);
	assert.doesNotMatch(componentSource, /Recent Activity/);
	assert.doesNotMatch(componentSource, /CPU|RAM|Network|Transcoding/);
	assert.doesNotMatch(routeSource, /<MediaShell active="home"/);
});

test('home preview fallback remains explicit for 401 responses', () => {
	const source = read('src/routes/+page.svelte');
	assert.match(source, /<LorivoMediaHome[\s\S]*previewMode=\{previewMode\}/);
	assert.doesNotMatch(source, /Add your first library/);
	assert.doesNotMatch(source, /Open Lorivo Settings/);
	assert.doesNotMatch(source, /effectivePreviewMode\s*=\s*true/);
	assert.doesNotMatch(source, /hasTruthyFlagInHref\(/);
});
