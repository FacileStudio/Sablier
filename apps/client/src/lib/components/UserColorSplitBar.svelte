<script lang="ts">
	import { cn, normalizeUserColor } from '@facile/muse';
	import UserColorDot from '$lib/components/UserColorDot.svelte';
	import UserAvatarBadge from '$lib/components/UserAvatarBadge.svelte';
	import { formatDuration } from '$lib/utils';

	type Segment = {
		key: string;
		label: string;
		color?: string;
		avatarUrl?: string;
		ms: number;
	};

	let {
		segments = [],
		barClass = 'h-3',
		showLegend = true,
		showAvatars = false,
		showDuration = false
	}: {
		segments?: Segment[];
		barClass?: string;
		showLegend?: boolean;
		showAvatars?: boolean;
		showDuration?: boolean;
	} = $props();

	const totalMs = $derived(segments.reduce((sum, segment) => sum + segment.ms, 0));

	function sharePercent(ms: number) {
		if (totalMs === 0) {
			return 0;
		}
		return Math.round((ms / totalMs) * 100);
	}
</script>

<div class="flex flex-col gap-3">
	<div class={cn('flex w-full gap-[3px]', barClass)}>
		{#if totalMs > 0}
			{#each segments as segment (segment.key)}
				<div
					class="h-full min-w-1.5 rounded-fc-pill"
					style:width={`${(segment.ms / totalMs) * 100}%`}
					style:background-color={normalizeUserColor(segment.color)}
					title={`${segment.label}: ${sharePercent(segment.ms)}%`}
					aria-label={`${segment.label}: ${sharePercent(segment.ms)}%`}
				></div>
			{/each}
		{:else}
			<div class="h-full w-full rounded-fc-pill bg-fc-surface"></div>
		{/if}
	</div>

	{#if showLegend && segments.length > 0}
		<div
			class={cn(
				'flex flex-wrap text-fc-xs text-fc-fg-muted',
				showAvatars ? 'gap-x-5 gap-y-3' : 'gap-x-4 gap-y-2'
			)}
		>
			{#each segments as segment (segment.key)}
				<div class={cn('flex items-center', showAvatars ? 'gap-2' : 'gap-1.5')}>
					{#if showAvatars}
						<UserAvatarBadge
							name={segment.label}
							avatarUrl={segment.avatarUrl}
							color={segment.color}
							class="h-6 w-6 text-fc-xs"
						/>
					{/if}
					<UserColorDot color={segment.color} class="h-2.5 w-2.5" />
					<span class="font-medium text-fc-fg">{segment.label}</span>
					{#if showDuration}
						<span class="font-mono tabular-nums">{formatDuration(segment.ms)}</span>
						<span class="text-fc-fg-muted/60">&middot;</span>
					{/if}
					<span>{sharePercent(segment.ms)}%</span>
				</div>
			{/each}
		</div>
	{/if}
</div>
