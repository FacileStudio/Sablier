<script lang="ts">
	import { OptionCards, SettingsRow, SettingsSection, icons } from '@facile/muse';
	import { theme, type ThemePreference } from '$lib/theme.svelte';

	const options = [
		{ value: 'system', label: 'System', icon: icons.monitor },
		{ value: 'light', label: 'Light', icon: icons.sun },
		{ value: 'dark', label: 'Dark', icon: icons.moon }
	];
</script>

<svelte:head>
	<title>Appearance — Sablier</title>
</svelte:head>

<SettingsSection title="Appearance" description="Applied to this browser only.">
	<SettingsRow
		stacked
		label="Colour scheme"
		description="System follows your operating system setting and changes with it."
	>
		<OptionCards
			label="Colour scheme"
			{options}
			value={theme.preference}
			onSelect={(next) => theme.set(next as ThemePreference)}
		/>
	</SettingsRow>
</SettingsSection>
