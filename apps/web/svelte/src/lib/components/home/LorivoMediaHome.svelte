<script lang="ts">
	export interface LorivoHero {
		title: string;
		meta: string;
		description: string;
		progressLabel: string;
		progressPercent: number;
		runtime: string;
		backdropUrl: string;
		posterUrl: string;
		resumeHref: string;
		detailsHref: string;
	}

	export interface LorivoWideItem {
		title: string;
		context: string;
		progressLabel: string;
		progressPercent: number;
		imageUrl: string;
		playHref: string;
	}

	export interface LorivoPosterItem {
		title: string;
		meta: string;
		imageUrl: string;
		href: string;
	}

	let {
		hero,
		continueItems = [],
		movieItems = [],
		tvItems = [],
		previewMode = false
	} = $props<{
		hero: LorivoHero;
		continueItems: LorivoWideItem[];
		movieItems: LorivoPosterItem[];
		tvItems: LorivoPosterItem[];
		previewMode?: boolean;
	}>();

	let query = $state('');

	const filteredContinue = $derived.by(() => {
		const needle = query.trim().toLowerCase();
		if (!needle) return continueItems;
		return continueItems.filter((item: LorivoWideItem) =>
			`${item.title} ${item.context} ${item.progressLabel}`.toLowerCase().includes(needle)
		);
	});

	const filteredMovies = $derived.by(() => {
		const needle = query.trim().toLowerCase();
		if (!needle) return movieItems;
		return movieItems.filter((item: LorivoPosterItem) =>
			`${item.title} ${item.meta}`.toLowerCase().includes(needle)
		);
	});

	const filteredTV = $derived.by(() => {
		const needle = query.trim().toLowerCase();
		if (!needle) return tvItems;
		return tvItems.filter((item: LorivoPosterItem) =>
			`${item.title} ${item.meta}`.toLowerCase().includes(needle)
		);
	});

	const renderMovies = $derived.by(() => filteredMovies.filter((item: LorivoPosterItem) => !!item.imageUrl));
	const renderTV = $derived.by(() => filteredTV.filter((item: LorivoPosterItem) => !!item.imageUrl));

	function widthFromPercent(value: number): string {
		const bounded = Math.max(0, Math.min(100, Number(value) || 0));
		return `${bounded}%`;
	}
</script>

<main class="lorivo-home" data-testid="lorivo-media-home" data-preview={previewMode ? '1' : '0'}>
	<header class="topbar">
		<div class="topbar-start">
			<button class="menu-button" type="button" aria-label="Open navigation">
				<span></span>
				<span></span>
				<span></span>
			</button>
			<a class="brand" href="/">
				<img src="/preview-art/lorivo/lorivo-logo.svg" alt="Lorivo" />
			</a>
		</div>

		<label class="search-box">
			<svg viewBox="0 0 24 24" aria-hidden="true">
				<path
					d="M10.5 4.5a6 6 0 1 0 3.74 10.7l3.53 3.53a1 1 0 0 0 1.41-1.41l-3.53-3.53A6 6 0 0 0 10.5 4.5Zm0 2a4 4 0 1 1 0 8 4 4 0 0 1 0-8Z"
				></path>
			</svg>
			<input type="search" bind:value={query} placeholder="Search" aria-label="Search media" />
			<span>Ctrl K</span>
		</label>

		<div class="topbar-end">
			<button class="search-button" type="button" aria-label="Search">
				<svg viewBox="0 0 24 24" aria-hidden="true">
					<path
						d="M10.5 4.5a6 6 0 1 0 3.74 10.7l3.53 3.53a1 1 0 0 0 1.41-1.41l-3.53-3.53A6 6 0 0 0 10.5 4.5Zm0 2a4 4 0 1 1 0 8 4 4 0 0 1 0-8Z"
					></path>
				</svg>
			</button>
			<button class="avatar" type="button" aria-label="Open profile menu">
				<img src="/preview-art/lorivo/locked/avatar-lorivo.png" alt="" aria-hidden="true" />
			</button>
			<svg class="avatar-caret" viewBox="0 0 24 24" aria-hidden="true">
				<path d="M7 10.5 12 15.5 17 10.5"></path>
			</svg>
		</div>
	</header>

	<div class="content">
		<section class="hero">
			<img class="hero-backdrop" src={hero.backdropUrl} alt="" aria-hidden="true" />
			<div class="hero-overlay"></div>
			<div class="hero-content">
				<div class="hero-copy">
					<span class="featured-pill">FEATURED</span>
					<h1>{hero.title}</h1>
					<p class="hero-meta">{hero.meta}</p>
					<p class="hero-description">{hero.description}</p>
					<div class="hero-controls">
						<div class="hero-actions">
							<a class="btn btn-primary" href={hero.resumeHref}>
								<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M8 6v12l10-6z"></path></svg>
								Resume
							</a>
							<a class="btn btn-secondary" href={hero.detailsHref}>
								<svg viewBox="0 0 24 24" aria-hidden="true">
									<path
										d="M12 2a10 10 0 1 0 .001 20A10 10 0 0 0 12 2Zm0 4a1.5 1.5 0 1 1 0 3 1.5 1.5 0 0 1 0-3Zm1 12h-2v-7h2v7Z"
									></path>
								</svg>
								Details
							</a>
						</div>
						<div class="hero-progress">
							<span>{hero.progressLabel}</span>
							<div class="track">
								<div class="fill" style={`width:${widthFromPercent(hero.progressPercent)};`}></div>
							</div>
							<small>{hero.runtime}</small>
						</div>
					</div>
				</div>
				<div class="hero-poster-wrap">
					<img class="hero-poster" src={hero.posterUrl} alt={`${hero.title} poster`} loading="eager" />
				</div>
			</div>
		</section>

		<section class="media-row">
			<header>
				<h2>Continue Watching</h2>
				<a href="/continue-watching">View all</a>
			</header>
			<div class="wide-rail">
				{#each filteredContinue as item (item.title)}
					<a class="wide-card" href={item.playHref}>
						<img src={item.imageUrl} alt={item.title} loading={previewMode ? 'eager' : 'lazy'} />
						<div class="wide-overlay"></div>
						<svg class="wide-play" viewBox="0 0 24 24" aria-hidden="true"><path d="M8 5v14l11-7z"></path></svg>
						<div class="wide-copy">
							<h3>{item.title}</h3>
							<p>{item.context}</p>
						</div>
						<div class="wide-progress">
							<div class="wide-fill" style={`width:${widthFromPercent(item.progressPercent)};`}></div>
						</div>
					</a>
				{/each}
			</div>
			<span class="rail-next" aria-hidden="true">›</span>
		</section>

		<section class="media-row">
			<header>
				<h2>Recently Added Movies</h2>
				<a href="/movies">View all</a>
			</header>
				<div class="poster-rail">
					{#each renderMovies as item (item.title)}
						<a class="poster-card" href={item.href}>
							<div class="poster-art">
								<img src={item.imageUrl} alt={item.title} loading={previewMode ? 'eager' : 'lazy'} />
							</div>
							<div class="poster-copy">
								<h3>{item.title}</h3>
								<p>{item.meta}</p>
							</div>
						</a>
					{/each}
				</div>
			</section>

		<section class="media-row last-row">
			<header>
				<h2>Recently Added TV</h2>
				<a href="/tv">View all</a>
			</header>
				<div class="poster-rail">
					{#each renderTV as item (item.title)}
						<a class="poster-card" href={item.href}>
							<div class="poster-art">
								<img src={item.imageUrl} alt={item.title} loading={previewMode ? 'eager' : 'lazy'} />
							</div>
							<div class="poster-copy">
								<h3>{item.title}</h3>
								<p>{item.meta}</p>
							</div>
						</a>
					{/each}
				</div>
			</section>
	</div>

	<nav class="bottom-nav" aria-label="Media navigation">
		<a class="active" href="/">
			<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 21V9l8-6 8 6v12h-6v-7h-4v7z"></path></svg>
			Home
		</a>
		<a href="/movies">
			<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 4h16a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2Zm0 4v10h16V8H4Zm2-3h2v2H6V5Zm4 0h2v2h-2V5Zm4 0h2v2h-2V5Z"></path></svg>
			Movies
		</a>
		<a href="/tv">
			<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M3 5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5Zm2 0v11h14V5H5Zm3 15h8v2H8v-2Z"></path></svg>
			TV Shows
		</a>
		<button type="button">
			<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 3v10.55A4 4 0 1 0 14 17V7h4V3h-6ZM10 19a2 2 0 1 1 0-4 2 2 0 0 1 0 4Z"></path></svg>
			Music
		</button>
		<button type="button">
			<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M6 10a2 2 0 1 1 0 4 2 2 0 0 1 0-4Zm6 0a2 2 0 1 1 0 4 2 2 0 0 1 0-4Zm6 0a2 2 0 1 1 0 4 2 2 0 0 1 0-4Z"></path></svg>
			More
		</button>
	</nav>
</main>

<style>
	.lorivo-home {
		--bg: #020814;
		--line: rgb(146 166 205 / 21%);
		--text: #f3f7ff;
		--muted: #c5d0e8;
		--soft: #93a3c2;
		--accent: #7c5cff;
		font-family:
			Inter,
			ui-sans-serif,
			system-ui,
			-apple-system,
			BlinkMacSystemFont,
			'Segoe UI',
			sans-serif;
		min-height: 100vh;
		background: #020812;
		color: var(--text);
	}

	.topbar {
		height: 74px;
		padding: 0 22px;
		border-bottom: 1px solid var(--line);
		display: grid;
		grid-template-columns: auto 1fr auto;
		align-items: center;
		gap: 18px;
	}

	.topbar-start {
		display: inline-flex;
		align-items: center;
		gap: 14px;
	}

	.menu-button {
		width: 34px;
		height: 34px;
		display: inline-flex;
		flex-direction: column;
		justify-content: center;
		gap: 4px;
		padding: 6px 4px;
		background: transparent;
		border: 0;
		color: inherit;
	}

	.menu-button span {
		height: 2px;
		width: 20px;
		border-radius: 999px;
		background: #f1f5ff;
	}

	.brand {
		display: inline-flex;
		align-items: center;
		gap: 10px;
		text-decoration: none;
		color: var(--text);
	}

	.brand img {
		display: block;
		width: 164px;
		height: 28px;
	}

	.search-box {
		justify-self: center;
		height: 42px;
		width: min(460px, 100%);
		border: 1px solid var(--line);
		border-radius: 999px;
		display: grid;
		grid-template-columns: 32px minmax(0, 1fr) auto;
		align-items: center;
		padding: 0 11px;
		background: rgb(6 14 30 / 55%);
	}

	.search-box svg {
		width: 18px;
		height: 18px;
		fill: #95a4c5;
	}

	.search-box input {
		width: 100%;
		height: 100%;
		border: 0;
		background: transparent;
		color: var(--text);
		font-size: 1.05rem;
	}

	.search-box input:focus {
		outline: none;
	}

	.search-box span {
		color: #9db8ff;
		font-size: 0.72rem;
		white-space: nowrap;
	}

	.topbar-end {
		display: inline-flex;
		align-items: center;
		gap: 8px;
	}

	.search-button {
		display: none;
		width: 34px;
		height: 34px;
		border: 0;
		padding: 0;
		background: transparent;
		color: inherit;
	}

	.search-button svg {
		width: 24px;
		height: 24px;
		fill: #f2f6ff;
	}

	.avatar {
		width: 42px;
		height: 42px;
		border-radius: 999px;
		border: 1px solid rgb(214 224 244 / 18%);
		background: #0b1427;
		overflow: hidden;
		padding: 0;
		display: inline-flex;
		align-items: center;
		justify-content: center;
	}

	.avatar img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		display: block;
	}

	.avatar-caret {
		width: 14px;
		height: 14px;
		stroke: #dde6f8;
		stroke-width: 2;
		fill: none;
		opacity: 0.82;
	}

	.content {
		padding: 18px 24px 32px;
	}

	.hero {
		position: relative;
		min-height: 410px;
		border: 1px solid rgb(132 149 182 / 28%);
		border-radius: 12px;
		overflow: hidden;
		background: #030a16;
	}

	.hero-backdrop {
		position: absolute;
		inset: 0;
		width: 100%;
		height: 100%;
		object-fit: cover;
		object-position: 74% center;
	}

	.hero-overlay {
		position: absolute;
		inset: 0;
		background:
			linear-gradient(90deg, rgb(3 8 18 / 94%) 0%, rgb(3 8 18 / 82%) 46%, rgb(3 8 18 / 58%) 100%),
			linear-gradient(180deg, rgb(3 7 14 / 22%) 0%, rgb(3 7 14 / 88%) 100%);
	}

	.hero-content {
		position: relative;
		z-index: 1;
		height: 100%;
		display: grid;
		grid-template-columns: minmax(0, 1fr) 300px;
		align-items: center;
		padding: 40px 38px 28px 56px;
		gap: 30px;
	}

	.hero-copy h1 {
		margin: 0;
		font-size: clamp(2.9rem, 3.25vw, 3.9rem);
		letter-spacing: 0;
		line-height: 1.02;
	}

	.featured-pill {
		display: inline-flex;
		align-items: center;
		height: 26px;
		margin-bottom: 20px;
		padding: 0 11px;
		border: 1px solid rgb(145 166 205 / 55%);
		border-radius: 999px;
		color: #d9e7ff;
		font-size: 0.72rem;
		font-weight: 600;
		letter-spacing: 0.08em;
	}

	.hero-meta {
		margin: 12px 0 0;
		font-size: 1.05rem;
		color: #c6d1e9;
	}

	.hero-description {
		margin: 14px 0 0;
		max-width: 620px;
		font-size: 1rem;
		line-height: 1.5;
		color: #d5deef;
	}

	.hero-controls {
		margin-top: 24px;
		display: flex;
		align-items: center;
		gap: 18px;
	}

	.hero-actions {
		display: flex;
		gap: 10px;
	}

	.btn {
		height: 44px;
		min-width: 120px;
		padding: 0 16px;
		border-radius: 8px;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		font-size: 0.96rem;
		font-weight: 560;
		text-decoration: none;
	}

	.btn svg {
		width: 18px;
		height: 18px;
		fill: currentColor;
	}

	.btn-primary {
		background: linear-gradient(140deg, #7f58ff, #6641ff);
		color: #f8f2ff;
	}

	.btn-secondary {
		border: 1px solid rgb(171 183 209 / 36%);
		background: rgb(6 12 24 / 58%);
		color: #edf2ff;
	}

	.hero-progress {
		display: flex;
		align-items: center;
		gap: 10px;
		padding-left: 18px;
		border-left: 1px solid rgb(158 170 194 / 28%);
		min-width: 156px;
		font-size: 0.92rem;
		color: #c7d2e9;
	}

	.hero-progress small {
		color: #9db8ff;
		font-size: inherit;
	}

	.track {
		width: 116px;
		height: 4px;
		border-radius: 999px;
		background: rgb(164 177 202 / 24%);
		overflow: hidden;
	}

	.fill {
		height: 100%;
		background: linear-gradient(90deg, #7f58ff, #8f5dff);
		border-radius: inherit;
	}

	.hero-poster-wrap {
		display: flex;
		justify-content: center;
		align-items: center;
	}

	.hero-poster {
		width: 264px;
		aspect-ratio: 2 / 3;
		object-fit: cover;
		border-radius: 10px;
		border: 1px solid rgb(187 198 220 / 24%);
		box-shadow: 0 14px 26px rgb(0 0 0 / 35%);
	}

	.media-row {
		margin-top: 24px;
		position: relative;
		isolation: isolate;
		display: grid;
		row-gap: 8px;
		contain: layout paint;
	}

	.media-row > header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 0;
		position: relative;
		z-index: 2;
		contain: layout paint;
	}

	.media-row h2 {
		margin: 0;
		font-size: 1.9rem;
		font-weight: 700;
	}

	.media-row header a {
		color: #8b5cff;
		text-decoration: none;
		font-size: 1rem;
	}

	.wide-rail,
	.poster-rail {
		display: flex;
		gap: 12px;
		flex-wrap: nowrap;
		align-items: flex-start;
		overflow-x: auto;
		overflow-y: hidden;
		scrollbar-width: none;
		padding-bottom: 2px;
		position: relative;
		z-index: 1;
		contain: paint;
	}

	.rail-next {
		display: none;
	}

	.wide-rail::-webkit-scrollbar,
	.poster-rail::-webkit-scrollbar {
		display: none;
	}

	.wide-card {
		position: relative;
		flex: 0 0 292px;
		aspect-ratio: 16 / 9;
		border-radius: 8px;
		overflow: hidden;
		border: 1px solid rgb(132 149 182 / 28%);
		text-decoration: none;
	}

	.wide-card img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		object-position: center 34%;
	}

	.wide-overlay {
		position: absolute;
		inset: 0;
		background: linear-gradient(180deg, rgb(0 0 0 / 16%) 46%, rgb(0 0 0 / 82%) 100%);
	}

	.wide-play {
		position: absolute;
		left: 12px;
		bottom: 42px;
		width: 14px;
		height: 14px;
		fill: rgba(242, 247, 255, 0.85);
		pointer-events: none;
	}

	.wide-copy {
		position: absolute;
		left: 12px;
		right: 10px;
		bottom: 10px;
		color: #f2f7ff;
	}

	.wide-copy p {
		margin: 0;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.wide-copy h3 {
		margin: 0;
		font-size: 0.82rem;
		font-weight: 530;
		line-height: 1.18;
		color: #f4f7ff;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.wide-copy p {
		margin-top: 1px;
		font-size: 0.76rem;
		font-weight: 450;
		color: #b5c2dc;
	}

	.wide-progress {
		position: absolute;
		left: 10px;
		right: 10px;
		bottom: 4px;
		height: 3px;
		border-radius: 999px;
		background: rgb(156 169 194 / 24%);
	}

	.wide-fill {
		height: 100%;
		border-radius: inherit;
		background: #7a4fff;
	}

	.poster-card {
		flex: 0 0 152px;
		text-decoration: none;
	}

	.poster-art {
		aspect-ratio: 2 / 3;
		border-radius: 7px;
		overflow: hidden;
		border: 1px solid rgb(132 149 182 / 28%);
	}

	.poster-art img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		display: block;
	}

	.poster-copy {
		padding-top: 8px;
	}

	.poster-copy h3,
	.poster-copy p {
		margin: 0;
	}

	.poster-copy h3 {
		font-size: 0.95rem;
		line-height: 1.3;
		color: #eff4ff;
	}

	.poster-copy p {
		margin-top: 4px;
		font-size: 0.82rem;
		color: #9ba7c5;
	}

	.last-row {
		padding-bottom: 16px;
	}

	.bottom-nav {
		display: none;
	}

	@media (min-width: 1200px) {
		.lorivo-home {
			position: relative;
			height: 100vh;
			min-height: 100vh;
			overflow: hidden;
			background:
				radial-gradient(circle at 54% 24%, rgb(23 74 122 / 18%), transparent 430px),
				linear-gradient(180deg, #031020 0%, #020814 52%, #010611 100%);
		}

		.topbar {
			position: relative;
			display: block;
			height: 73px;
			padding: 0;
			border-bottom: 1px solid rgb(111 129 159 / 22%);
			background: rgb(3 8 16 / 58%);
		}

		.topbar-start {
			position: absolute;
			left: 24px;
			top: 22px;
			width: 214px;
			height: 30px;
			gap: 0;
		}

		.menu-button {
			position: absolute;
			left: 0;
			top: 4px;
			width: 22px;
			height: 18px;
			padding: 0;
			gap: 5px;
		}

		.menu-button span {
			width: 20px;
			height: 2px;
			background: rgb(231 238 251 / 92%);
		}

		.brand {
			position: absolute;
			left: 50px;
			top: -1px;
			width: 164px;
			height: 30px;
		}

		.brand img {
			width: 100%;
			height: 100%;
		}

		.search-box {
			position: absolute;
			left: calc(50vw - 230px);
			top: 16px;
			width: 460px;
			height: 42px;
			padding: 0 15px;
			grid-template-columns: 28px minmax(0, 1fr) auto;
			background: rgb(5 13 30 / 48%);
			border-color: rgb(129 146 183 / 24%);
		}

		.search-box svg {
			width: 16px;
			height: 16px;
		}

		.search-box input {
			font-size: 16px;
		}

		.topbar-end {
			position: absolute;
			right: 22px;
			top: 17px;
			width: 40px;
			height: 40px;
		}

		.avatar {
			width: 40px;
			height: 40px;
			border-color: rgb(214 224 244 / 18%);
		}

		.avatar-caret {
			display: none;
		}

		.content {
			position: relative;
			height: calc(100vh - 73px);
			padding: 0 20px;
			overflow: hidden;
		}

		.hero {
			position: absolute;
			left: 20px;
			top: 15px;
			width: calc(100vw - 40px);
			height: 390px;
			min-height: 0;
			border: 1px solid rgb(132 149 182 / 25%);
			border-radius: 12px;
		}

		.hero-backdrop {
			object-position: 56% center;
			filter: brightness(0.72) contrast(1.06) saturate(1.02);
		}

		.hero-overlay {
			background:
				linear-gradient(90deg, rgb(2 8 18 / 96%) 0%, rgb(2 8 18 / 80%) 42%, rgb(2 8 18 / 60%) 100%),
				linear-gradient(180deg, rgb(2 7 14 / 18%) 0%, rgb(2 7 14 / 88%) 100%);
		}

		.hero-content {
			position: absolute;
			inset: 0;
			display: block;
			padding: 0;
		}

		.hero-copy {
			position: absolute;
			left: 74px;
			top: 31px;
			width: 540px;
			height: 330px;
		}

		.featured-pill {
			height: 26px;
			margin-bottom: 20px;
			font-size: 12px;
		}

		.hero-copy h1 {
			width: 480px;
			min-height: 52px;
			font-size: 52px;
			font-weight: 750;
			line-height: 1.04;
		}

		.hero-meta {
			margin-top: 14px;
			width: 300px;
			min-height: 24px;
			font-size: 22px;
			line-height: 1.25;
			color: #d7e6ff;
		}

		.hero-description {
			margin-top: 18px;
			max-width: 520px;
			min-height: 64px;
			font-size: 17px;
			line-height: 1.45;
			color: #f1f6ff;
		}

		.hero-controls {
			position: absolute;
			left: 0;
			top: 246px;
			width: 720px;
			height: 84px;
			margin-top: 0;
			gap: 12px;
			flex-wrap: wrap;
		}

		.btn {
			height: 48px;
			min-width: 128px;
			padding: 0 22px;
			border-radius: 8px;
			font-size: 16px;
			font-weight: 700;
		}

		.btn-secondary {
			background: rgb(9 13 23 / 72%);
			border-color: rgb(171 183 209 / 30%);
		}

		.hero-progress {
			position: absolute;
			left: 0;
			top: 68px;
			width: 730px;
			gap: 10px;
			padding-left: 0;
			border-left: 0;
			font-size: 15px;
		}

		.track {
			width: 130px;
			height: 4px;
			margin-left: 410px;
		}

		.hero-poster-wrap {
			position: absolute;
			right: 50px;
			top: 31px;
			display: block;
			width: 220px;
			height: 330px;
		}

		.hero-poster {
			width: 100%;
			height: 100%;
			border-radius: 8px;
			border: 1px solid rgb(183 197 223 / 20%);
			box-shadow: 0 18px 30px rgb(0 0 0 / 30%);
		}

		.media-row {
			position: absolute;
			display: block;
			margin-top: 0;
			contain: layout paint;
		}

		.content > .media-row:nth-of-type(2) {
			left: 20px;
			top: 428px;
			width: calc(100vw - 40px);
			height: 270px;
		}

		.content > .media-row:nth-of-type(3) {
			left: 20px;
			top: 651px;
			width: calc(100vw - 40px);
			height: 330px;
		}

		.content > .media-row:nth-of-type(4) {
			left: 20px;
			top: 902px;
			width: calc(100vw - 40px);
			height: 260px;
		}

		.media-row > header {
			height: 42px;
			min-height: 0;
		}

		.media-row h2 {
			font-size: 30px;
			font-weight: 750;
			line-height: 1.2;
		}

		.media-row header a {
			display: inline-flex;
			align-items: center;
			font-size: 16px;
		}

		.wide-rail,
		.poster-rail {
			gap: 12px;
			align-items: flex-start;
			padding-bottom: 0;
		}

		.wide-rail {
			position: absolute;
			left: 0;
			top: 52px;
			width: 100%;
			height: 146px;
		}

		.poster-rail {
			position: absolute;
			left: 0;
			top: 54px;
			width: 100%;
		}

		.content > .media-row:nth-of-type(2) .rail-next {
			display: none;
		}

		.wide-card {
			flex: 0 0 260px;
			width: 260px;
			height: 146px;
			aspect-ratio: auto;
			border-radius: 7px;
		}

		.wide-play {
			left: 11px;
			bottom: 40px;
			width: 18px;
			height: 18px;
		}

		.wide-copy {
			left: 40px;
			right: 8px;
			bottom: 13px;
		}

		.wide-copy h3 {
			font-size: 16px;
			font-weight: 750;
		}

		.wide-copy p {
			font-size: 14px;
			color: #f2f7ff;
		}

		.wide-progress {
			left: 10px;
			right: 10px;
			bottom: 4px;
			height: 4px;
		}

		.poster-card,
		.content > .media-row:nth-of-type(3) .poster-card {
			flex: 0 0 142px;
		}

		.poster-art,
		.content > .media-row:nth-of-type(3) .poster-art {
			width: 142px;
			height: 213px;
			aspect-ratio: auto;
			border-radius: 7px;
			background: #060d1e;
		}

		.poster-copy {
			display: none;
		}

		.bottom-nav {
			display: none !important;
		}
	}

	@media (max-width: 1199px) {
		.content {
			padding: 10px 16px 102px;
		}

		.search-box {
			width: min(430px, 100%);
		}

		.search-button {
			display: inline-flex;
			align-items: center;
			justify-content: center;
		}

		.hero {
			min-height: 312px;
		}

		.hero-content {
			grid-template-columns: minmax(0, 1fr) 206px;
			padding: 24px 20px 20px 44px;
			gap: 22px;
		}

		.hero-copy h1 {
			font-size: clamp(2rem, 4.1vw, 3rem);
		}

		.hero-meta {
			font-size: 1.15rem;
		}

		.hero-description {
			font-size: 0.98rem;
		}

		.hero-poster {
			width: 184px;
		}

		.wide-card {
			flex-basis: 236px;
		}

		.poster-card {
			flex-basis: 130px;
		}

		.bottom-nav {
			position: fixed;
			left: 0;
			right: 0;
			bottom: 0;
			height: 74px;
			background: rgb(2 8 18 / 94%);
			border-top: 1px solid rgb(131 150 186 / 25%);
			display: grid;
			grid-template-columns: repeat(5, minmax(0, 1fr));
			z-index: 30;
		}

		.bottom-nav a,
		.bottom-nav button {
			height: 100%;
			border: 0;
			background: transparent;
			display: inline-flex;
			flex-direction: column;
			align-items: center;
			justify-content: center;
			gap: 4px;
			color: #aab5cd;
			font-size: 0.72rem;
			font-weight: 500;
			text-decoration: none;
		}

		.bottom-nav svg {
			width: 22px;
			height: 22px;
			fill: currentColor;
		}

		.bottom-nav a.active {
			color: #8a64ff;
		}
	}

	@media (max-width: 767px) {
		.topbar {
			height: 68px;
			padding: 0 14px;
		}

		.topbar-start {
			gap: 14px;
		}

		.brand img {
			width: 145px;
			height: 25px;
		}

		.search-box {
			display: none;
		}

		.content {
			padding: 10px 12px 98px;
		}

		.hero {
			min-height: 374px;
		}

		.hero-content {
			grid-template-columns: 1fr;
			padding: 20px 14px 16px;
			gap: 12px;
		}

		.featured-pill {
			height: 23px;
			margin-bottom: 13px;
			font-size: 0.66rem;
		}

		.hero-copy h1 {
			margin-top: 0;
			font-size: 2rem;
			line-height: 1.08;
		}

		.hero-meta {
			margin-top: 12px;
			font-size: 1rem;
		}

		.hero-description {
			max-width: 220px;
			margin-top: 14px;
			font-size: 1rem;
			line-height: 1.36;
		}

		.hero-controls {
			margin-top: 16px;
			display: grid;
			grid-template-columns: 1fr;
			gap: 13px;
			width: 100%;
		}

		.btn {
			height: 43px;
			min-width: 110px;
			font-size: 1rem;
		}

		.hero-progress {
			margin-top: 0;
			display: grid;
			padding-left: 0;
			border-left: 0;
			min-width: 0;
			width: 100%;
			grid-template-columns: auto minmax(120px, 1fr);
			align-items: center;
			font-size: 0.92rem;
		}

		.hero-progress small {
			display: none;
		}

		.track {
			width: auto;
		}

		.hero-poster-wrap {
			position: absolute;
			right: 12px;
			top: 38px;
		}

		.hero-poster {
			width: 112px;
		}

		.media-row {
			margin-top: 14px;
		}

		.media-row h2 {
			font-size: 1.6rem;
		}

		.media-row header a {
			font-size: 0.9rem;
		}

		.wide-card {
			flex-basis: 255px;
		}

		.poster-card {
			flex-basis: 146px;
		}

		.poster-copy h3 {
			font-size: 0.98rem;
		}

		.bottom-nav a,
		.bottom-nav button {
			font-size: 0.68rem;
		}
	}
</style>
