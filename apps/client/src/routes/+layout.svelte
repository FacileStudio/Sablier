<script lang="ts">
	import '../app.css';
	import { browser } from '$app/environment';
	import { Toaster } from '@facile/muse';
	import { theme } from '$lib/theme.svelte';

	if (browser) {
		theme.restore();
		void import('iconify-icon');
	}

	let { children } = $props();
</script>

{@render children()}
<Toaster class="pb-28 md:pb-6" />
