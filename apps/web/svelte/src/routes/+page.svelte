<script lang="ts">
	import Hero from '$lib/lorivo/Hero.svelte';
	import LandscapeCard from '$lib/lorivo/LandscapeCard.svelte';
	import PosterCard from '$lib/lorivo/PosterCard.svelte';
	import Row from '$lib/lorivo/Row.svelte';
	import TopBar from '$lib/lorivo/TopBar.svelte';

	const tmdb = (path: string) => `https://image.tmdb.org/t/p/w500${path}`;
	const tmdbBg = (path: string) => `https://image.tmdb.org/t/p/original${path}`;

	// Hero: Dune: Part Two — real poster + matching backdrop
	const HERO_POSTER = tmdb('/1pdfLvkbY9ohJlCjQH2CZjjYVvJ.jpg');
	const HERO_BACKDROP = tmdbBg('/xOMo8BRK7PfcJv9JCnx7s5hj0PX.jpg');

	const continueWatching = [
		{ title: 'Dune: Part Two', sub: '1h 14m left', progress: 55, img: tmdb('/1pdfLvkbY9ohJlCjQH2CZjjYVvJ.jpg') },
		{ title: 'The Last of Us', sub: 'S1 E4 · 24m left', progress: 70, img: tmdb('/uKvVjHNqB5VmOrdxqAt2F7J78ED.jpg') },
		{ title: 'Foundation', sub: 'S2 E3 · 32m left', progress: 40, img: tmdb('/tg9I5pOY4M9CKj8U0cxVBTsm5eh.jpg') },
		{ title: 'Oppenheimer', sub: '1h 20m left', progress: 35, img: tmdb('/8Gxv8gSFCU0XGDykEGv7zR1n2ua.jpg') },
		{ title: 'The Batman', sub: '40m left', progress: 80, img: tmdb('/b0PlSFdDwbyK0cf5RxwDpaOJQvQ.jpg') },
		{ title: 'Interstellar', sub: '1h 7m left', progress: 50, img: tmdb('/gEU2QniE6E77NI6lCU6MxlNBvIx.jpg') },
	];

	const recentMovies = [
		{ title: 'Mufasa: The Lion King', img: tmdb('/lurEK87kukWNaHd0zYnsi3yzJrs.jpg') },
		{ title: 'Twisters', img: tmdb('/pjnD08FlMAIXsfOLKQbvmO0f0MD.jpg') },
		{ title: 'A Quiet Place: Day One', img: tmdb('/hU42CRk14JuPEdqZG3AWmagiPAP.jpg') },
		{ title: 'The Fall Guy', img: tmdb('/e7olqFmzcIX5c23kX4zSmLPJi8c.jpg') },
		{ title: 'Kingdom of the Planet of the Apes', img: tmdb('/gKkl37BQuKTanygYQG1pyYgLVgf.jpg') },
		{ title: 'Civil War', img: tmdb('/dXNAPwY7VrqMAo51EKhhCJfaGb5.jpg') },
		{ title: 'Dune: Part Two', img: tmdb('/1pdfLvkbY9ohJlCjQH2CZjjYVvJ.jpg') },
		{ title: 'Top Gun: Maverick', img: tmdb('/62HCnUTziyWcpDaBO2i1DX17ljH.jpg') },
		{ title: 'Deadpool & Wolverine', img: tmdb('/8cdWjvZQUExUUTzyp4t6EDMubfO.jpg') },
		{ title: 'The Wild Robot', img: tmdb('/9w0Vh9eizfBXrcomiaFWTIPdboo.jpg') },
	];

	const recentTV = [
		{ title: 'The Penguin', ep: 'S1 E4', img: tmdb('/vOWcqC4oDQws1doDWLO7d3dh5qc.jpg') },
		{ title: 'House of the Dragon', ep: 'S2 E2', img: tmdb('/7QMsOTMUswlwxJP0rTTZfmz2tX2.jpg') },
		{ title: 'The Boys', ep: 'S4 E6', img: tmdb('/2zmTngn1tYC1AvfnrFLhxeD82hz.jpg') },
		{ title: 'Slow Horses', ep: 'S4 E3', img: tmdb('/dnpatlJrEPiDSn5fzgzvxtiSnMo.jpg') },
		{ title: 'Reacher', ep: 'S3 E4', img: tmdb('/31GlRQMiDunO8cl3NxTz34U64rf.jpg') },
		{ title: 'Severance', ep: 'S2 E4', img: tmdb('/pPHpeI2X1qEd1CS1SeyrdhZ4qnT.jpg') },
		{ title: 'The Night Agent', ep: 'S2 E2', img: tmdb('/4c5yUNcaff4W4aPrkXE6zr7papX.jpg') },
		{ title: 'Silo', ep: 'S2 E1', img: tmdb('/tlliQuCupf8fpTH7RAor3aKMGy.jpg') },
		{ title: '3 Body Problem', ep: 'S1 E7', img: tmdb('/ykZ7hlShkdRQaL2aiieXdEMmrLb.jpg') },
		{ title: 'Monarch: Legacy of Monsters', ep: 'S1 E8', img: tmdb('/7LBbaEaLSbqdviBYaSS1rRPMnrs.jpg') },
	];
</script>

<svelte:head>
	<title>Lorivo — Stream Movies & TV</title>
	<meta name="description" content="Lorivo: your personal streaming hub for movies and TV." />
</svelte:head>

<div class="min-h-screen bg-[#0B1120] font-sans text-white antialiased">
	<TopBar />
	<Hero heroPoster={HERO_POSTER} heroBackdrop={HERO_BACKDROP} />
	<Row title="Continue Watching">
		{#each continueWatching as m (m.title)}
			<LandscapeCard item={m} />
		{/each}
	</Row>
	<Row title="Recently Added Movies">
		{#each recentMovies as m (m.title)}
			<PosterCard img={m.img} title={m.title} />
		{/each}
	</Row>
	<Row title="Recently Added TV">
		{#each recentTV as m (m.title)}
			<PosterCard img={m.img} title={m.title} ep={m.ep} />
		{/each}
	</Row>
	<div class="h-16"></div>
</div>
