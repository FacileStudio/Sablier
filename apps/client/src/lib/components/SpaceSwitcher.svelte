<script lang="ts">
	import { goto } from '$app/navigation';
	import type { Space } from '$lib/backend';
	import { getActiveSpaceId, setActiveSpaceId } from '$lib/space-context.svelte';

	let { spaces }: { spaces: Space[] } = $props();

	let open = $state(false);
	let activeId = $derived(getActiveSpaceId());
	let activeSpace = $derived(spaces.find((s) => s.id === activeId) ?? null);

	function select(id: string | null) {
		setActiveSpaceId(id);
		open = false;
		window.location.reload();
	}

	function handleClickOutside(e: MouseEvent) {
		const target = e.target as HTMLElement;
		if (!target.closest('.space-switcher')) {
			open = false;
		}
	}

	$effect(() => {
		if (open) {
			document.addEventListener('click', handleClickOutside);
			return () => document.removeEventListener('click', handleClickOutside);
		}
	});
</script>

<div class="space-switcher relative px-3 pb-2">
	<button
		type="button"
		class="flex w-full items-center gap-2.5 rounded-lg border border-border/60 bg-muted/30 px-3 py-2 text-left text-sm transition-colors hover:bg-muted/60"
		onclick={() => (open = !open)}
	>
		<iconify-icon
			icon={activeSpace ? 'solar:users-group-rounded-bold-duotone' : 'solar:user-circle-bold-duotone'}
			width="18"
			class="shrink-0 text-muted-foreground"
		></iconify-icon>
		<span class="min-w-0 flex-1 truncate font-medium">
			{activeSpace?.name ?? 'Tous les espaces'}
		</span>
		<iconify-icon
			icon="solar:alt-arrow-down-linear"
			width="14"
			class="shrink-0 text-muted-foreground transition-transform {open ? 'rotate-180' : ''}"
		></iconify-icon>
	</button>

	{#if open}
		<div class="absolute left-3 right-3 z-50 mt-1 overflow-hidden rounded-lg border border-border bg-background shadow-lg">
			<div class="max-h-64 overflow-auto p-1">
				<button
					type="button"
					class="flex w-full items-center gap-2.5 rounded-md px-3 py-2 text-left text-sm transition-colors {!activeId ? 'bg-foreground text-background' : 'text-foreground hover:bg-muted'}"
					onclick={() => select(null)}
				>
					<iconify-icon icon="solar:user-circle-bold-duotone" width="16" class="shrink-0"></iconify-icon>
					Tous les espaces
				</button>

				{#each spaces as space}
					<button
						type="button"
						class="flex w-full items-center gap-2.5 rounded-md px-3 py-2 text-left text-sm transition-colors {activeId === space.id ? 'bg-foreground text-background' : 'text-foreground hover:bg-muted'}"
						onclick={() => select(space.id)}
					>
						<iconify-icon icon="solar:users-group-rounded-bold-duotone" width="16" class="shrink-0"></iconify-icon>
						<span class="min-w-0 flex-1 truncate">{space.name}</span>
					</button>
				{/each}
			</div>

			<div class="border-t border-border p-1">
				<a
					href="/spaces"
					class="flex w-full items-center gap-2.5 rounded-md px-3 py-2 text-left text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
					onclick={() => (open = false)}
				>
					<iconify-icon icon="solar:settings-linear" width="16" class="shrink-0"></iconify-icon>
					Gérer les espaces
				</a>
			</div>
		</div>
	{/if}
</div>
