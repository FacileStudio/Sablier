<script lang="ts">
	import { Avatar, cn, normalizeUserColor } from '@facile/muse';

	let {
		name,
		avatarUrl = '',
		color,
		agent = false,
		class: className = ''
	}: {
		name: string;
		avatarUrl?: string;
		color?: string;
		agent?: boolean;
		class?: string;
	} = $props();

	const identity = $derived(normalizeUserColor(color));
	let failed = $state(false);
	$effect(() => {
		void avatarUrl;
		failed = false;
	});
</script>

{#if agent && avatarUrl && !failed}
	<img
		src={avatarUrl}
		alt={name}
		loading="lazy"
		onerror={() => (failed = true)}
		class={cn('size-8 shrink-0 scale-125 object-contain', className)}
	/>
{:else}
	<Avatar
		{name}
		src={avatarUrl || undefined}
		size="sm"
		class={className}
		style="background-color: {identity}"
	/>
{/if}
