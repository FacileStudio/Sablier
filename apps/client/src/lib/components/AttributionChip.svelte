<script lang="ts">
	import { Badge } from '@facile/muse';
	import type { AttributionSource } from '$lib/backend';
	import { t } from '$lib/i18n/index.svelte';

	type Props = {
		source?: AttributionSource | null;
		confidence?: number | null;
		booked?: boolean;
	};

	let { source = null, confidence = null, booked = false }: Props = $props();

	const BOOKED_SOURCES = new Set<AttributionSource>(['ai', 'opus', 'fallback']);

	const visible = $derived(!booked || (source !== null && BOOKED_SOURCES.has(source)));

	const tone = $derived(
		source === 'ai' ? 'info' : source === 'opus' ? 'success' : !source || source === 'fallback' ? 'warning' : 'neutral'
	);

	const label = $derived.by(() => {
		if (source === 'ai') return t('attribution.ai', { percent: Math.round((confidence ?? 0) * 100) });
		if (!source || source === 'fallback') return t('attribution.fallback');
		return t(`attribution.${source}`);
	});
</script>

{#if visible}
	<Badge {tone} title={t('attribution.title')}>{label}</Badge>
{/if}
