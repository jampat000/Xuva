const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

const appRoot = path.resolve(__dirname, '..');

function read(file) {
	return fs.readFileSync(path.join(appRoot, file), 'utf8');
}

test('admin route redirects to settings compatibility surface', () => {
	const source = read('src/routes/admin/+page.svelte');

	assert.match(source, /window\.location\.replace\('\/settings#server'\)/);
	assert.match(source, /Opening Settings/);
	assert.match(source, /Open Settings/);
	assert.doesNotMatch(source, /Dashboard/);
	assert.doesNotMatch(source, /Controls/);
	assert.doesNotMatch(source, /legacy/i);
});
