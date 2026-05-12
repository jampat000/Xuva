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

async function waitForCondition(check, message, timeout = 5000) {
	const startedAt = Date.now();
	while (Date.now() - startedAt < timeout) {
		if (check()) return;
		await new Promise((resolve) => setTimeout(resolve, 40));
	}
	throw new Error(message);
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

function playbackPolicyInfo(id) {
	const policy = id === 'light' || id === 'full' || id === 'cinema' ? id : 'original_only';
	const labels = {
		original_only: 'Original files only',
		light: 'Direct play with audio fixes',
		full: 'Compatibility preferred',
		cinema: 'Broadest device support'
	};
	const descriptions = {
		original_only:
			'Keep playback as close to the original file as possible. If a device needs help, Lorivo offers fallback choices instead of converting automatically.',
		light: 'Prefer the original video. Lorivo can repackage playback or convert audio when that is enough.',
		full: 'Allow temporary video conversion while playing when a device needs more help.',
		cinema: 'Allow heavier live compatibility work for the widest range of devices.'
	};
	return { id: policy, label: labels[policy], description: descriptions[policy] };
}

function apiPayload(pathname, state = {}) {
	if (pathname === '/api/auth/session') {
		if (state.authDisabled) return { authDisabled: true };
		if (state.devAuthBypass) {
			return {
				devAuthBypass: true,
				devAuthBypassMessage:
					'Development access is active. User management will be enabled before production.',
				user: {
					id: 'dev-owner',
					username: 'development-owner',
					displayName: 'Development Owner',
					role: 'admin'
				},
				session: { id: 'dev-auth-bypass' }
			};
		}
		if (!state.signedIn) return { user: null };
		return {
			user: { id: 'local', username: 'local', displayName: 'Local User', role: 'admin' },
			session: { id: 'session-local', expiresAt: '2026-05-12T21:15:00Z' }
		};
	}
	if (pathname === '/api/client/bootstrap') {
		return {
			auth: {
				required: !state.authDisabled,
				devAuthBypass: Boolean(state.devAuthBypass),
				bootstrapAllowed: Boolean(state.bootstrapAllowed),
				defaultUsername: state.defaultUsername || 'owner',
				bootstrapEndpoint: '/api/auth/bootstrap'
			},
			server: {
				name: state.serverName
			}
		};
	}
	if (pathname === '/api/client/home') return { rows: [], actions: {} };
	if (pathname === '/api/playback/recent') return { recent: [] };
	if (pathname === '/api/libraries') return { libraries: state.libraries || [] };
	if (pathname === '/api/movies') return { movies: [] };
	if (pathname === '/api/series') return { series: [] };
	if (pathname === '/api/catalog/summary') return { movies: 42, series: 6, episodes: 48 };
	if (pathname === '/api/catalog/health') return {};
	if (pathname === '/api/system/status') return { cpu: {}, memory: {} };
	if (pathname === '/api/settings') {
		return {
			config: {
				serverName: state.serverName,
				librarySyncMode: state.librarySyncMode,
				syncIntervalMins: state.syncIntervalMins,
				watchDebounceSecs: state.watchDebounceSecs,
				probeBatchLimit: state.probeBatchLimit,
				playbackPolicy: state.playbackPolicy,
				metadataProviders: { automatic: [], managedOverrides: [] }
			},
			libraries: state.libraries || []
		};
	}
	if (pathname === '/api/settings/performance') {
		return {
			playbackPolicy: playbackPolicyInfo(state.playbackPolicy),
			hardwareAcceleration: { status: 'Available' }
		};
	}
	if (pathname === '/api/scans') return { scans: state.scans || [] };
	if (pathname === '/api/probes') return { probes: [] };
	if (pathname === '/api/work') return { work: [] };
	if (pathname === '/api/downloads') return { downloads: [] };
	if (pathname === '/api/sessions') {
		if (!state.signedIn && !state.authDisabled && !state.devAuthBypass) return { error: 'authentication required' };
		return { sessions: state.sessions || [] };
	}
	return {};
}

async function installApiMocks(page, options = {}) {
	const state = {
		serverName: Object.hasOwn(options, 'serverName') ? options.serverName : 'Living Room Lorivo',
		librarySyncMode: 'daily',
		syncIntervalMins: 1440,
		watchDebounceSecs: 30,
		probeBatchLimit: 50,
		playbackPolicy: 'original_only',
		restartRequired: true,
		signedIn: !options.signedOut,
		devAuthBypass: Boolean(options.devAuthBypass),
		authDisabled: Boolean(options.authDisabled),
		bootstrapAllowed: Boolean(options.bootstrapAllowed),
		defaultUsername: 'owner',
		libraries: [
			{ id: 'movies-main', name: 'Movies', kind: 'movies', path: 'D:\\Media\\Movies', storageType: 'local' },
			{ id: 'tv-main', name: 'TV', kind: 'tv', path: 'D:\\Media\\TV', storageType: 'network' }
		],
		scans: [
			{ id: 'scan-tv-main', status: 'completed', libraryId: 'tv-main', updatedAt: '2026-05-12T18:20:00Z' }
		],
		sessions: [
			{
				id: 'session-1',
				title: 'Heat',
				sourceName: 'Heat (1995)',
				deviceId: 'living-room-tv',
				mode: 'direct',
				route: 'direct'
			}
		],
		playbackUpdates: [],
		scanningUpdates: [],
		movieScanCount: 0,
		tvScanCount: 0,
		deletedLibraries: [],
		scannedLibraries: [],
		logoutCount: 0
	};
	await page.route('**/api/**', (route) => {
		const url = new URL(route.request().url());
		if (!url.pathname.startsWith('/api/')) return route.continue();
		if (url.pathname === '/api/events') return route.fulfill({ status: 204, body: '' });
		if (url.pathname === '/api/libraries/movies/scan' && route.request().method() === 'POST') {
			state.movieScanCount += 1;
			return route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({ id: `scan-movies-${state.movieScanCount}`, status: 'queued', kind: 'movie' })
			});
		}
		if (url.pathname === '/api/libraries/tv/scan' && route.request().method() === 'POST') {
			state.tvScanCount += 1;
			return route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({ id: `scan-tv-${state.tvScanCount}`, status: 'queued', kind: 'series' })
			});
		}
		if (url.pathname === '/api/settings' && route.request().method() === 'PUT') {
			const body = route.request().postDataJSON() || {};
			const nextName = String(body.serverName ?? '').trim();
			if (Object.hasOwn(body, 'serverName')) {
				if (!nextName) {
					return route.fulfill({
						status: 400,
						contentType: 'application/json',
						body: JSON.stringify({ error: 'server name is required' })
					});
				}
				if ([...nextName].length > 50) {
					return route.fulfill({
						status: 400,
						contentType: 'application/json',
						body: JSON.stringify({ error: 'server name must be 50 characters or fewer' })
					});
				}
				state.serverName = nextName;
			}
			if (Object.hasOwn(body, 'playbackPolicy')) {
				const nextPolicy = String(body.playbackPolicy || '').trim();
				if (!['original_only', 'light', 'full', 'cinema'].includes(nextPolicy)) {
					return route.fulfill({
						status: 400,
						contentType: 'application/json',
						body: JSON.stringify({ error: 'invalid playback policy' })
					});
				}
				state.playbackPolicy = nextPolicy;
				state.playbackUpdates.push(body);
			}
			if (
				Object.hasOwn(body, 'librarySyncMode') ||
				Object.hasOwn(body, 'syncIntervalMins') ||
				Object.hasOwn(body, 'watchDebounceSecs') ||
				Object.hasOwn(body, 'probeBatchLimit')
			) {
				const nextMode = String(body.librarySyncMode || state.librarySyncMode).trim();
				if (!['manual', 'daily', 'watch'].includes(nextMode)) {
					return route.fulfill({
						status: 400,
						contentType: 'application/json',
						body: JSON.stringify({ error: 'invalid library scan mode' })
					});
				}
				const nextInterval = Number(body.syncIntervalMins ?? state.syncIntervalMins);
				const nextWatchDelay = Number(body.watchDebounceSecs ?? state.watchDebounceSecs);
				const nextProbeBatchLimit = Number(body.probeBatchLimit ?? state.probeBatchLimit);
				if (nextMode === 'daily' && (!Number.isFinite(nextInterval) || nextInterval < 15)) {
					return route.fulfill({
						status: 400,
						contentType: 'application/json',
						body: JSON.stringify({ error: 'scan interval must be at least 15 minutes' })
					});
				}
				if (nextMode === 'watch' && (!Number.isFinite(nextWatchDelay) || nextWatchDelay < 5 || nextWatchDelay > 300)) {
					return route.fulfill({
						status: 400,
						contentType: 'application/json',
						body: JSON.stringify({ error: 'folder watch delay must be between 5 and 300 seconds' })
					});
				}
				if (!Number.isFinite(nextProbeBatchLimit) || nextProbeBatchLimit <= 0) {
					return route.fulfill({
						status: 400,
						contentType: 'application/json',
						body: JSON.stringify({ error: 'media check batch size must be greater than 0' })
					});
				}
				state.librarySyncMode = nextMode;
				state.syncIntervalMins = nextInterval;
				state.watchDebounceSecs = nextWatchDelay;
				state.probeBatchLimit = nextProbeBatchLimit;
				state.scanningUpdates.push({
					librarySyncMode: state.librarySyncMode,
					syncIntervalMins: state.syncIntervalMins,
					watchDebounceSecs: state.watchDebounceSecs,
					probeBatchLimit: state.probeBatchLimit
				});
			}
			return route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({
					...apiPayload('/api/settings', state),
					restartRequired: state.restartRequired
				})
			});
		}
		if (url.pathname.startsWith('/api/libraries/') && url.pathname.endsWith('/scan') && route.request().method() === 'POST') {
			const libraryID = url.pathname.split('/')[3];
			state.scannedLibraries.push(libraryID);
			state.scans = [
				{
					id: `scan-${libraryID}`,
					status: 'queued',
					libraryId: libraryID,
					updatedAt: '2026-05-12T19:30:00Z'
				},
				...state.scans.filter((item) => item.libraryId !== libraryID)
			];
			return route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({ id: `scan-${libraryID}`, status: 'queued', libraryId: libraryID })
			});
		}
		if (url.pathname.startsWith('/api/libraries/') && route.request().method() === 'DELETE') {
			const libraryID = url.pathname.split('/')[3];
			state.deletedLibraries.push(libraryID);
			state.libraries = state.libraries.filter((library) => library.id !== libraryID);
			state.scans = state.scans.filter((item) => item.libraryId !== libraryID);
			return route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({ id: libraryID })
			});
		}
		if (url.pathname === '/api/auth/logout' && route.request().method() === 'POST') {
			if (state.devAuthBypass) {
				state.logoutCount += 1;
				return route.fulfill({
					status: 200,
					contentType: 'application/json',
					body: JSON.stringify({
						status: 'dev_bypass_active',
						devAuthBypass: true,
						devAuthBypassMessage:
							'Development access bypass remains active until LORIVO_DEV_AUTH_BYPASS is turned off.'
					})
				});
			}
			state.signedIn = false;
			state.logoutCount += 1;
			return route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({ status: 'logged_out' })
			});
		}
		if (url.pathname === '/api/sessions' && !state.signedIn && !state.authDisabled && !state.devAuthBypass) {
			return route.fulfill({
				status: 401,
				contentType: 'application/json',
				body: JSON.stringify({ error: 'authentication required' })
			});
		}
		return route.fulfill({
			status: 200,
			contentType: 'application/json',
			body: JSON.stringify(apiPayload(url.pathname, state))
		});
	});
	await page.route('**/build-info.json', (route) =>
		route.fulfill({
			status: 200,
			contentType: 'application/json',
			body: JSON.stringify({ buildID: 'test', gitCommit: 'test', sourceApp: 'apps/web/svelte' })
		})
	);
	return state;
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

async function readTopbarBrandState(page) {
	const brand = page.getByTestId('topbar-brand');
	assert.equal(await brand.count(), 1);
	return brand.evaluate((element) => {
		const style = window.getComputedStyle(element);
		const rect = element.getBoundingClientRect();
		return {
			ariaHidden: element.getAttribute('aria-hidden'),
			opacity: Number(style.opacity),
			width: rect.width,
			text: element.textContent?.trim() || ''
		};
	});
}

async function assertTopbarBrandVisible(page) {
	const startedAt = Date.now();
	let state = await readTopbarBrandState(page);
	while (Date.now() - startedAt < 1200) {
		if (state.ariaHidden !== 'true' && state.opacity > 0.8 && state.width > 60) break;
		await page.waitForTimeout(40);
		state = await readTopbarBrandState(page);
	}
	assert.equal(state.ariaHidden === 'true', false);
	assert.ok(state.opacity > 0.8, `top bar brand should be visible, got opacity ${state.opacity}`);
	assert.ok(state.width > 60, `top bar brand should occupy space, got width ${state.width}`);
	assert.match(state.text, /LORIVO/i);
}

async function assertTopbarBrandCollapsed(page) {
	const startedAt = Date.now();
	let state = await readTopbarBrandState(page);
	while (Date.now() - startedAt < 1200) {
		if (state.ariaHidden === 'true' && (state.opacity < 0.25 || state.width < 12)) break;
		await page.waitForTimeout(40);
		state = await readTopbarBrandState(page);
	}
	assert.equal(state.ariaHidden, 'true');
	assert.ok(
		state.opacity < 0.25 || state.width < 12,
		`top bar brand should collapse while drawer brand is visible, got opacity ${state.opacity} and width ${state.width}`
	);
}

async function assertSurfaceLayout(page, surface, beforeBox, viewport, shouldShift) {
	let box = await surface.boundingBox();
	assert.ok(box, 'app surface should have a bounding box');
	const startedAt = Date.now();
	while (Date.now() - startedAt < 1200) {
		const shift = Math.round(box.x - beforeBox.x);
		const rightEdge = box.x + box.width;
		const fitsViewport = rightEdge <= viewport.width + 2;
		if ((shouldShift && shift >= 300 && fitsViewport) || (!shouldShift && Math.abs(shift) <= 1)) break;
		await page.waitForTimeout(40);
		box = await surface.boundingBox();
		assert.ok(box, 'app surface should have a bounding box');
	}
	const shift = Math.round(box.x - beforeBox.x);
	const rightEdge = box.x + box.width;
	if (shouldShift) {
		assert.ok(shift >= 300, `desktop/tablet app surface should shift right, got ${shift}`);
		assert.ok(
			rightEdge <= viewport.width + 2,
			`desktop/tablet app surface should fit within viewport after shifting, got right edge ${rightEdge} in ${viewport.width}px viewport`
		);
		assert.ok(
			box.width <= beforeBox.width - 300 + 4,
			`desktop/tablet app surface should resize to remaining width, got ${box.width} from ${beforeBox.width}`
		);
	} else {
		assert.ok(Math.abs(shift) <= 1, `mobile app surface should not shift, got ${shift}`);
		assert.ok(
			box.width >= beforeBox.width - 2,
			`mobile/closed app surface should keep its full width, got ${box.width} from ${beforeBox.width}`
		);
		assert.ok(
			rightEdge <= viewport.width + 2,
			`mobile/closed app surface should fit within viewport, got right edge ${rightEdge} in ${viewport.width}px viewport`
		);
	}
	await assertNoHorizontalOverflow(page);
	return box.x;
}

async function verifyMediaMenu(page, baseURL, viewport) {
	await page.goto(`${baseURL}/`, { waitUntil: 'domcontentloaded' });
	await page.waitForLoadState('networkidle', { timeout: 10000 });
	await assertNoPersistentMediaPills(page);
	await assertNoHorizontalOverflow(page);
	await assertTopbarBrandVisible(page);
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
	const drawerBrand = mediaDrawer.getByTestId('drawer-brand');
	assert.equal(await drawerBrand.count(), 1);
	assert.match(await drawerBrand.innerText(), /LORIVO/i);
	await assertTopbarBrandCollapsed(page);
	await assertSurfaceLayout(page, surface, beforeSurface, viewport, shouldShift);
	for (const label of ['Home', 'Movies', 'TV', 'Libraries', 'Settings']) {
		assert.equal(await mediaDrawer.getByRole('link', { name: label, exact: true }).count(), 1);
	}
	await page.mouse.click(viewport.width - 12, 12);
	await waitForDrawerState(mediaDrawer, 'closed');
	await assertTopbarBrandVisible(page);
	await assertSurfaceLayout(page, surface, beforeSurface, viewport, false);

	await mediaButton.click();
	await waitForDrawerState(mediaDrawer, 'open');
	await page.keyboard.press('Escape');
	await waitForDrawerState(mediaDrawer, 'closed');
	await assertTopbarBrandVisible(page);
	await assertSurfaceLayout(page, surface, beforeSurface, viewport, false);

	await mediaButton.click();
	await waitForDrawerState(mediaDrawer, 'open');
	await assertLeftDrawer(page, mediaDrawer, viewport);
	await mediaDrawer.getByRole('link', { name: 'Movies', exact: true }).click();
	await page.waitForURL(`${baseURL}/movies`, { timeout: 10000 });
	await waitForDrawerState(mediaDrawer, 'closed');
	await assertNoHorizontalOverflow(page);

	const moviesMediaButton = page.getByTestId('media-menu-button');
	assert.equal(await moviesMediaButton.count(), 1);
	await moviesMediaButton.click();
	await waitForDrawerState(page.getByTestId('media-menu-drawer'), 'open');
	await page.getByTestId('media-menu-drawer').getByRole('link', { name: 'TV', exact: true }).click();
	await page.waitForURL(`${baseURL}/tv`, { timeout: 10000 });
	await waitForDrawerState(page.getByTestId('media-menu-drawer'), 'closed');
	await assertNoHorizontalOverflow(page);
}

async function verifySettingsMenu(page, baseURL, viewport, state) {
	await page.goto(`${baseURL}/settings`, { waitUntil: 'domcontentloaded' });
	await page.waitForLoadState('networkidle', { timeout: 10000 });
	await assertNoHorizontalOverflow(page);
	assert.equal(await page.getByTestId('settings-dashboard').count(), 1);
	assert.equal(await page.getByTestId('settings-section-content').count(), 0);
	assert.equal(await page.title(), 'Living Room Lorivo · Lorivo');
	assert.match(await page.getByTestId('settings-server-name').innerText(), /Living Room Lorivo/);
	assert.match(await page.locator('body').innerText(), /Dashboard/);
	assert.match(await page.locator('body').innerText(), /Library Setup/);
	assert.match(await page.locator('body').innerText(), /Playback/);
	assert.match(await page.locator('body').innerText(), /Access/);
	assert.equal(await page.getByPlaceholder('Search', { exact: true }).count(), 0);
	await assertSettingsSafetyCopy(page);
	if (viewport.width >= 981) {
		const sidebar = page.getByTestId('settings-mode-sidebar');
		assert.equal(await sidebar.count(), 1);
		assert.equal(await sidebar.isVisible(), true);
		assert.match(await sidebar.innerText(), /LORIVO/);
		assert.match(await sidebar.innerText(), /Settings/i);
		for (const label of ['Dashboard', 'Library', 'Scanning', 'Metadata', 'Playback', 'Access', 'About', 'Back to Media']) {
			assert.equal(await sidebar.getByRole('link', { name: label, exact: true }).count(), 1);
		}
		assert.equal(await sidebar.getByRole('link', { name: 'Server', exact: true }).count(), 0);
		assert.equal(await sidebar.getByRole('link', { name: 'Appearance', exact: true }).count(), 0);
		assert.equal(await sidebar.getByRole('link', { name: 'Diagnostics', exact: true }).count(), 0);
		const backToMedia = sidebar.getByRole('link', { name: 'Back to Media', exact: true });
		assert.equal(await backToMedia.count(), 1);
		assert.equal(await backToMedia.evaluate((element) => element.classList.contains('sidebar-item--back')), true);
		await verifySettingsSections(page, sidebar, baseURL, false, state);
		await sidebar.getByRole('link', { name: 'Back to Media', exact: true }).click();
	} else {
		const settingsButton = page.getByTestId('settings-menu-button');
		assert.equal(await settingsButton.count(), 1);
		assert.equal(await settingsButton.isVisible(), true);
		await settingsButton.click();
		const settingsDrawer = page.getByTestId('settings-menu-drawer');
		await waitForDrawerState(settingsDrawer, 'open');
		await assertLeftDrawer(page, settingsDrawer, viewport);
		assert.match(await settingsDrawer.innerText(), /LORIVO/);
		assert.match(await settingsDrawer.innerText(), /Settings/i);
		assert.equal(await settingsDrawer.getByTestId('drawer-brand').count(), 1);
		for (const label of ['Dashboard', 'Library', 'Scanning', 'Metadata', 'Playback', 'Access', 'About', 'Back to Media']) {
			assert.equal(await settingsDrawer.getByRole('link', { name: label, exact: true }).count(), 1);
		}
		assert.equal(await settingsDrawer.getByRole('link', { name: 'Server', exact: true }).count(), 0);
		assert.equal(await settingsDrawer.getByRole('link', { name: 'Appearance', exact: true }).count(), 0);
		assert.equal(await settingsDrawer.getByRole('link', { name: 'Diagnostics', exact: true }).count(), 0);
		const backToMedia = settingsDrawer.getByRole('link', { name: 'Back to Media', exact: true });
		assert.equal(await backToMedia.count(), 1);
		assert.equal(await backToMedia.evaluate((element) => element.classList.contains('app-drawer__link--back')), true);
		await verifySettingsSections(page, settingsDrawer, baseURL, true, state);
		await settingsButton.click();
		await waitForDrawerState(settingsDrawer, 'open');
		await settingsDrawer.getByRole('link', { name: 'Back to Media', exact: true }).click();
	}
	await page.waitForURL(`${baseURL}/`, { timeout: 10000 });
	await assertNoHorizontalOverflow(page);
}

async function verifySettingsSections(page, navContainer, baseURL, reopensDrawer, state) {
	const sections = [
		['Library', 'library'],
		['Scanning', 'scanning'],
		['Metadata', 'metadata'],
		['Playback', 'playback'],
		['About', 'about'],
		['Access', 'access']
	];
	for (let index = 0; index < sections.length; index += 1) {
		const [label, hash] = sections[index];
		const link = navContainer.getByRole('link', { name: label, exact: true });
		assert.equal(await link.count(), 1);
		await link.click();
		await page.waitForURL(`${baseURL}/settings#${hash}`, { timeout: 10000 });
		const selectedSection = page.getByTestId('settings-section-content');
		await selectedSection.waitFor({ state: 'visible', timeout: 5000 });
		await page.waitForFunction(
			(expected) =>
				document.querySelector('[data-testid="settings-section-content"]')?.getAttribute('data-section') === expected,
			hash,
			{ timeout: 5000 }
		);
		assert.equal(await selectedSection.count(), 1);
		assert.equal(await selectedSection.getAttribute('data-section'), hash);
		assert.equal(await page.locator(`section#${hash}`).count(), 1);
		assert.equal(await page.getByTestId('settings-dashboard').count(), 0);
		if (hash === 'library') {
			await assertLibrarySection(page, state);
		}
		if (hash === 'scanning') {
			await assertScanningSection(page, state);
		}
		if (hash === 'playback') {
			await assertPlaybackSection(page, state);
		}
		if (hash === 'about') {
			assert.match(await selectedSection.innerText(), /Server name/);
			const serverNameInput = selectedSection.locator('input[placeholder="Lorivo"]');
			assert.equal(await serverNameInput.count(), 1);
			await serverNameInput.fill('Family Library');
			await selectedSection.getByRole('button', { name: 'Save Server Name', exact: true }).click();
			await page.waitForFunction(() => document.title === 'Family Library · Lorivo', { timeout: 5000 });
			assert.equal(await page.title(), 'Family Library · Lorivo');
			assert.match(await page.getByTestId('settings-server-name').innerText(), /Family Library/);
		}
		if (hash === 'access') {
			await assertAccessSection(page, state);
		}
		await assertSettingsSafetyCopy(page);
		for (const [, otherHash] of sections) {
			if (otherHash === hash) continue;
			assert.equal(await page.locator(`section#${otherHash}`).count(), 0);
		}
		await assertNoHorizontalOverflow(page);
		if (reopensDrawer && index < sections.length - 1) {
			const settingsButton = page.getByTestId('settings-menu-button');
			await waitForDrawerState(page.getByTestId('settings-menu-drawer'), 'closed');
			await settingsButton.click();
			await waitForDrawerState(page.getByTestId('settings-menu-drawer'), 'open');
		}
	}
	if (state.logoutCount > 0) {
		await verifySignedOutLibraryState(page, baseURL, reopensDrawer);
	}
}

async function assertSettingsSafetyCopy(page) {
	const body = await page.locator('body').innerText();
	assert.doesNotMatch(body, /\bAdmin\b/i);
	assert.doesNotMatch(body, /\bOperator\b/i);
	assert.doesNotMatch(body, /Operational telemetry/i);
	assert.doesNotMatch(body, /Manage Server/i);
	assert.doesNotMatch(body, /Write Controls/i);
	assert.doesNotMatch(body, /\bAPI\b/i);
	assert.doesNotMatch(body, /endpoint/i);
	assert.doesNotMatch(body, /provider internals/i);
	assert.doesNotMatch(body, /Transcode Workers/i);
}

async function assertLibrarySection(page, state) {
	const selectedSection = page.getByTestId('settings-section-content');
	assert.equal(await selectedSection.getByRole('link', { name: 'Library Setup', exact: true }).count(), 1);
	assert.equal(await page.getByTestId('library-list').count(), 1);
	assert.equal(await page.getByTestId('library-item').count(), state.libraries.length);
	assert.match(await selectedSection.innerText(), /Movies library/);
	assert.match(await selectedSection.innerText(), /TV library/);
	assert.match(await selectedSection.innerText(), /Folder/);
	assert.equal(await selectedSection.getByRole('button', { name: 'Edit', exact: true }).count(), 0);
	assert.equal(await selectedSection.getByRole('button', { name: 'Scan', exact: true }).count(), state.libraries.length);
	assert.equal(await selectedSection.getByRole('button', { name: 'Remove', exact: true }).count(), state.libraries.length);

	const firstLibrary = page.getByTestId('library-item').first();
	await firstLibrary.getByRole('button', { name: 'Scan', exact: true }).click();
	await waitForCondition(
		() => state.scannedLibraries.includes('movies-main'),
		'expected library scan action to be recorded'
	);
	await page.waitForFunction(
		() => document.body.innerText.includes('Movies scan started.'),
		null,
		{ timeout: 5000 }
	);
	assert.match(await page.locator('body').innerText(), /Movies scan started\./);

	await Promise.all([
		page.waitForEvent('dialog').then(async (dialog) => {
			assert.match(dialog.message(), /Media files are not deleted\./);
			await dialog.accept();
		}),
		firstLibrary.getByRole('button', { name: 'Remove', exact: true }).click()
	]);
	await waitForCondition(
		() => state.deletedLibraries.includes('movies-main'),
		'expected library remove action to be recorded'
	);
	await page.waitForFunction(
		() => document.querySelectorAll('[data-testid="library-item"]').length === 1,
		null,
		{ timeout: 5000 }
	);
	assert.equal(await page.getByTestId('library-item').count(), 1);
	assert.match(await selectedSection.innerText(), /TV library/);
	assert.doesNotMatch(await selectedSection.innerText(), /D:\\Media\\Movies/);
	assert.equal(await selectedSection.getByRole('button', { name: 'Scan', exact: true }).count(), 1);
	assert.equal(await selectedSection.getByRole('button', { name: 'Remove', exact: true }).count(), 1);
	await assertNoHorizontalOverflow(page);
}

async function assertScanningSection(page, state) {
	const selectedSection = page.getByTestId('settings-section-content');
	assert.match(await selectedSection.innerText(), /Manual scans/);
	assert.match(await selectedSection.innerText(), /Automation/);
	assert.match(await selectedSection.innerText(), /Library scan mode/);
	assert.match(await selectedSection.innerText(), /Scheduled/);
	assert.match(await selectedSection.innerText(), /Scan interval/);
	assert.match(await selectedSection.innerText(), /Media check batch size/);
	assert.equal(await selectedSection.getByRole('button', { name: 'Scan Movies', exact: true }).count(), 1);
	assert.equal(await selectedSection.getByRole('button', { name: 'Scan TV', exact: true }).count(), 1);
	assert.equal(await page.getByTestId('scanning-automation-form').count(), 1);
	await selectedSection.getByRole('button', { name: 'Scan Movies', exact: true }).click();
	await waitForCondition(() => state.movieScanCount === 1, 'expected movie scan action to be recorded');
	await page.waitForFunction(() => document.body.innerText.includes('Movie scan started.'), null, { timeout: 5000 });
	await selectedSection.getByRole('button', { name: 'Scan TV', exact: true }).click();
	await waitForCondition(() => state.tvScanCount === 1, 'expected TV scan action to be recorded');
	await page.waitForFunction(() => document.body.innerText.includes('TV scan started.'), null, { timeout: 5000 });

	const scanIntervalInput = selectedSection.locator('input[placeholder="1440"]');
	assert.equal(await scanIntervalInput.count(), 1);
	await scanIntervalInput.fill('0');
	await selectedSection.getByRole('button', { name: 'Save Scanning Settings', exact: true }).click();
	assert.match(await selectedSection.innerText(), /Enter a scan interval of at least 15 minutes\./);
	assert.equal(state.scanningUpdates.length, 0);

	await selectedSection.locator('input[value="watch"]').check();
	await selectedSection.locator('input[placeholder="30"]').fill('45');
	const advancedToggle = selectedSection.locator('details[data-testid="scanning-advanced"]');
	await advancedToggle.locator('summary').click();
	await selectedSection.locator('input[placeholder="50"]').fill('75');
	await selectedSection.getByRole('button', { name: 'Save Scanning Settings', exact: true }).click();
	await waitForCondition(
		() =>
			state.scanningUpdates.some(
				(item) =>
					item.librarySyncMode === 'watch' &&
					item.watchDebounceSecs === 45 &&
					item.probeBatchLimit === 75
			),
		'expected watch-mode scanning settings update to be recorded'
	);
	await page.waitForFunction(
		() => document.body.innerText.includes('Saved. Restart Lorivo for this change to fully take effect.'),
		null,
		{ timeout: 5000 }
	);
	assert.match(await page.locator('body').innerText(), /Saved\. Restart Lorivo for this change to fully take effect\./);

	state.restartRequired = false;
	await selectedSection.locator('input[value="daily"]').check();
	await selectedSection.locator('input[placeholder="1440"]').fill('120');
	await selectedSection.getByRole('button', { name: 'Save Scanning Settings', exact: true }).click();
	await waitForCondition(
		() =>
			state.scanningUpdates.some(
				(item) =>
					item.librarySyncMode === 'daily' &&
					item.syncIntervalMins === 120 &&
					item.watchDebounceSecs === 45 &&
					item.probeBatchLimit === 75
			),
		'expected scheduled scanning settings update to be recorded'
	);
	await page.waitForFunction(
		() =>
			document.body.innerText.includes('Scanning settings saved.') &&
			!document.body.innerText.includes('Saved. Restart Lorivo for this change to fully take effect.'),
		null,
		{ timeout: 5000 }
	);
	const body = await page.locator('body').innerText();
	assert.match(body, /Scanning settings saved\./);
	assert.doesNotMatch(body, /\bLibrarySyncMode\b|\bSyncIntervalMins\b|\bWatchDebounceSecs\b|\bProbeBatchLimit\b/);
	await assertNoHorizontalOverflow(page);
}

async function assertPlaybackSection(page, state) {
	const selectedSection = page.getByTestId('settings-section-content');
	state.restartRequired = true;
	assert.equal(await page.getByTestId('playback-policy-form').count(), 1);
	assert.match(await selectedSection.innerText(), /Original files only/);
	await selectedSection.locator('input[value="full"]').check();
	await selectedSection.getByRole('button', { name: 'Save Playback Policy', exact: true }).click();
	await waitForCondition(
		() => state.playbackUpdates.some((item) => item.playbackPolicy === 'full'),
		'expected playback policy update to be recorded'
	);
	await page.waitForFunction(
		() => document.body.innerText.includes('Playback setting saved. Restart Lorivo to apply it.'),
		null,
		{ timeout: 5000 }
	);
	assert.match(await page.locator('body').innerText(), /Playback setting saved\. Restart Lorivo to apply it\./);

	state.restartRequired = false;
	await selectedSection.locator('input[value="cinema"]').check();
	await selectedSection.getByRole('button', { name: 'Save Playback Policy', exact: true }).click();
	await waitForCondition(
		() => state.playbackUpdates.some((item) => item.playbackPolicy === 'cinema'),
		'expected second playback policy update to be recorded'
	);
	await page.waitForFunction(
		() =>
			document.body.innerText.includes('Playback setting saved.') &&
			!document.body.innerText.includes('Playback setting saved. Restart Lorivo to apply it.'),
		null,
		{ timeout: 5000 }
	);
	const body = await page.locator('body').innerText();
	assert.match(body, /Playback setting saved\./);
	assert.doesNotMatch(body, /Playback setting saved\. Restart Lorivo to apply it\./);
	await assertNoHorizontalOverflow(page);
}

async function assertAccessSection(page, state) {
	const selectedSection = page.getByTestId('settings-section-content');
	if (state.devAuthBypass) {
		assert.match(await selectedSection.innerText(), /Development Owner/);
		assert.match(
			await selectedSection.innerText(),
			/Development access is active\. User management will be enabled before production\./
		);
		assert.match(await selectedSection.innerText(), /User management is not available yet\./);
		assert.match(
			await selectedSection.innerText(),
			/Device pairing will appear here when client pairing is implemented\./
		);
		assert.equal(await selectedSection.getByRole('button', { name: 'Sign Out', exact: true }).count(), 0);
		assert.equal(await selectedSection.getByRole('link', { name: 'Sign In', exact: true }).count(), 0);
		assert.equal(await selectedSection.getByRole('link', { name: 'Create Owner Account', exact: true }).count(), 0);
		await assertNoHorizontalOverflow(page);
		return;
	}
	assert.match(await selectedSection.innerText(), /Local User/);
	assert.match(await selectedSection.innerText(), /Owner account/);
	assert.equal(await selectedSection.getByRole('button', { name: 'Sign Out', exact: true }).count(), 1);
	assert.equal(await selectedSection.getByRole('link', { name: 'Sign In', exact: true }).count(), 0);
	await selectedSection.getByRole('button', { name: 'Sign Out', exact: true }).click();
	await waitForCondition(() => state.logoutCount === 1, 'expected logout to be recorded');
	await selectedSection.getByRole('link', { name: 'Sign In', exact: true }).waitFor({ state: 'visible', timeout: 5000 });
	assert.equal(await selectedSection.getByRole('button', { name: 'Sign Out', exact: true }).count(), 0);
	assert.match(await selectedSection.innerText(), /Sign in as the owner to manage Lorivo settings\./);
	assert.doesNotMatch(await selectedSection.innerText(), /Create User|Delete User/i);
	await assertNoHorizontalOverflow(page);
}

async function verifySignedOutLibraryState(page, baseURL, reopensDrawer) {
	const navContainer = reopensDrawer ? page.getByTestId('settings-menu-drawer') : page.getByTestId('settings-mode-sidebar');
	if (reopensDrawer) {
		const settingsButton = page.getByTestId('settings-menu-button');
		await settingsButton.click();
		await waitForDrawerState(navContainer, 'open');
	}
	await navContainer.getByRole('link', { name: 'Library', exact: true }).click();
	await page.waitForURL(`${baseURL}/settings#library`, { timeout: 10000 });
	await page.waitForFunction(
		() => document.querySelector('[data-testid="settings-section-content"]')?.getAttribute('data-section') === 'library',
		null,
		{ timeout: 5000 }
	);
	const selectedSection = page.getByTestId('settings-section-content');
	await selectedSection.waitFor({ state: 'visible', timeout: 5000 });
	assert.equal(await selectedSection.getByRole('button', { name: 'Scan', exact: true }).count(), 0);
	assert.equal(await selectedSection.getByRole('button', { name: 'Remove', exact: true }).count(), 0);
	assert.match(await selectedSection.innerText(), /Sign in as the owner to manage Lorivo settings\./);
	assert.equal(await selectedSection.getByRole('link', { name: 'Sign In', exact: true }).count(), 1);
	await assertNoHorizontalOverflow(page);
}

async function verifySetupBelongsToSettingsMode(page, baseURL, viewport) {
	await page.goto(`${baseURL}/setup`, { waitUntil: 'domcontentloaded' });
	await page.waitForLoadState('networkidle', { timeout: 10000 });
	assert.match(await page.title(), /(?:Living Room Lorivo|Family Library) · Lorivo/);
	assert.match(await page.locator('body').innerText(), /Library Setup/);
	assert.equal(await page.getByPlaceholder('Search', { exact: true }).count(), 0);
	const setupServerName = page.locator('input[placeholder="Lorivo"]');
	assert.equal(await setupServerName.count(), 1);
	assert.match(await page.locator('body').innerText(), /Lorivo uses this name in the browser title and shares it with local clients\./);
	assert.match(await page.locator('body').innerText(), /Local network discovery is not available in this build yet\./);
	await setupServerName.fill('   ');
	await page.getByRole('button', { name: 'Save library', exact: true }).click();
	assert.match(await page.locator('body').innerText(), /Enter a server name\./);
	await setupServerName.fill('Living Room Lorivo');
	if (viewport.width >= 981) {
		const sidebar = page.getByTestId('settings-mode-sidebar');
		assert.equal(await sidebar.count(), 1);
		assert.equal(await sidebar.isVisible(), true);
		assert.equal(await sidebar.getByRole('link', { name: 'Library', exact: true }).count(), 1);
		assert.equal(await sidebar.getByRole('link', { name: 'Access', exact: true }).count(), 1);
		assert.equal(await sidebar.getByRole('link', { name: 'Back to Media', exact: true }).count(), 1);
		assert.equal(await sidebar.getByRole('link', { name: 'Server', exact: true }).count(), 0);
	} else {
		const settingsButton = page.getByTestId('settings-menu-button');
		assert.equal(await settingsButton.count(), 1);
		assert.equal(await settingsButton.isVisible(), true);
		await settingsButton.click();
		const settingsDrawer = page.getByTestId('settings-menu-drawer');
		await waitForDrawerState(settingsDrawer, 'open');
		assert.equal(await settingsDrawer.getByRole('link', { name: 'Access', exact: true }).count(), 1);
		assert.equal(await settingsDrawer.getByRole('link', { name: 'Back to Media', exact: true }).count(), 1);
		assert.equal(await settingsDrawer.getByRole('link', { name: 'Server', exact: true }).count(), 0);
		await settingsDrawer.getByRole('link', { name: 'Back to Media', exact: true }).click();
		await page.waitForURL(`${baseURL}/`, { timeout: 10000 });
	}
	await assertNoHorizontalOverflow(page);
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
			const state = await installApiMocks(page);
			await verifyMediaMenu(page, baseURL, viewport);
			await verifySetupBelongsToSettingsMode(page, baseURL, viewport);
			await verifySettingsMenu(page, baseURL, viewport, state);
			await page.close();
		}

		const fallbackPage = await browser.newPage({ viewport: { width: 390, height: 844 }, isMobile: true, hasTouch: true });
		await installApiMocks(fallbackPage, { serverName: '' });
		await fallbackPage.goto(`${baseURL}/settings`, { waitUntil: 'domcontentloaded' });
		await fallbackPage.waitForLoadState('networkidle', { timeout: 10000 });
		assert.equal(await fallbackPage.title(), 'Lorivo');
		assert.match(await fallbackPage.getByTestId('settings-server-name').innerText(), /Lorivo/);
		await fallbackPage.close();
	} finally {
		await browser.close();
		await server.close();
	}
});

test('settings shows owner access guidance when the server still needs its first account', async () => {
	const { server, baseURL } = await launchDevServer();
	const browser = await chromium.launch();

	try {
		const page = await browser.newPage({ viewport: { width: 1600, height: 1000 } });
		await installApiMocks(page, { signedOut: true, bootstrapAllowed: true });
		await page.goto(`${baseURL}/settings`, { waitUntil: 'domcontentloaded' });
		await page.waitForLoadState('networkidle', { timeout: 10000 });

		const body = page.locator('body');
		await body.getByRole('link', { name: 'Create Owner Account', exact: true }).waitFor({
			state: 'visible',
			timeout: 5000
		});
		assert.match(await body.innerText(), /Sign in as the owner to manage Lorivo settings\./);
		assert.match(await body.innerText(), /This server still needs its first owner account\./);

		for (const [label, hash] of [
			['Library', 'library'],
			['Scanning', 'scanning'],
			['Playback', 'playback'],
			['Access', 'access'],
			['About', 'about']
		]) {
			await page.getByTestId('settings-mode-sidebar').getByRole('link', { name: label, exact: true }).click();
			await page.waitForURL(`${baseURL}/settings#${hash}`, { timeout: 10000 });
			await page.waitForFunction(
				(expected) =>
					document.querySelector('[data-testid="settings-section-content"]')?.getAttribute('data-section') === expected,
				hash,
				{ timeout: 5000 }
			);
			const selectedSection = page.getByTestId('settings-section-content');
			await selectedSection.waitFor({ state: 'visible', timeout: 5000 });
			assert.equal(await selectedSection.getAttribute('data-section'), hash);
			assert.match(await selectedSection.innerText(), /Sign in as the owner to manage Lorivo settings\./);
			assert.equal(await selectedSection.getByRole('link', { name: 'Create Owner Account', exact: true }).count(), 1);
		}

		await assertNoHorizontalOverflow(page);
		await page.close();
	} finally {
		await browser.close();
		await server.close();
	}
});

test('settings shows development owner controls when dev auth bypass is active', async () => {
	const { server, baseURL } = await launchDevServer();
	const browser = await chromium.launch();

	try {
		const page = await browser.newPage({ viewport: { width: 1600, height: 1000 } });
		const state = await installApiMocks(page, { signedOut: true, devAuthBypass: true });
		await verifySettingsMenu(page, baseURL, { width: 1600, height: 1000 }, state);
		await page.close();
	} finally {
		await browser.close();
		await server.close();
	}
});
