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
	assert.match(source, /updateSettings\(\{\s*serverName:\s*nextName\s*\},\s*apiClient\)/);
	assert.match(source, /updateSettings\(\{\s*playbackPolicy:\s*playbackPolicyDraft\s*\},\s*apiClient\)/);
	assert.match(source, /getPerformanceSettings\(apiClient\)/);
	assert.match(source, /getSystemStatus\(apiClient\)/);
	assert.match(source, /getCatalogSummary\(apiClient\)/);
	assert.match(source, /getCatalogHealth\(apiClient\)/);
	assert.match(source, /getScans\(apiClient\)/);
	assert.match(source, /getProbes\(apiClient\)/);
	assert.match(source, /getWork\(apiClient\)/);
	assert.match(source, /getDownloads\(apiClient\)/);
	assert.match(source, /getSessions\(apiClient\)/);
	assert.match(source, /logout\(apiClient\)/);
	assert.match(source, /startLibraryScan\(id,\s*apiClient\)/);
	assert.match(source, /deleteLibrary\(id,\s*apiClient\)/);
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
	assert.match(source, /Server name/);
	assert.match(source, /Movies library/);
	assert.match(source, /TV library/);
	assert.match(source, /Folder/);
	assert.match(source, /Scan/);
	assert.match(source, /Remove/);
	assert.match(source, /Save Playback Setting/);
	assert.match(source, /Sign Out|Sign In/);
	assert.match(source, /This is the name devices on your home network will use to identify this Lorivo server\./);
	assert.match(source, /Server name must be 50 characters or fewer\./);
	assert.match(source, /Playback setting saved\. Restart Lorivo to apply it\./);
	assert.match(source, /Sign in with the owner account to scan or remove libraries\./);
	assert.match(source, /Sign in with the owner account to change playback settings\./);
	assert.match(source, /<ServerShell/);
	assert.match(source, /Needs attention|everything looks ready/i);
	assert.doesNotMatch(source, /legacy/i);
	assert.doesNotMatch(source, /Admin Dashboard/);
	assert.doesNotMatch(source, /Admin Controls/);
	assert.doesNotMatch(source, /<SettingsPanel title="Server"/);
	assert.doesNotMatch(source, /User management/);
	assert.doesNotMatch(source, /Provider setup/);
	assert.doesNotMatch(source, /Transcode Workers/);
	assert.doesNotMatch(source, /Editing existing folders is not available here yet\./);
	assert.doesNotMatch(source, /Playback preference editing is not available in this build\./);
	assert.doesNotMatch(source, /Provider API/i);
	assert.doesNotMatch(source, /hostname/i);
});

test('setup route asks for a user-facing server name', () => {
	const source = read('src/routes/setup/+page.svelte');
	assert.match(source, /Server name/);
	assert.match(source, /This is the name devices on your home network will use to identify this Lorivo server\./);
	assert.match(source, /maxlength="50"/);
	assert.match(source, /updateSettings\(\{\s*serverName:\s*serverNameValue\s*\}/);
	assert.match(source, /serverName = \$state\('Lorivo'\)/);
	assert.match(source, /!session\?\.user && !session\?\.authDisabled/);
	assert.doesNotMatch(source, /hostname/i);
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
	assert.doesNotMatch(source, /label="Diagnostics"/);
	assert.doesNotMatch(source, /href="\/admin"/);
});

test('settings shell does not render media search', () => {
	const source = read('src/lib/components/shell/ServerShell.svelte');
	assert.doesNotMatch(source, /AppTopbar/);
	assert.doesNotMatch(source, /bind:searchValue/);
	assert.doesNotMatch(source, /searchValue/);
});

test('home first-run copy stays user-facing', () => {
	const source = read('src/routes/+page.svelte');
	assert.match(source, /Review your library setup and scan status\./);
	assert.doesNotMatch(source, /providers, and server status/i);
});
