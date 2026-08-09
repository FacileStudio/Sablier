<script lang="ts">
	import { tick } from 'svelte';
	import { Input, cn, icons as museIcons } from '@facile/muse';
	import { PROJECT_ICONS, toIconify } from '$lib/icons';

	let { value = 'Layout', onSelect }: { value?: string; onSelect: (icon: string) => void } = $props();

	const COLUMNS = 8;

	let search = $state('');
	let open = $state(false);
	let triggerEl = $state<HTMLButtonElement | null>(null);
	let dropdownStyle = $state('');
	let cellEls: HTMLButtonElement[] = [];

	const filtered = $derived(
		search
			? PROJECT_ICONS.filter((name) => name.toLowerCase().includes(search.toLowerCase()))
			: PROJECT_ICONS
	);

	const selectedIndex = $derived(filtered.indexOf(value));
	const focusIndex = $derived(selectedIndex < 0 ? 0 : selectedIndex);

	function close() {
		open = false;
		search = '';
		triggerEl?.focus();
	}

	async function toggle() {
		if (open) {
			close();
			return;
		}
		if (triggerEl) {
			const rect = triggerEl.getBoundingClientRect();
			const spaceBelow = window.innerHeight - rect.bottom;
			const dropdownHeight = 320;
			if (spaceBelow < dropdownHeight) {
				dropdownStyle = `position:fixed;left:${rect.left}px;bottom:${window.innerHeight - rect.top + 4}px;`;
			} else {
				dropdownStyle = `position:fixed;left:${rect.left}px;top:${rect.bottom + 4}px;`;
			}
		}
		open = true;
		await tick();
		cellEls[focusIndex]?.focus();
	}

	function focusAt(index: number) {
		const icon = filtered[index];
		if (icon === undefined) return;
		onSelect(icon);
		cellEls[index]?.focus();
		cellEls[index]?.scrollIntoView({ block: 'nearest' });
	}

	function move(delta: number) {
		const n = filtered.length;
		if (n === 0) return;
		if (selectedIndex < 0) return focusAt(delta > 0 ? 0 : n - 1);
		focusAt((selectedIndex + delta + n * COLUMNS) % n);
	}

	function handleKeydown(event: KeyboardEvent) {
		switch (event.key) {
			case 'ArrowRight':
				event.preventDefault();
				move(1);
				break;
			case 'ArrowLeft':
				event.preventDefault();
				move(-1);
				break;
			case 'ArrowDown':
				event.preventDefault();
				move(COLUMNS);
				break;
			case 'ArrowUp':
				event.preventDefault();
				move(-COLUMNS);
				break;
			case 'Home':
				event.preventDefault();
				focusAt(0);
				break;
			case 'End':
				event.preventDefault();
				focusAt(filtered.length - 1);
				break;
		}
	}

	function choose(index: number) {
		const icon = filtered[index];
		if (icon === undefined) return;
		onSelect(icon);
		close();
	}
</script>

<svelte:window
	onkeydown={(e) => {
		if (open && e.key === 'Escape') {
			e.preventDefault();
			close();
		}
	}}
/>

{#if open}
	<button
		type="button"
		aria-label="Close icon picker"
		class="fixed inset-0 z-40 cursor-default"
		onclick={close}
	></button>
{/if}

<div>
	<button
		bind:this={triggerEl}
		type="button"
		aria-haspopup="dialog"
		aria-expanded={open}
		class="flex h-11 items-center gap-2 rounded-fc-md border border-fc-border bg-fc-bg px-3 text-fc-sm text-fc-fg transition-colors hover:bg-fc-surface focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
		onclick={toggle}
	>
		<iconify-icon icon={toIconify(value)} width="18" height="18" class="block"></iconify-icon>
		<span class="text-fc-fg-muted">{value || 'Select icon'}</span>
		<iconify-icon icon={museIcons.chevronDown} width="14" height="14" class="block shrink-0"
		></iconify-icon>
	</button>

	{#if open}
		<div class="z-50 w-80 rounded-fc-lg border border-fc-border bg-fc-component p-3 shadow-lg" style={dropdownStyle}>
			<Input type="text" placeholder="Search icons..." bind:value={search} class="mb-2" />
			<div
				role="radiogroup"
				tabindex={-1}
				aria-label="Project icon"
				class="grid max-h-60 grid-cols-8 gap-1 overflow-y-auto"
				onkeydown={handleKeydown}
			>
				{#each filtered as icon, i (icon)}
					{@const selected = i === selectedIndex}
					<button
						bind:this={cellEls[i]}
						type="button"
						role="radio"
						aria-checked={selected}
						aria-label={icon}
						title={icon}
						tabindex={i === focusIndex ? 0 : -1}
						class={cn(
							'flex h-9 w-9 items-center justify-center rounded-fc-md transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring',
							selected
								? 'bg-fc-accent text-fc-accent-fg'
								: 'text-fc-fg-muted hover:bg-fc-surface hover:text-fc-fg'
						)}
						onclick={() => choose(i)}
					>
						<iconify-icon icon={toIconify(icon)} width="18" height="18" class="block"></iconify-icon>
					</button>
				{/each}
			</div>
		</div>
	{/if}
</div>
