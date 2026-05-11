const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

const appRoot = path.resolve(__dirname, '..');

function read(file) {
	return fs.readFileSync(path.join(appRoot, file), 'utf8');
}

test('admin route uses existing live operator APIs and event stream', () => {
	const source = read('src/routes/admin/+page.svelte');

	assert.match(source, /getCatalogSummary\(apiClient\)/);
	assert.match(source, /getCatalogHealth\(apiClient\)/);
	assert.match(source, /getSystemStatus\(apiClient\)/);
	assert.match(source, /getScans\(apiClient\)/);
	assert.match(source, /getProbes\(apiClient\)/);
	assert.match(source, /getWork\(apiClient\)/);
	assert.match(source, /getDownloads\(apiClient\)/);
	assert.match(source, /getSessions\(apiClient\)/);
	assert.match(source, /createEventStream\(\)/);
	assert.match(source, /Admin controls are read-only in this build\./);
	assert.match(source, /<ServerShell/);
	assert.doesNotMatch(source, /legacy/i);
	assert.doesNotMatch(source, /send<.*\/api\//s);
});
