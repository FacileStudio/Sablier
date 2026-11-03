<script lang="ts">
	import '../app.css';
	import { browser } from '$app/environment';
	import { Toaster } from '@facile/muse';
	import { theme } from '$lib/theme.svelte';
	import { registerDeepLinks } from '$lib/native';
	import { isApp } from '$lib/platform';

	if (browser) {
		theme.restore();
		void import('iconify-icon');
		if (isApp()) registerDeepLinks();
	}

	let { children } = $props();
</script>

{@render children()}
<Toaster class="pb-28 md:pb-6" />
