const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

const appRoot = path.resolve(__dirname, '..');

function read(file) {
	return fs.readFileSync(path.join(appRoot, file), 'utf8');
}

test('movies route uses real browse API and Lorivo controls without product preview mode', () => {
	const source = read('src/routes/movies/+page.svelte');
	assert.match(source, /getMovies\(apiClient,\s*500\)/);
	assert.match(source, /scanMovies\(apiClient,\s*50\)/);
	assert.match(source, /refreshMetadataBatch\('movie'/);
	assert.match(source, /Search movies/);
	assert.match(source, /Scan Movies/);
	assert.match(source, /Refresh Metadata/);
	assert.match(source, /<LorivoShell>/);
	assert.doesNotMatch(source, /resolvePreviewMode/);
	assert.doesNotMatch(source, /previewMovieRows/);
	assert.doesNotMatch(source, /preview=1/);
});

test('tv route uses real browse API and Lorivo controls without product preview mode', () => {
	const source = read('src/routes/tv/+page.svelte');
	assert.match(source, /getSeries\(apiClient,\s*500\)/);
	assert.match(source, /scanTV\(apiClient,\s*50\)/);
	assert.match(source, /refreshMetadataBatch\('series'/);
	assert.match(source, /Search TV/);
	assert.match(source, /Scan TV/);
	assert.match(source, /Refresh Metadata/);
	assert.match(source, /Title/);
	assert.match(source, /Year/);
	assert.match(source, /Seasons/);
	assert.match(source, /Episodes/);
	assert.match(source, /<LorivoShell>/);
	assert.doesNotMatch(source, /resolvePreviewMode/);
	assert.doesNotMatch(source, /previewSeriesRows/);
	assert.doesNotMatch(source, /preview=1/);
});

test('home route uses backend data and keeps media-first Lorivo sections', () => {
	const routeSource = read('src/routes/+page.svelte');
	assert.match(routeSource, /getClientHome\(apiClient,\s*24\)/);
	assert.match(routeSource, /getPlaybackRecent\(apiClient,\s*12\)/);
	assert.match(routeSource, /buildHomeViewModel/);
	assert.match(routeSource, /Continue Watching/);
	assert.match(routeSource, /Recently Added Movies/);
	assert.match(routeSource, /Recently Added TV/);
	assert.doesNotMatch(routeSource, /resolvePreviewMode/);
	assert.doesNotMatch(routeSource, /LorivoMediaHome/);
	assert.doesNotMatch(routeSource, /preview=1/);
});
