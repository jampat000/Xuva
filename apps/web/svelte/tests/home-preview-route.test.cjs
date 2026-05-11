const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

const routePath = path.resolve(__dirname, '../src/routes/+page.svelte');

test('preview route is hard-gated by ?preview=1 and bypasses normal home states', () => {
	const source = fs.readFileSync(routePath, 'utf8');
	assert.match(source, /resolvePreviewMode\(page\.url\.searchParams\)/);
	assert.match(source, /<LorivoMediaHome[\s\S]*previewMode=\{previewMode\}/);
	assert.doesNotMatch(source, /Preview mode is currently unavailable/);
	assert.doesNotMatch(source, /Add your first library/i);
	assert.doesNotMatch(source, /Open Vyrden Settings/i);
	assert.doesNotMatch(source, /Calo/i);
});
