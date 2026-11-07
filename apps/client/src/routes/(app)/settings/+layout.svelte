<script lang="ts">
	import { page } from '$app/state';
	import { Divider, Tabs, icons } from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';

	let { children } = $props();

	const sections = $derived([
		{ id: 'profile', label: t('settings.tab.profile'), icon: icons.userCircle, href: '/settings/profile' },
		{ id: 'appearance', label: t('settings.tab.appearance'), icon: icons.palette, href: '/settings/appearance' },
		{ id: 'api', label: t('settings.tab.api'), icon: icons.key, href: '/settings/api' },
		{ id: 'integrations', label: t('settings.tab.integrations'), icon: icons.plug, href: '/settings/integrations' },
		{ id: 'advanced', label: t('settings.tab.advanced'), icon: icons.settings, href: '/settings/advanced' }
	]);

	const active = $derived(
		sections.find((section) => page.url.pathname.startsWith(section.href))?.id ?? 'profile'
	);
</script>

<div class="mx-auto flex w-full max-w-fc-lg flex-col gap-10">
	<div class="flex flex-col gap-4">
		<Tabs items={sections} value={active} />
		<Divider />
	</div>

	{@render children()}
</div>
