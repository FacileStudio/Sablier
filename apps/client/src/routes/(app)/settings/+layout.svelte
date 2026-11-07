<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { SettingsLayout, icons, settingsRoute } from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';

	let { children: panel } = $props();

	const sections = $derived([
		{ id: 'profile', label: t('settings.tab.profile'), icon: icons.userCircle, href: '/settings/profile' },
		{ id: 'preferences', label: t('settings.tab.preferences'), icon: icons.palette, href: '/settings/preferences' },
		{ id: 'tracking', label: t('settings.tab.tracking'), icon: icons.clock, href: '/settings/tracking' },
		{ id: 'integrations', label: t('settings.tab.integrations'), icon: icons.plug, href: '/settings/integrations' },
		{ id: 'tokens', label: t('settings.tab.tokens'), icon: icons.key, href: '/settings/tokens' },
		{ id: 'advanced', label: t('settings.tab.advanced'), icon: icons.settings, href: '/settings/advanced' }
	]);

	const requested = $derived(page.url.pathname.split('/')[2] || undefined);
	const route = $derived(settingsRoute(requested, sections));

	$effect(() => {
		if (!route.known) void goto('/settings', { replaceState: true });
	});
</script>

<div class="mx-auto w-full max-w-fc-lg">
	<SettingsLayout {sections} active={requested} backHref="/settings">
		{#snippet children()}
			{@render panel()}
		{/snippet}
	</SettingsLayout>
</div>
