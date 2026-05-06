<script lang="ts">
	import { normalizeUserColor } from '$lib/user-colors';
	import { cn } from '$lib/utils';

	let {
		name,
		avatarUrl = '',
		color,
		class: className = ''
	}: {
		name: string;
		avatarUrl?: string;
		color?: string;
		class?: string;
	} = $props();

	function getInitials(value: string) {
		return value
			.trim()
			.split(/\s+/)
			.slice(0, 2)
			.map((part) => part[0]?.toUpperCase() ?? '')
			.join('') || '?';
	}
</script>

{#if avatarUrl}
	<img
		src={avatarUrl}
		alt={name}
		class={cn('h-8 w-8 shrink-0 rounded-full object-cover ring-1 ring-black/10', className)}
	/>
{:else}
	<div
		class={cn(
			'flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-xs font-semibold text-white ring-1 ring-black/10',
			className
		)}
		style={`background-color: ${normalizeUserColor(color)};`}
	>
		{getInitials(name)}
	</div>
{/if}
