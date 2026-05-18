import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	server: {
		proxy: {
			// Forward all /api calls from the Vite dev server to the Go backend.
			// This lets you visit http://localhost:5173 with hot reload while the
			// Go server handles all data — no CORS issues, no dual-origin pain.
			'/api': {
				target: 'http://127.0.0.1:8097',
				changeOrigin: true
			}
		}
	}
});
