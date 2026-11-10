<script lang="ts">
	import '../app.css';
	import { browser } from '$app/environment';
	import { Toaster, provideLocale } from '@facile/muse';
	import { bindLocale, localeSource } from '$lib/i18n/index.svelte';
	import { theme } from '$lib/theme.svelte';
	import { registerDeepLinks } from '$lib/native';
	import { isApp } from '$lib/platform';
	import { dropLegacyToken } from '$lib/constants';

	if (browser) {
		theme.restore();
		dropLegacyToken();
		void import('iconify-icon');
		if (isApp()) registerDeepLinks();
	}

	let { children } = $props();

	const i18n = provideLocale(localeSource);
	bindLocale(() => i18n.locale);
</script>

{@render children()}
<Toaster class="pb-28 md:pb-6" />
