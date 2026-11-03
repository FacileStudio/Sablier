<script lang="ts">
	import { Badge } from '@facile/muse';

	let { class: className = '' }: { class?: string } = $props();
</script>

<Badge tone="info" class="px-1.5 py-0 leading-tight {className}" title="AI agent, internal time only">agent</Badge>
