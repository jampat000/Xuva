import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

const webDevPort = Number.parseInt(process.env.XUVA_WEB_DEV_PORT ?? '5174', 10);
const resolvedWebDevPort = Number.isFinite(webDevPort) ? webDevPort : 5174;
const apiOrigin = process.env.XUVA_API_ORIGIN ?? 'http://127.0.0.1:8097';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	server: {
		port: resolvedWebDevPort,
		strictPort: true,
		proxy: {
			// Forward /api from Vite to the Go backend so dev uses one browser origin.
			'/api': {
				target: apiOrigin,
				changeOrigin: true,
				configure: (proxy) => {
					proxy.on('error', (_err, _req, res) => {
						if (!res || ('destroyed' in res && res.destroyed)) return;
						const response = res as import('http').ServerResponse;
						if (!response.headersSent) {
							response.writeHead(502, { 'Content-Type': 'application/json' });
						}
						response.end(JSON.stringify({ error: `Xuva API dev proxy could not reach ${apiOrigin}` }));
					});
				}
			}
		}
	}
});
