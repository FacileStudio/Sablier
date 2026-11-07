<script lang="ts">
	import { Badge } from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';

	let { class: className = '' }: { class?: string } = $props();
</script>

<Badge tone="info" class="px-1.5 py-0 leading-tight {className}" title={t('agents.badgeTitle')}>{t('agents.badge')}</Badge>
