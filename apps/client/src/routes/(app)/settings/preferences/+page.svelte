<script lang="ts">
	import { Button, PreferencesSection, SettingsRow, SettingsSection, icons } from '@facile/muse';
	import { rememberLocale, t } from '$lib/i18n/index.svelte';
	import { isApp } from '$lib/platform';
	import { theme } from '$lib/theme.svelte';
</script>

<svelte:head>
	<title>{t('preferences.title')}</title>
</svelte:head>

<PreferencesSection
	subscription={{ href: 'https://sso.facile.studio/spaces' }}
	theme={theme.preference}
	onThemeChange={(next) => theme.set(next)}
	onLocaleChange={rememberLocale}
/>

{#if !isApp()}
	<SettingsSection title={t('nav.downloads')}>
		<SettingsRow label={t('nav.downloads')} description={t('downloads.intro')}>
			<Button href="/downloads" variant="secondary" icon={icons.download}>{t('nav.downloads')}</Button>
		</SettingsRow>
	</SettingsSection>
{/if}
