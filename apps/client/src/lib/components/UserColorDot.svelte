<script lang="ts">
	import { cn, normalizeUserColor } from '@facile/muse';

	let { color, class: className = '' }: { color?: string; class?: string } = $props();
</script>

<span
	class={cn('inline-block h-2.5 w-2.5 shrink-0 rounded-fc-pill', className)}
	style:background-color={normalizeUserColor(color)}
	aria-hidden="true"
></span>
