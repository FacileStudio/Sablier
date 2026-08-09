<script lang="ts">
	import { page } from '$app/state';
	import { Divider, Tabs, icons } from '@facile/muse';

	let { children } = $props();

	const sections = [
		{ id: 'profile', label: 'Profile', icon: icons.userCircle, href: '/settings/profile' },
		{ id: 'appearance', label: 'Appearance', icon: icons.palette, href: '/settings/appearance' },
		{ id: 'api', label: 'API', icon: icons.key, href: '/settings/api' },
		{ id: 'integrations', label: 'Integrations', icon: icons.plug, href: '/settings/integrations' },
		{ id: 'advanced', label: 'Advanced', icon: icons.settings, href: '/settings/advanced' }
	];

	const active = $derived(
		sections.find((section) => page.url.pathname.startsWith(section.href))?.id ?? 'profile'
	);
</script>

<div class="mx-auto flex w-full max-w-fc-lg flex-col gap-10 p-6 md:pt-24">
	<div class="flex flex-col gap-4">
		<Tabs items={sections} value={active} />
		<Divider class="my-0" />
	</div>

	{@render children()}
</div>
