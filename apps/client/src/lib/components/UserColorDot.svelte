<script lang="ts">
	import { normalizeUserColor } from '$lib/user-colors';

	let { color, class: className = '' }: { color?: string; class?: string } = $props();
</script>

<span
	class={`inline-block h-2.5 w-2.5 shrink-0 rounded-full border border-black/5 ${className}`}
	style={`background-color: ${normalizeUserColor(color)};`}
	aria-hidden="true"
></span>
