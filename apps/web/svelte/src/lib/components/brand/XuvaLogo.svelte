<script lang="ts">
	let { size = 22, title = 'Xuva' } = $props<{ size?: number; title?: string }>();
</script>

<svg
	class="xuva-logo"
	width={size}
	height={size}
	viewBox="0 0 64 64"
	fill="none"
	xmlns="http://www.w3.org/2000/svg"
	role="img"
	aria-label={title}
>
	<title>{title}</title>
	<defs>
		<linearGradient id="xuva-logo-main" x1="4" y1="32" x2="60" y2="32" gradientUnits="userSpaceOnUse">
			<stop offset="0%" stop-color="var(--xuva-brand-stop-1, #A78BFA)" />
			<stop offset="55%" stop-color="var(--xuva-brand-stop-2, #7C3AED)" />
			<stop offset="100%" stop-color="var(--xuva-brand-stop-3, #DB2777)" />
		</linearGradient>
		<linearGradient id="xuva-logo-shadow" x1="0" y1="0" x2="0" y2="64" gradientUnits="userSpaceOnUse">
			<stop offset="0%" stop-color="#000000" stop-opacity="0.35" />
			<stop offset="100%" stop-color="#000000" stop-opacity="0" />
		</linearGradient>
	</defs>

	<g transform="translate(0.6, 1.2)" opacity="0.5">
		<path d="M 4 4 L 16 4 L 60 32 L 30 32 Z" fill="#000" />
		<path d="M 4 60 L 16 60 L 60 32 L 30 32 Z" fill="#000" />
	</g>

	<path d="M 4 4 L 16 4 L 60 32 L 30 32 Z" fill="url(#xuva-logo-main)" />
	<path d="M 4 60 L 16 60 L 60 32 L 30 32 Z" fill="url(#xuva-logo-main)" />

	<path d="M 6 4.6 L 14 4.6 L 58.5 31 L 56 32 Z" fill="#FFFFFF" opacity="0.32" />
	<path d="M 30 32 L 33 30.2 L 33 33.8 Z" fill="#FFFFFF" opacity="0.55" />
	<path d="M 4 60 L 16 60 L 60 32 L 30 32 Z" fill="url(#xuva-logo-shadow)" />
</svg>

<style>
	.xuva-logo {
		display: block;
		filter: var(--xuva-brand-mark-shadow, drop-shadow(0 8px 18px rgb(124 92 255 / 20%)));
		flex: 0 0 auto;
	}
</style>

