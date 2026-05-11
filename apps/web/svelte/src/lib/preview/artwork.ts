type ArtworkKind = 'poster' | 'backdrop' | 'hero';

interface ArtworkTheme {
	accent: string;
	glow: string;
	shape: string;
	line: string;
	recipe: 0 | 1 | 2 | 3;
}

const TITLE_THEMES: Record<string, ArtworkTheme> = {
	'Ember Harbor': { accent: '#d79056', glow: '#efb67c', shape: '#f4d2a8', line: '#ffd8b3', recipe: 0 },
	'Atlas of Dawn': { accent: '#5fa0e8', glow: '#8dc3ff', shape: '#c7dfff', line: '#b8d8ff', recipe: 1 },
	'Polar Night': { accent: '#72a7ca', glow: '#9bc8e1', shape: '#d3e6f3', line: '#b8d5e7', recipe: 2 },
	'The Last Orchard': { accent: '#a38762', glow: '#cfb085', shape: '#ead9b9', line: '#dbc59e', recipe: 3 },
	Coastline: { accent: '#56a6a3', glow: '#7fcac8', shape: '#bce8e6', line: '#a6d6d3', recipe: 0 },
	'Violet Signal': { accent: '#8671d3', glow: '#b29de9', shape: '#d5c8f5', line: '#c2b1ee', recipe: 1 },
	'Night Archive': { accent: '#5e7aa5', glow: '#8aa5d1', shape: '#c5d4ec', line: '#aec0df', recipe: 2 },
	'Broken Current': { accent: '#4f8fc2', glow: '#7db3de', shape: '#b7d5ef', line: '#a1c7e8', recipe: 3 },
	'Glass Canyon': { accent: '#b4976c', glow: '#d1b98f', shape: '#eadac0', line: '#dcc8aa', recipe: 0 },
	'Copper Sky': { accent: '#b8764f', glow: '#dda27d', shape: '#efc9ad', line: '#e4b493', recipe: 1 },
	Hinterland: { accent: '#7f9573', glow: '#abc199', shape: '#d5e4ca', line: '#c3d5b4', recipe: 2 },
	'Return Vector': { accent: '#4f8db8', glow: '#78b0d8', shape: '#b3d3eb', line: '#9ec3df', recipe: 3 },
	'Atlas Watch': { accent: '#5e9ac7', glow: '#8fc0e6', shape: '#c9e1f4', line: '#b2d3ec', recipe: 1 },
	'Ember Shore': { accent: '#c88659', glow: '#e3aa7f', shape: '#f1d0b5', line: '#e8ba99', recipe: 0 },
	Littoral: { accent: '#4f9ca4', glow: '#78c5cc', shape: '#b8e4e7', line: '#a4d4d9', recipe: 2 },
	'Low Country': { accent: '#7f8f5e', glow: '#a9bd80', shape: '#d5dfbd', line: '#c2d09f', recipe: 3 },
	'Orchard Line': { accent: '#94795b', glow: '#be9f7a', shape: '#e0ccb1', line: '#d3ba9d', recipe: 1 },
	Sunward: { accent: '#cb8a57', glow: '#e5af7a', shape: '#f1cfaa', line: '#e6bd90', recipe: 0 }
};

const DEFAULT_THEME: ArtworkTheme = {
	accent: '#58c9b0',
	glow: '#88d8c9',
	shape: '#c5ebe3',
	line: '#a3ddd0',
	recipe: 0
};

function hashString(value: string): number {
	let hash = 0;
	for (let index = 0; index < value.length; index += 1) {
		hash = (hash << 5) - hash + value.charCodeAt(index);
		hash |= 0;
	}
	return Math.abs(hash);
}

function esc(value: string): string {
	return value
		.replace(/&/g, '&amp;')
		.replace(/</g, '&lt;')
		.replace(/>/g, '&gt;')
		.replace(/"/g, '&quot;')
		.replace(/'/g, '&#39;');
}

function themeFor(title: string): ArtworkTheme {
	return TITLE_THEMES[title] || DEFAULT_THEME;
}

function recipePath(recipe: ArtworkTheme['recipe'], width: number, height: number): string {
	if (recipe === 1) {
		return `<path d='M0 ${Math.round(height * 0.66)} C ${Math.round(width * 0.2)} ${Math.round(height * 0.5)}, ${Math.round(width * 0.58)} ${Math.round(height * 0.78)}, ${width} ${Math.round(height * 0.62)} L ${width} ${height} L 0 ${height} Z' fill='rgba(255,246,229,0.11)'/>
<path d='M${Math.round(width * 0.18)} ${Math.round(height * 0.3)} L${Math.round(width * 0.54)} ${Math.round(height * 0.3)} L${Math.round(width * 0.36)} ${Math.round(height * 0.14)} Z' fill='rgba(255,246,229,0.12)'/>`;
	}
	if (recipe === 2) {
		return `<rect x='${Math.round(width * 0.14)}' y='${Math.round(height * 0.28)}' width='${Math.round(width * 0.72)}' height='${Math.round(height * 0.28)}' rx='18' fill='rgba(255,246,229,0.1)'/>
<circle cx='${Math.round(width * 0.76)}' cy='${Math.round(height * 0.28)}' r='${Math.round(width * 0.11)}' fill='rgba(255,246,229,0.12)'/>
<path d='M0 ${Math.round(height * 0.74)} C ${Math.round(width * 0.24)} ${Math.round(height * 0.62)}, ${Math.round(width * 0.58)} ${Math.round(height * 0.84)}, ${width} ${Math.round(height * 0.72)} L ${width} ${height} L 0 ${height} Z' fill='rgba(255,246,229,0.1)'/>`;
	}
	if (recipe === 3) {
		return `<path d='M0 ${Math.round(height * 0.7)} C ${Math.round(width * 0.22)} ${Math.round(height * 0.56)}, ${Math.round(width * 0.56)} ${Math.round(height * 0.86)}, ${width} ${Math.round(height * 0.68)} L ${width} ${height} L 0 ${height} Z' fill='rgba(255,246,229,0.11)'/>
<rect x='${Math.round(width * 0.18)}' y='${Math.round(height * 0.22)}' width='${Math.round(width * 0.62)}' height='${Math.round(height * 0.2)}' rx='16' fill='rgba(255,246,229,0.1)' transform='rotate(14 360 320)'/>`;
	}
	return `<path d='M0 ${Math.round(height * 0.68)} C ${Math.round(width * 0.18)} ${Math.round(height * 0.57)}, ${Math.round(width * 0.54)} ${Math.round(height * 0.75)}, ${width} ${Math.round(height * 0.62)} L ${width} ${height} L 0 ${height} Z' fill='rgba(255,246,229,0.11)'/>
<path d='M${Math.round(width * 0.12)} ${Math.round(height * 0.35)} L${Math.round(width * 0.44)} ${Math.round(height * 0.35)} L${Math.round(width * 0.28)} ${Math.round(height * 0.18)} Z' fill='rgba(255,246,229,0.11)'/>`;
}

export function previewArtwork(title: string, kind: ArtworkKind): string {
	const safeTitle = esc(title || 'Media');
	const theme = themeFor(title);
	const seed = hashString(`${title}-${kind}`);
	const width = kind === 'poster' ? 720 : 1600;
	const height = kind === 'poster' ? 1080 : 900;
	const titleSize = kind === 'poster' ? 52 : 70;
	const lowerY = kind === 'poster' ? Math.round(height * 0.8) : Math.round(height * 0.78);
	const glowX = 62 + (seed % 26);
	const glowY = 14 + (seed % 20);
	const topBandY = Math.round(height * 0.18);
	const shapeBlock = recipePath(theme.recipe, width, height);

	const svg = `<svg xmlns='http://www.w3.org/2000/svg' width='${width}' height='${height}' viewBox='0 0 ${width} ${height}'>
<defs>
<linearGradient id='bg' x1='0' y1='0' x2='0' y2='1'>
<stop offset='0%' stop-color='#1f231d'/>
<stop offset='58%' stop-color='#141710'/>
<stop offset='100%' stop-color='#0d100a'/>
</linearGradient>
<radialGradient id='glow' cx='${glowX}%' cy='${glowY}%' r='58%'>
<stop offset='0%' stop-color='${theme.glow}' stop-opacity='0.36'/>
<stop offset='100%' stop-color='${theme.glow}' stop-opacity='0'/>
</radialGradient>
</defs>
<rect width='100%' height='100%' fill='url(#bg)'/>
<rect width='100%' height='100%' fill='url(#glow)'/>
<path d='M${Math.round(width * 0.08)} ${topBandY} L${Math.round(width * 0.92)} ${topBandY}' stroke='${theme.line}' stroke-opacity='0.52' stroke-width='3'/>
${shapeBlock}
<text x='${Math.round(width * 0.08)}' y='${lowerY}' fill='rgba(255,246,229,0.96)' font-size='${titleSize}' font-family='Segoe UI, system-ui, -apple-system, sans-serif' font-weight='630' letter-spacing='0.2'>${safeTitle}</text>
<text x='${Math.round(width * 0.08)}' y='${lowerY + Math.round(titleSize * 0.9)}' fill='rgba(227,216,190,0.78)' font-size='${Math.round(titleSize * 0.42)}' font-family='Segoe UI, system-ui, -apple-system, sans-serif'>Lorivo Fictional Library</text>
</svg>`;
	return `data:image/svg+xml;charset=UTF-8,${encodeURIComponent(svg)}`;
}

export function previewPoster(title: string): string {
	return previewArtwork(title, 'poster');
}

export function previewBackdrop(title: string): string {
	return previewArtwork(title, 'backdrop');
}

export function previewHero(title: string): string {
	const safeTitle = esc(title || 'Media');
	const theme = themeFor(title);
	const width = 1920;
	const height = 1080;
	const svg = `<svg xmlns='http://www.w3.org/2000/svg' width='${width}' height='${height}' viewBox='0 0 ${width} ${height}'>
<defs>
<linearGradient id='bg' x1='0' y1='0' x2='0' y2='1'>
<stop offset='0%' stop-color='#191c17'/>
<stop offset='64%' stop-color='#11140e'/>
<stop offset='100%' stop-color='#0a0d08'/>
</linearGradient>
<radialGradient id='glow' cx='76%' cy='16%' r='52%'>
<stop offset='0%' stop-color='${theme.glow}' stop-opacity='0.34'/>
<stop offset='100%' stop-color='${theme.glow}' stop-opacity='0'/>
</radialGradient>
</defs>
<rect width='100%' height='100%' fill='url(#bg)'/>
<rect width='100%' height='100%' fill='url(#glow)'/>
<rect x='980' y='160' width='660' height='430' rx='24' fill='rgba(255,246,229,0.08)'/>
<path d='M0 740 C 360 610, 980 860, 1920 700 L1920 1080 L0 1080 Z' fill='rgba(255,246,229,0.1)'/>
<path d='M130 230 L860 230' stroke='${theme.line}' stroke-opacity='0.55' stroke-width='4'/>
<path d='M520 300 L910 300 L720 104 Z' fill='rgba(255,246,229,0.12)'/>
<text x='132' y='810' fill='rgba(255,246,229,0.96)' font-size='90' font-family='Segoe UI, system-ui, -apple-system, sans-serif' font-weight='640' letter-spacing='0.3'>${safeTitle}</text>
</svg>`;
	return `data:image/svg+xml;charset=UTF-8,${encodeURIComponent(svg)}`;
}
