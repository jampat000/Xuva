import adapter from '@sveltejs/adapter-static';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	compilerOptions: {
		// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
		runes: ({ filename }) => (filename.split(/[/\\]/).includes('node_modules') ? undefined : true)
	},
	kit: {
		// NOTE: adapter-static calls rimraf() on the output directory before every build.
		// When the target is the SMB share (Z:\) this fails with EPERM on the
		// static-next/icons/pwa/ ghost directory (a filesystem permission artifact on the
		// NAS that cannot be deleted remotely).
		//
		// BUILD WORKAROUND (required until the NAS permissions are fixed):
		//   1. Temporarily change 'pages' and 'assets' to 'C:/xbld'
		//   2. Run: npm run build
		//   3. Run: robocopy C:\xbld\\_app static-next\\_app /MIR  (and /LEV:1 for root files)
		//   4. Restore this config
		//
		// Do NOT commit the config with C:/xbld — always restore before committing.
		adapter: adapter({
			pages: '../../../server/internal/webapp/static-next',
			assets: '../../../server/internal/webapp/static-next',
			fallback: 'index.html'
		})
	}
};

export default config;
