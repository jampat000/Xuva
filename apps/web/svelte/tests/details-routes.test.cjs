const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

const appRoot = path.resolve(__dirname, '..');

function read(file) {
	return fs.readFileSync(path.join(appRoot, file), 'utf8');
}

test('movie details route is wired to existing movie/media/playback APIs', () => {
	const source = read('src/routes/movies/[id]/+page.svelte');
	assert.match(source, /getMovieDetail\(/);
	assert.match(source, /listMediaSources\(/);
	assert.match(source, /getMediaSourceDetail\(/);
	assert.match(source, /getMediaSourceTracks\(/);
	assert.match(source, /getMediaSourceSubtitles\(/);
	assert.match(source, /getPlaybackState\(/);
	assert.match(source, /getPlaybackDecision\(/);
	assert.match(source, /getPlaybackRoute\(/);
	assert.match(source, /\/play\/\$\{encodeURIComponent\(source\.mediaSourceId\)\}/);
	assert.match(source, /previewMode && isMoviePreviewId\(routeMovieId\)/);
	assert.match(source, /buildMovieDetailPreview/);
	assert.doesNotMatch(source, /arrive in Phase 5/i);
});

test('tv details route is wired to existing series/media/playback APIs', () => {
	const source = read('src/routes/tv/[id]/+page.svelte');
	assert.match(source, /getSeriesDetail\(/);
	assert.match(source, /listMediaSources\(/);
	assert.match(source, /getMediaSourceDetail\(/);
	assert.match(source, /getMediaSourceTracks\(/);
	assert.match(source, /getMediaSourceSubtitles\(/);
	assert.match(source, /getPlaybackState\(/);
	assert.match(source, /getPlaybackDecision\(/);
	assert.match(source, /\/play\/\$\{encodeURIComponent\(episode\.mediaSourceId\)\}/);
	assert.match(source, /previewMode && isSeriesPreviewId\(routeSeriesId\)/);
	assert.match(source, /buildSeriesDetailPreview/);
	assert.doesNotMatch(source, /arrive in Phase 5/i);
});
