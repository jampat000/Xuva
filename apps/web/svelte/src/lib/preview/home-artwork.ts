const BASE = '/preview-art/lorivo';

const HERO_BACKDROP = `${BASE}/ember-harbor-backdrop.webp`;
const HERO_POSTER = `${BASE}/ember-harbor-poster.webp`;

const POSTERS: Record<string, string> = {
	'Ember Harbor': `${BASE}/ember-harbor-poster.webp`,
	'Atlas of Dawn': `${BASE}/atlas-of-dawn-poster.webp`,
	'Polar Night': `${BASE}/polar-night-poster.webp`,
	'Night Archive': `${BASE}/night-archive-poster.webp`,
	'The Last Orchard': `${BASE}/the-last-orchard-poster.webp`,
	'Violet Signal': `${BASE}/violet-signal-poster.webp`,
	'Broken Current': `${BASE}/broken-current-poster.webp`,
	'Glass Canyon': `${BASE}/glass-canyon-poster.webp`,
	Coastline: `${BASE}/coastline-poster.webp`,
	'Return Vector': `${BASE}/return-vector-poster.webp`,
	'Night Orbit': `${BASE}/night-orbit-poster.webp`,
	Littoral: `${BASE}/littoral-poster.webp`,
	Sunward: `${BASE}/sunward-poster.webp`,
	Archipelago: `${BASE}/archipelago-poster.webp`
};

const CONTINUE_BACKDROPS: Record<string, string> = {
	'Atlas of Dawn': `${BASE}/atlas-of-dawn-wide.webp`,
	'The Last Orchard': `${BASE}/the-last-orchard-wide.webp`,
	'Violet Signal': `${BASE}/violet-signal-wide.webp`,
	Coastline: `${BASE}/coastline-wide.webp`,
	'Polar Night': `${BASE}/polar-night-wide.webp`,
	'Return Vector': `${BASE}/return-vector-wide.webp`
};

const BACKDROPS: Record<string, string> = {
	'Ember Harbor': HERO_BACKDROP,
	'Atlas of Dawn': CONTINUE_BACKDROPS['Atlas of Dawn'],
	'The Last Orchard': CONTINUE_BACKDROPS['The Last Orchard'],
	'Violet Signal': CONTINUE_BACKDROPS['Violet Signal'],
	Coastline: CONTINUE_BACKDROPS.Coastline,
	'Polar Night': CONTINUE_BACKDROPS['Polar Night'],
	'Return Vector': CONTINUE_BACKDROPS['Return Vector'],
	'Night Archive': `${BASE}/night-archive-poster.webp`,
	'Broken Current': `${BASE}/broken-current-poster.webp`,
	'Glass Canyon': `${BASE}/glass-canyon-poster.webp`,
	'Night Orbit': `${BASE}/night-orbit-poster.webp`,
	Littoral: `${BASE}/littoral-poster.webp`,
	Sunward: `${BASE}/sunward-poster.webp`,
	Archipelago: `${BASE}/archipelago-poster.webp`
};

export function homePreviewPoster(title: string): string {
	return POSTERS[title] || HERO_POSTER;
}

export function homePreviewBackdrop(title: string): string {
	return BACKDROPS[title] || CONTINUE_BACKDROPS[title] || HERO_BACKDROP;
}

export function homePreviewContinueBackdrop(title: string): string {
	return CONTINUE_BACKDROPS[title] || homePreviewBackdrop(title);
}

export function homePreviewHeroBackdrop(): string {
	return HERO_BACKDROP;
}

export function homePreviewHeroPoster(): string {
	return HERO_POSTER;
}
