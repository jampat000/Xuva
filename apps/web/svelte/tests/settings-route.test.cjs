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
	assert.match(source, /getClientBootstrap\(apiClient\)/);
	assert.match(source, /updateSettings\(\{\s*serverName:\s*nextName\s*\},\s*apiClient\)/);
	assert.match(source, /updateSettings\(\s*\{\s*librarySyncMode,\s*syncIntervalMins:\s*syncIntervalMins\s*\?\?\s*1440,\s*watchDebounceSecs:\s*watchDebounceSecs\s*\?\?\s*30,\s*probeBatchLimit\s*\}\s*,\s*apiClient\s*\)/);
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
	assert.match(source, /Library scan mode/);
	assert.match(source, /Scan interval/);
	assert.match(source, /Folder watch delay/);
	assert.match(source, /Media check batch size/);
	assert.match(source, /Save Scanning Settings/);
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
	assert.match(source, /Save Playback Policy/);
	assert.match(source, /Sign Out|Sign In/);
	assert.match(source, /Create Owner Account/);
	assert.match(source, /Open Access/);
	assert.match(source, /Sign in as the owner to manage Lorivo settings\./);
	assert.match(source, /Development access is active\. User management will be enabled before production\./);
	assert.match(source, /User management is not available yet\./);
	assert.match(source, /Device pairing will appear here when client pairing is implemented\./);
	assert.match(source, /Lorivo uses this name in the browser title and shares it with local clients\./);
	assert.match(source, /Local network discovery is not available in this build yet\./);
	assert.match(source, /Server name must be 50 characters or fewer\./);
	assert.match(source, /Playback setting saved\. Restart Lorivo to apply it\./);
	assert.match(source, /Saved\. Restart Lorivo for this change to fully take effect\./);
	assert.match(source, /Enter a scan interval of at least 15 minutes\./);
	assert.match(source, /This server still needs its first owner account\./);
	assert.match(source, /<ServerShell/);
	assert.match(source, /Needs attention|everything looks ready/i);
	assert.match(source, /Manual scans/);
	assert.match(source, /Automation/);
	assert.match(source, /Advanced scanning/);
	assert.match(source, /Title and artwork lookup/);
	assert.match(source, /Save Playback Policy/);
	assert.doesNotMatch(source, /legacy/i);
	assert.doesNotMatch(source, /Admin Dashboard/);
	assert.doesNotMatch(source, /Admin Controls/);
	assert.doesNotMatch(source, /<SettingsPanel title="Server"/);
	assert.doesNotMatch(source, /Provider setup/);
	assert.doesNotMatch(source, /Transcode Workers/);
	assert.doesNotMatch(source, /Editing existing folders is not available here yet\./);
	assert.doesNotMatch(source, /Playback preference editing is not available in this build\./);
	assert.doesNotMatch(source, /Provider API/i);
	assert.doesNotMatch(source, /\bLibrarySyncMode\b/);
	assert.doesNotMatch(source, /\bSyncIntervalMins\b/);
	assert.doesNotMatch(source, /\bWatchDebounceSecs\b/);
	assert.doesNotMatch(source, /\bProbeBatchLimit\b/);
	assert.doesNotMatch(source, /hostname/i);
});

test('setup route asks for a user-facing server name', () => {
	const source = read('src/routes/setup/+page.svelte');
	assert.match(source, /Server name/);
	assert.match(source, /Lorivo uses this name in the browser title and shares it with local clients\./);
	assert.match(source, /Local network discovery is not available in this build yet\./);
	assert.match(source, /maxlength="50"/);
	assert.match(source, /updateSettings\(\{\s*serverName:\s*serverNameValue\s*\}/);
	assert.match(source, /serverName = \$state\('Lorivo'\)/);
	assert.match(source, /!session\?\.user && !session\?\.authDisabled/);
	assert.doesNotMatch(source, /hostname/i);
});

test('settings navigation stays shared across desktop sidebar and settings drawer', () => {
	const navSource = read('src/lib/components/shell/SettingsNav.svelte');
	const sidebarSource = read('src/lib/components/shell/ServerSidebar.svelte');
	const shellSource = read('src/lib/components/shell/ServerShell.svelte');

	assert.match(sidebarSource, /<SettingsBrand/);
	assert.match(sidebarSource, /<SettingsNav\s+\{active\}\s*\/>/);
	assert.match(sidebarSource, /<SettingsNav section="secondary"\s*\/>/);
	assert.match(shellSource, /<SettingsBrand/);
	assert.match(shellSource, /<SettingsNav\s+\{active\}\s*\/>/);
	assert.match(shellSource, /<SettingsNav section="secondary"\s*\/>/);
	assert.doesNotMatch(shellSource, /class="app-drawer__link"\s+href="\/settings#/);

	assert.match(navSource, /label="Dashboard"\s+href="\/settings#dashboard"/);
	assert.match(navSource, /label="Library"\s+href="\/settings#library"/);
	assert.match(navSource, /label="Scanning"\s+href="\/settings#scanning"/);
	assert.match(navSource, /label="Metadata"\s+href="\/settings#metadata"/);
	assert.match(navSource, /label="Playback"\s+href="\/settings#playback"/);
	assert.match(navSource, /label="Access"\s+href="\/settings#access"/);
	assert.match(navSource, /label="About"\s+href="\/settings#about"/);
	assert.match(navSource, /label="Back to Media"\s+href="\/"/);
	assert.doesNotMatch(navSource, /label="Server"\s+href="\/settings#server"/);
	assert.doesNotMatch(navSource, /label="Appearance"/);
	assert.doesNotMatch(navSource, /label="Diagnostics"/);
	assert.doesNotMatch(navSource, /href="\/admin"/);
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
