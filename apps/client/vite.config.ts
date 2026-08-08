import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	optimizeDeps: { exclude: ['@facile/muse'] },
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
