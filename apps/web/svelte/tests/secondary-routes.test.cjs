const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

const appRoot = path.resolve(__dirname, '..');

function read(file) {
	return fs.readFileSync(path.join(appRoot, file), 'utf8');
}

test('media shell drawer keeps viewer-first navigation and secondary server links', () => {
	const source = read('src/lib/components/shell/MediaShell.svelte');
	assert.match(source, /label:\s*'Home',\s*href:\s*'\/'/);
	assert.match(source, /label:\s*'Movies',\s*href:\s*'\/movies'/);
	assert.match(source, /label:\s*'TV Shows',\s*href:\s*'\/tv'/);
	assert.match(source, /label:\s*'Continue Watching',\s*href:\s*'\/continue-watching'/);
	assert.match(source, /label:\s*'Recently Added',\s*href:\s*'\/recently-added'/);
	assert.match(source, /label:\s*'Watchlist',\s*href:\s*'\/watchlist'/);
	assert.match(source, /label:\s*'Collections',\s*href:\s*'\/collections'/);
	assert.match(source, /label:\s*'Manage Server',\s*href:\s*'\/admin'/);
	assert.match(source, /label:\s*'Server Settings',\s*href:\s*'\/settings'/);
});

test('watchlist route uses shared secondary loader and explicit unavailable state', () => {
	const source = read('src/routes/watchlist/+page.svelte');
	assert.match(source, /loadSecondaryRouteContext/);
	assert.match(source, /resolvePreviewMode/);
	assert.match(source, /Watchlist is not available yet/);
});

test('continue-watching route uses existing home\/playback context and explicit state messaging', () => {
	const source = read('src/routes/continue-watching/+page.svelte');
	assert.match(source, /loadSecondaryRouteContext/);
	assert.match(source, /Continue Watching is not available yet/);
	assert.match(source, /ResumeTile/);
});

test('recently-added route uses existing home row and media card rendering', () => {
	const source = read('src/routes/recently-added/+page.svelte');
	assert.match(source, /findRow\(context\.homePayload,\s*'recently-added'\)/);
	assert.match(source, /Recently Added Movies/);
	assert.match(source, /Recently Added TV/);
	assert.match(source, /<MediaRow/);
});

test('collections route is explicit about current availability', () => {
	const source = read('src/routes/collections/+page.svelte');
	assert.match(source, /findRow\(context\.homePayload,\s*'collections'\)/);
	assert.match(source, /Collections are not available yet/);
});

test('media row hides scrollbars and uses overflow-aware arrow controls', () => {
	const source = read('src/lib/components/media/MediaRow.svelte');
	assert.match(source, /scrollbar-width:\s*none/);
	assert.match(source, /::-webkit-scrollbar\s*\{\s*display:\s*none/);
	assert.match(source, /canScrollPrev/);
	assert.match(source, /canScrollNext/);
	assert.match(source, /Scroll media row left/);
	assert.match(source, /Scroll media row right/);
	assert.match(source, /max-width:\s*900px[\s\S]*display:\s*none !important/);
});
