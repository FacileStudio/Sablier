<script lang="ts">
	import { PreferencesSection } from '@facile/muse';
	import { rememberLocale, t } from '$lib/i18n/index.svelte';
	import { theme } from '$lib/theme.svelte';
</script>

<svelte:head>
	<title>{t('preferences.title')}</title>
</svelte:head>

<PreferencesSection
	theme={theme.preference}
	onThemeChange={(next) => theme.set(next)}
	onLocaleChange={rememberLocale}
/>
