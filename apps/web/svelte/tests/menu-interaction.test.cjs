const test = require('node:test');
const assert = require('node:assert/strict');
const net = require('node:net');
const { chromium } = require('playwright');

async function getFreePort() {
	return new Promise((resolve, reject) => {
		const server = net.createServer();
		server.unref();
		server.on('error', reject);
		server.listen(0, '127.0.0.1', () => {
			const address = server.address();
			server.close(() => resolve(address.port));
		});
	});
}

async function waitForServer(url) {
	const startedAt = Date.now();
	while (Date.now() - startedAt < 30000) {
		try {
			const response = await fetch(url);
			if (response.status < 500) return;
		} catch {
			await new Promise((resolve) => setTimeout(resolve, 250));
		}
	}
	throw new Error('dev server did not become ready');
}

async function launchDevServer() {
	const port = await getFreePort();
	const { createServer } = await import('vite');
	const server = await createServer({
		root: process.cwd(),
		logLevel: 'silent',
		server: {
			host: '127.0.0.1',
			port,
			strictPort: true
		}
	});
	await server.listen();
	const baseURL = `http://127.0.0.1:${port}`;
	await waitForServer(baseURL);
	return { server, baseURL };
}

function apiPayload(pathname) {
	if (pathname === '/api/auth/session') {
		return { user: { id: 'local', username: 'local', displayName: 'Local User', role: 'Local Account' } };
	}
	if (pathname === '/api/client/home') return { rows: [], actions: {} };
	if (pathname === '/api/playback/recent') return { recent: [] };
	if (pathname === '/api/libraries') return { libraries: [] };
	if (pathname === '/api/movies') return { movies: [] };
	if (pathname === '/api/series') return { series: [] };
	if (pathname === '/api/catalog/summary') return { movies: 0, series: 0, episodes: 0 };
	if (pathname === '/api/catalog/health') return {};
	if (pathname === '/api/system/status') return { cpu: {}, memory: {} };
	if (pathname === '/api/settings') {
		return { config: { metadataProviders: { automatic: [], managedOverrides: [] } }, libraries: [] };
	}
	if (pathname === '/api/settings/performance') return {};
	if (pathname === '/api/scans') return { scans: [] };
	if (pathname === '/api/probes') return { probes: [] };
	if (pathname === '/api/work') return { work: [] };
	if (pathname === '/api/downloads') return { downloads: [] };
	if (pathname === '/api/sessions') return { sessions: [] };
	return {};
}

async function installApiMocks(page) {
	await page.route('**/api/**', (route) => {
		const url = new URL(route.request().url());
		if (!url.pathname.startsWith('/api/')) return route.continue();
		if (url.pathname === '/api/events') return route.fulfill({ status: 204, body: '' });
		return route.fulfill({
			status: 200,
			contentType: 'application/json',
			body: JSON.stringify(apiPayload(url.pathname))
		});
	});
	await page.route('**/build-info.json', (route) =>
		route.fulfill({
			status: 200,
			contentType: 'application/json',
			body: JSON.stringify({ buildID: 'test', gitCommit: 'test', sourceApp: 'apps/web/svelte' })
		})
	);
}

async function assertNoPersistentMediaPills(page) {
	for (const label of ['Home', 'Movies', 'TV']) {
		assert.equal(
			await page.locator('header nav[aria-label="Media navigation"]').getByRole('link', { name: label, exact: true }).count(),
			0
		);
	}
}

async function assertLeftDrawer(page, drawer, viewport) {
	let box = await drawer.boundingBox();
	assert.ok(box, 'drawer should have a bounding box');
	const startedAt = Date.now();
	while (Date.now() - startedAt < 1200) {
		if (box.x >= -1 && box.x <= 1) break;
		await page.waitForTimeout(40);
		box = await drawer.boundingBox();
		assert.ok(box, 'drawer should have a bounding box');
	}
	assert.ok(box.x >= -1 && box.x <= 1, `drawer should be anchored to the left edge, got x=${box.x}`);
	assert.ok(box.y <= 1, `drawer should start at the top edge, got y=${box.y}`);
	assert.ok(box.height >= viewport.height - 2, `drawer should span the viewport height, got ${box.height}`);
	assert.ok(box.width >= 260, `drawer should be sidebar width, got ${box.width}`);
	assert.ok(box.width <= Math.ceil(viewport.width * 0.9), `drawer should fit viewport width, got ${box.width}`);
}

async function waitForDrawerState(drawer, state) {
	await drawer.waitFor({ state: 'attached', timeout: 5000 });
	await drawer.evaluate(
		(element, expectedState) =>
			new Promise((resolve, reject) => {
				if (element.getAttribute('data-state') === expectedState) {
					resolve(true);
					return;
				}
				const timeout = setTimeout(() => {
					observer.disconnect();
					reject(new Error(`drawer did not reach ${expectedState}`));
				}, 5000);
				const observer = new MutationObserver(() => {
					if (element.getAttribute('data-state') === expectedState) {
						clearTimeout(timeout);
						observer.disconnect();
						resolve(true);
					}
				});
				observer.observe(element, { attributes: true, attributeFilter: ['data-state'] });
			}),
		state
	);
}

async function assertNoHorizontalOverflow(page) {
	const overflow = await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth + 1);
	assert.equal(overflow, false);
}

async function assertSurfaceShift(page, surface, beforeX, viewport, shouldShift) {
	let box = await surface.boundingBox();
	assert.ok(box, 'app surface should have a bounding box');
	const startedAt = Date.now();
	while (Date.now() - startedAt < 1200) {
		const shift = Math.round(box.x - beforeX);
		if ((shouldShift && shift >= 300) || (!shouldShift && Math.abs(shift) <= 1)) break;
		await page.waitForTimeout(40);
		box = await surface.boundingBox();
		assert.ok(box, 'app surface should have a bounding box');
	}
	const shift = Math.round(box.x - beforeX);
	if (shouldShift) {
		assert.ok(shift >= 300, `desktop/tablet app surface should shift right, got ${shift}`);
	} else {
		assert.ok(Math.abs(shift) <= 1, `mobile app surface should not shift, got ${shift}`);
	}
	await assertNoHorizontalOverflow(page);
	return box.x;
}

async function verifyMediaMenu(page, baseURL, viewport) {
	await page.goto(`${baseURL}/`, { waitUntil: 'domcontentloaded' });
	await page.waitForLoadState('networkidle', { timeout: 10000 });
	await assertNoPersistentMediaPills(page);
	await assertNoHorizontalOverflow(page);
	const shouldShift = viewport.width >= 768;
	const surface = page.getByTestId('lorivo-app-surface');
	const beforeSurface = await surface.boundingBox();
	assert.ok(beforeSurface, 'media app surface should exist');

	const mediaButton = page.getByTestId('media-menu-button');
	assert.equal(await mediaButton.count(), 1);
	await mediaButton.click();
	const mediaDrawer = page.getByTestId('media-menu-drawer');
	await waitForDrawerState(mediaDrawer, 'open');
	await assertLeftDrawer(page, mediaDrawer, viewport);
	await assertSurfaceShift(page, surface, beforeSurface.x, viewport, shouldShift);
	for (const label of ['Home', 'Movies', 'TV', 'Libraries', 'Settings']) {
		assert.equal(await mediaDrawer.getByRole('link', { name: label, exact: true }).count(), 1);
	}
	await page.mouse.click(viewport.width - 12, 12);
	await waitForDrawerState(mediaDrawer, 'closed');
	await assertSurfaceShift(page, surface, beforeSurface.x, viewport, false);

	await mediaButton.click();
	await waitForDrawerState(mediaDrawer, 'open');
	await page.keyboard.press('Escape');
	await waitForDrawerState(mediaDrawer, 'closed');
	await assertSurfaceShift(page, surface, beforeSurface.x, viewport, false);

	await mediaButton.click();
	await waitForDrawerState(mediaDrawer, 'open');
	await assertLeftDrawer(page, mediaDrawer, viewport);
	await mediaDrawer.getByRole('link', { name: 'Movies', exact: true }).click();
	await page.waitForURL(`${baseURL}/movies`, { timeout: 10000 });
	await waitForDrawerState(mediaDrawer, 'closed');
}

async function verifySettingsMenu(page, baseURL, viewport) {
	await page.goto(`${baseURL}/settings`, { waitUntil: 'domcontentloaded' });
	await page.waitForLoadState('networkidle', { timeout: 10000 });
	await assertNoHorizontalOverflow(page);
	const shouldShift = viewport.width >= 768;
	const surface = page.getByTestId('settings-shell-surface');
	const beforeSurface = await surface.boundingBox();
	assert.ok(beforeSurface, 'settings app surface should exist');
	const settingsButton = page.getByTestId('settings-menu-button');
	assert.equal(await settingsButton.count(), 1);
	await settingsButton.click();
	const settingsDrawer = page.getByTestId('settings-menu-drawer');
	await waitForDrawerState(settingsDrawer, 'open');
	await assertLeftDrawer(page, settingsDrawer, viewport);
	await assertSurfaceShift(page, surface, beforeSurface.x, viewport, shouldShift);
	for (const label of ['Library', 'Scanning', 'Metadata', 'Playback', 'Server', 'About', 'Back to Media']) {
		assert.equal(await settingsDrawer.getByRole('link', { name: label, exact: true }).count(), 1);
	}
	await settingsDrawer.getByRole('link', { name: 'Back to Media', exact: true }).click();
	await page.waitForURL(`${baseURL}/`, { timeout: 10000 });
	assert.equal(await page.getByTestId('settings-menu-drawer').count(), 0);
}

test('hamburger media and settings menus open and navigate across viewports', async () => {
	const { server, baseURL } = await launchDevServer();
	const browser = await chromium.launch();

	try {
		const viewports = [
			{ width: 1600, height: 1000, isMobile: false },
			{ width: 1024, height: 768, isMobile: false },
			{ width: 390, height: 844, isMobile: true }
		];

		for (const viewport of viewports) {
			const page = await browser.newPage({
				viewport: { width: viewport.width, height: viewport.height },
				isMobile: viewport.isMobile,
				hasTouch: viewport.isMobile
			});
			await installApiMocks(page);
			await verifyMediaMenu(page, baseURL, viewport);
			await verifySettingsMenu(page, baseURL, viewport);
			await page.close();
		}

		const page = await browser.newPage({ viewport: { width: 390, height: 844 }, isMobile: true, hasTouch: true });
		await installApiMocks(page);
		await page.goto(`${baseURL}/admin`, { waitUntil: 'domcontentloaded' });
		await page.waitForURL(`${baseURL}/settings#server`, { timeout: 10000 });
		assert.doesNotMatch(await page.locator('body').innerText(), /Admin/);
		await page.close();
	} finally {
		await browser.close();
		await server.close();
	}
});
