import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	optimizeDeps: { exclude: ['@facile/muse'] },
	build: {
		// 'hidden' emits the maps but omits the //# sourceMappingURL comment, so
		// no browser fetches them and the sources are not published. Journal
		// resolves stacks server-side from maps uploaded at boot; the Dockerfile
		// moves them out of the served directory so they are never reachable
		// even by guessing a URL.
		sourcemap: 'hidden'
	},
	server: {
		proxy: {
			'/api': {
				target: 'http://localhost:4000',
				changeOrigin: true
			},
			'/files': {
				target: 'http://localhost:4000',
				changeOrigin: true
			}
		}
	}
});
