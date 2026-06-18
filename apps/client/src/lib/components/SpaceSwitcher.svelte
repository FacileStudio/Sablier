<script lang="ts">
	import { goto } from '$app/navigation';
	import type { Space } from '$lib/backend';
	import { getActiveSpaceId, setActiveSpaceId } from '$lib/space-context.svelte';
	import { Users, User, ChevronDown } from 'lucide-svelte';

	let { spaces }: { spaces: Space[] } = $props();

	let open = $state(false);
	let activeId = $derived(getActiveSpaceId());
	let activeSpace = $derived(spaces.find((s) => s.id === activeId) ?? null);

	function select(id: string | null) {
		setActiveSpaceId(id);
		open = false;
		window.location.reload();
	}
</script>

<svelte:window
	onclick={(e) => {
		if (open && !(e.target as HTMLElement).closest('.space-switcher')) {
			open = false;
		}
	}}
/>

<div class="space-switcher relative px-3 pb-2">
	<button
		type="button"
		class="flex w-full items-center gap-2.5 rounded-lg border border-border/60 bg-muted/30 px-3 py-2 text-sm transition-colors hover:bg-muted/60"
		onclick={() => (open = !open)}
	>
		{#if activeSpace}
			<Users class="h-4 w-4 shrink-0 text-muted-foreground" />
		{:else}
			<User class="h-4 w-4 shrink-0 text-muted-foreground" />
		{/if}
		<span class="flex-1 truncate text-left font-medium">
			{activeSpace?.name ?? 'Tous les espaces'}
		</span>
		<ChevronDown class="h-3.5 w-3.5 shrink-0 text-muted-foreground transition-transform {open ? 'rotate-180' : ''}" />
	</button>

	{#if open}
		<div class="absolute left-3 right-3 z-50 mt-1 max-h-64 overflow-auto rounded-lg border border-border bg-background p-1 shadow-lg">
			<button
				type="button"
				class="flex w-full items-center gap-2.5 rounded-md px-3 py-2 text-sm transition-colors {!activeId ? 'bg-foreground text-background' : 'hover:bg-muted'}"
				onclick={() => select(null)}
			>
				<User class="h-4 w-4 shrink-0" />
				<span class="truncate">Tous les espaces</span>
			</button>
			{#each spaces as space}
				<button
					type="button"
					class="flex w-full items-center gap-2.5 rounded-md px-3 py-2 text-sm transition-colors {activeId === space.id ? 'bg-foreground text-background' : 'hover:bg-muted'}"
					onclick={() => select(space.id)}
				>
					<Users class="h-4 w-4 shrink-0" />
					<span class="truncate">{space.name}</span>
				</button>
			{/each}
			<div class="mt-1 border-t border-border pt-1">
				<button
					type="button"
					class="flex w-full items-center gap-2.5 rounded-md px-3 py-2 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
					onclick={() => { open = false; goto('/spaces'); }}
				>
					Gérer les espaces
				</button>
			</div>
		</div>
	{/if}
</div>
