const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

const appRoot = path.resolve(__dirname, '..');

function read(file) {
	return fs.readFileSync(path.join(appRoot, file), 'utf8');
}

test('settings route uses existing operator/settings APIs in read-only mode', () => {
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
	assert.match(source, /Settings editing is not available yet\./);
	assert.match(source, /<ServerShell/);
	assert.doesNotMatch(source, /legacy/i);
	assert.doesNotMatch(source, /send<.*\/api\/settings/s);
});

test('server shell keeps operator navigation in management mode', () => {
	const source = read('src/lib/components/shell/ServerSidebar.svelte');
	assert.match(source, /label="Dashboard"\s+href="\/admin"/);
	assert.match(source, /label="Settings"\s+href="\/settings"/);
	assert.match(source, /label="Back to Media"\s+href="\/"/);
});
