const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

const appRoot = path.resolve(__dirname, '..');

test('svelte config uses root base path for cutover', () => {
	const configPath = path.join(appRoot, 'svelte.config.js');
	const config = fs.readFileSync(configPath, 'utf8');

	assert.match(config, /static-next/);
});

test('go static-next directory exists', () => {
	const staticNextPath = path.resolve(appRoot, '../../../server/internal/webapp/static-next');
	assert.equal(fs.existsSync(staticNextPath), true);
	assert.equal(fs.statSync(staticNextPath).isDirectory(), true);
});

test('package scripts include root smoke verification only', () => {
	const packagePath = path.join(appRoot, 'package.json');
	const payload = JSON.parse(fs.readFileSync(packagePath, 'utf8'));
	assert.equal(typeof payload.scripts?.['smoke:root'], 'string');
	assert.match(payload.scripts['smoke:root'], /verify-root-routes\.mjs/);
	assert.equal(payload.scripts?.['smoke:next'], undefined);
});

test('publish script generates build marker and required smoke route list', () => {
	const scriptPath = path.join(appRoot, 'scripts/publish-go-static.mjs');
	const source = fs.readFileSync(scriptPath, 'utf8');
	assert.match(source, /build-info\.json/);
	assert.match(source, /requiredSmokeRoutes/);
	assert.match(source, /\/continue-watching/);
	assert.match(source, /\/recently-added/);
	assert.match(source, /routePatterns/);
	assert.doesNotMatch(source, /compatibilityBasePaths/);
});

test('design-system component exports include Lorivo primitives', () => {
	const indexPath = path.join(appRoot, 'src/lib/components/index.ts');
	const exportsFile = fs.readFileSync(indexPath, 'utf8');

	const required = [
		'LorivoShell',
		'LorivoBrand',
		'LorivoSidebar',
		'SidebarItem',
		'SidebarSection',
		'SidebarUser',
		'LorivoButton',
		'LorivoPanel',
		'LorivoSurface',
		'LorivoSearch',
		'LorivoEmptyState',
		'LorivoStat',
		'LorivoActionList',
		'HeroSurface',
		'MediaRow',
		'ResumeTile',
		'PosterCard',
		'ArtworkFallback',
		'ProgressBar',
		'MediaCompanionPanel',
		'ViewerQuickActions',
		'AdminPanel',
		'SettingsSection',
		'FormRow',
		'LiveStatusBadge',
		'ActivityListShell'
	];

	for (const symbol of required) {
		assert.match(exportsFile, new RegExp(`\\b${symbol}\\b`));
	}
});

test('root home route no longer includes showroom copy', () => {
	const pagePath = path.join(appRoot, 'src/routes/+page.svelte');
	const homeComponentPath = path.join(appRoot, 'src/lib/components/home/LorivoMediaHome.svelte');
	const pageSource = fs.readFileSync(pagePath, 'utf8');
	const homeComponentSource = fs.readFileSync(homeComponentPath, 'utf8');

	assert.doesNotMatch(pageSource, /Design System Preview/);
	assert.doesNotMatch(pageSource, /Operator and Dashboard Primitives/);
	assert.match(pageSource, /<LorivoMediaHome/);
	assert.match(homeComponentSource, /Continue Watching/);
	assert.match(homeComponentSource, /Recently Added Movies/);
});
