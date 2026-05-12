const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

const appRoot = path.resolve(__dirname, '..');

function read(file) {
	return fs.readFileSync(path.join(appRoot, file), 'utf8');
}

test('settings route uses existing settings APIs and real user actions', () => {
	const source = read('src/routes/settings/+page.svelte');

	assert.match(source, /getSettings\(apiClient\)/);
	assert.match(source, /getPerformanceSettings\(apiClient\)/);
	assert.match(source, /getSystemStatus\(apiClient\)/);
	assert.match(source, /getCatalogSummary\(apiClient\)/);
	assert.match(source, /getCatalogHealth\(apiClient\)/);
	assert.match(source, /getScans\(apiClient\)/);
	assert.match(source, /getProbes\(apiClient\)/);
	assert.match(source, /getWork\(apiClient\)/);
	assert.match(source, /getDownloads\(apiClient\)/);
	assert.match(source, /getSessions\(apiClient\)/);
	assert.match(source, /createEventStream\(\)/);
	assert.match(source, /scanMovies\(apiClient,\s*50\)/);
	assert.match(source, /scanTV\(apiClient,\s*50\)/);
	assert.match(source, /refreshMetadataBatch\('movie'/);
	assert.match(source, /refreshMetadataBatch\('series'/);
	assert.match(source, /Library/);
	assert.match(source, /Scanning/);
	assert.match(source, /Metadata/);
	assert.match(source, /Playback/);
	assert.match(source, /Access/);
	assert.match(source, /About/);
	assert.match(source, /Dashboard/);
	assert.match(source, /Library Setup/);
	assert.match(source, /Sign Out/);
	assert.match(source, /Playback preference editing is not available in this build\./);
	assert.match(source, /<ServerShell/);
	assert.doesNotMatch(source, /legacy/i);
	assert.doesNotMatch(source, /Admin Dashboard/);
	assert.doesNotMatch(source, /Admin Controls/);
	assert.doesNotMatch(source, /<SettingsPanel title="Server"/);
	assert.doesNotMatch(source, /Provider setup/);
	assert.doesNotMatch(source, /Transcode Workers/);
	assert.doesNotMatch(source, /send<.*\/api\/settings/s);
});

test('server shell keeps settings mode section navigation', () => {
	const source = read('src/lib/components/shell/ServerSidebar.svelte');
	assert.match(source, /label="Dashboard"\s+href="\/settings#dashboard"/);
	assert.match(source, /label="Library"\s+href="\/settings#library"/);
	assert.match(source, /label="Scanning"\s+href="\/settings#scanning"/);
	assert.match(source, /label="Metadata"\s+href="\/settings#metadata"/);
	assert.match(source, /label="Playback"\s+href="\/settings#playback"/);
	assert.match(source, /label="Access"\s+href="\/settings#access"/);
	assert.match(source, /label="About"\s+href="\/settings#about"/);
	assert.match(source, /label="Back to Media"\s+href="\/"/);
	assert.doesNotMatch(source, /label="Server"\s+href="\/settings#server"/);
	assert.doesNotMatch(source, /label="Appearance"/);
	assert.doesNotMatch(source, /href="\/admin"/);
});
