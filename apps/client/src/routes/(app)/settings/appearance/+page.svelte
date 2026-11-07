<script lang="ts">
	import { OptionCards, SettingsRow, SettingsSection, icons } from '@facile/muse';
	import { theme, type ThemePreference } from '$lib/theme.svelte';
	import { t } from '$lib/i18n/index.svelte';

	const options = $derived([
		{ value: 'system', label: t('appearance.system'), icon: icons.monitor },
		{ value: 'light', label: t('appearance.light'), icon: icons.sun },
		{ value: 'dark', label: t('appearance.dark'), icon: icons.moon }
	]);
</script>

<svelte:head>
	<title>{t('appearance.title')}</title>
</svelte:head>

<SettingsSection title={t('appearance.heading')} description={t('appearance.description')}>
	<SettingsRow
		stacked
		label={t('appearance.scheme')}
		description={t('appearance.schemeDescription')}
	>
		<OptionCards
			label={t('appearance.scheme')}
			{options}
			value={theme.preference}
			onSelect={(next) => theme.set(next as ThemePreference)}
		/>
	</SettingsRow>
</SettingsSection>
