<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { Select } from '@facile/muse';
	import { PERSONAL, getSpaces, spaceFilter } from '$lib/space-context.svelte';
	import { t } from '$lib/i18n/index.svelte';

	const options = $derived([
		{ value: 'all', label: t('spaceFilter.all') },
		{ value: PERSONAL, label: t('spaceFilter.personal') },
		...getSpaces().map((s) => ({ value: s.id, label: s.name }))
	]);

	function choose(value: string) {
		const url = new URL(page.url);
		if (value === 'all') url.searchParams.delete('space');
		else url.searchParams.set('space', value);
		void goto(url, { replaceState: true, keepFocus: true, noScroll: true });
	}
</script>

<Select
	wrapperClass="w-full sm:w-56"
	aria-label={t('spaceFilter.label')}
	{options}
	bind:value={() => spaceFilter(page.url), choose}
/>
