<script lang="ts">
	import UserColorDot from '$lib/components/UserColorDot.svelte';
	import UserAvatarBadge from '$lib/components/UserAvatarBadge.svelte';
	import { normalizeUserColor } from '$lib/user-colors';
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

<div class="space-y-3">
	<div class={`flex w-full gap-[3px] ${barClass}`}>
		{#if totalMs > 0}
			{#each segments as segment (segment.key)}
				<div
					class="h-full rounded-full"
					style={`width: ${(segment.ms / totalMs) * 100}%; min-width: 6px; background-color: ${normalizeUserColor(segment.color)};`}
					title={`${segment.label}: ${sharePercent(segment.ms)}%`}
					aria-label={`${segment.label}: ${sharePercent(segment.ms)}%`}
				></div>
			{/each}
		{:else}
			<div class="h-full w-full rounded-full bg-muted/40"></div>
		{/if}
	</div>

	{#if showLegend && segments.length > 0}
		<div class="flex flex-wrap gap-x-4 gap-y-2 text-xs text-muted-foreground" class:gap-x-5={showAvatars} class:gap-y-3={showAvatars}>
			{#each segments as segment (segment.key)}
				<div class="flex items-center gap-1.5" class:gap-2={showAvatars}>
					{#if showAvatars}
						<UserAvatarBadge
							name={segment.label}
							avatarUrl={segment.avatarUrl}
							color={segment.color}
							class="h-6 w-6 text-[10px]"
						/>
					{/if}
					<UserColorDot color={segment.color} class="h-2.5 w-2.5" />
					<span class="font-medium text-foreground/90">{segment.label}</span>
					{#if showDuration}
						<span class="font-mono tabular-nums">{formatDuration(segment.ms)}</span>
						<span class="text-muted-foreground/60">&middot;</span>
					{/if}
					<span>{sharePercent(segment.ms)}%</span>
				</div>
			{/each}
		</div>
	{/if}
</div>
