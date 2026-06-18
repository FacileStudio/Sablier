<script lang="ts">
	import { goto } from '$app/navigation';
	import type { Space } from '$lib/backend';
	import { getActiveSpaceId, setActiveSpaceId } from '$lib/space-context';
	import { Button } from '$lib/components/ui/button';
	import { Separator } from '$lib/components/ui/separator';
	import { Building2, ChevronDown, Check, Plus, Settings } from 'lucide-svelte';

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

<div class="relative px-3 mb-2">
	<button
		type="button"
		class="flex w-full items-center gap-2 rounded-md border border-border/70 bg-muted/40 px-3 py-2 text-sm transition-colors hover:bg-muted"
		onclick={() => (open = !open)}
	>
		<Building2 class="h-4 w-4 shrink-0 text-muted-foreground" />
		<span class="flex-1 truncate text-left font-medium">
			{activeSpace?.name ?? 'Tous les espaces'}
		</span>
		<ChevronDown class="h-3.5 w-3.5 shrink-0 text-muted-foreground transition-transform {open ? 'rotate-180' : ''}" />
	</button>

	{#if open}
		<div class="absolute left-3 right-3 top-full z-50 mt-1 flex flex-col rounded-md border bg-popover shadow-md">
			<button
				type="button"
				class="flex items-center gap-2 px-3 py-2 text-sm transition-colors hover:bg-accent {!activeId ? 'font-medium' : ''}"
				onclick={() => select(null)}
			>
				{#if !activeId}<Check class="h-3.5 w-3.5" />{:else}<span class="w-3.5"></span>{/if}
				Tous les espaces
			</button>
			<Separator />
			{#each spaces as space}
				<button
					type="button"
					class="flex items-center gap-2 px-3 py-2 text-sm transition-colors hover:bg-accent {activeId === space.id ? 'font-medium' : ''}"
					onclick={() => select(space.id)}
				>
					{#if activeId === space.id}<Check class="h-3.5 w-3.5" />{:else}<span class="w-3.5"></span>{/if}
					<span class="truncate">{space.name}</span>
				</button>
			{/each}
			<Separator />
			<button
				type="button"
				class="flex items-center gap-2 px-3 py-2 text-sm text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
				onclick={() => { open = false; goto('/spaces'); }}
			>
				<Settings class="h-3.5 w-3.5" />
				Gérer les espaces
			</button>
			<button
				type="button"
				class="flex items-center gap-2 px-3 py-2 text-sm text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
				onclick={() => { open = false; goto('/spaces/new'); }}
			>
				<Plus class="h-3.5 w-3.5" />
				Nouvel espace
			</button>
		</div>
	{/if}
</div>
