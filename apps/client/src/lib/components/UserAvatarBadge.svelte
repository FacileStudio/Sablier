<script lang="ts">
	import { Avatar, normalizeUserColor } from '@facile/muse';

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

	const identity = $derived(normalizeUserColor(color));
</script>

<Avatar
	{name}
	src={avatarUrl || undefined}
	size="sm"
	class={className}
	style="background-color: {identity}"
/>
