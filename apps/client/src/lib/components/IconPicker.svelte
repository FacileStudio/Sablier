<script lang="ts">
	import { PROJECT_ICONS, toIconify } from '$lib/icons';

	let { value = 'Layout', onSelect }: { value?: string; onSelect: (icon: string) => void } = $props();
	let search = $state('');
	let open = $state(false);

	let filtered = $derived(
		search
			? PROJECT_ICONS.filter((name) => name.toLowerCase().includes(search.toLowerCase()))
			: PROJECT_ICONS
	);
</script>

{#if open}
	<div class="fixed inset-0 z-40" onclick={() => (open = false)} onkeydown={() => {}} role="presentation"></div>
{/if}

<div class="relative">
	<button
		type="button"
		class="flex h-9 items-center gap-2 rounded-md border border-input bg-background px-3 text-sm transition-colors hover:bg-muted"
		onclick={() => (open = !open)}
	>
		<iconify-icon icon={toIconify(value)} width="18" height="18"></iconify-icon>
		<span class="text-muted-foreground">{value || 'Select icon'}</span>
	</button>

	{#if open}
		<div class="absolute left-0 top-full z-50 mt-1 w-80 rounded-lg border border-border bg-popover p-3 shadow-lg">
			<input
				type="text"
				placeholder="Search icons..."
				class="mb-2 h-8 w-full rounded-md border border-input bg-background px-3 text-sm"
				bind:value={search}
			/>
			<div class="grid max-h-60 grid-cols-8 gap-1 overflow-y-auto">
				{#each filtered as icon}
					<button
						type="button"
						class="flex h-9 w-9 items-center justify-center rounded-md transition-colors {value === icon ? 'bg-primary text-primary-foreground' : 'hover:bg-muted'}"
						onclick={() => { onSelect(icon); open = false; search = ''; }}
						title={icon}
					>
						<iconify-icon icon={toIconify(icon)} width="18" height="18"></iconify-icon>
					</button>
				{/each}
			</div>
		</div>
	{/if}
</div>
