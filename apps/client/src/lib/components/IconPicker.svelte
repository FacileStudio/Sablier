<script lang="ts">
	import { tick } from 'svelte';
	import { Input, cn, icons as museIcons } from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';
	import { PROJECT_ICONS, toIconify } from '$lib/icons';

	let { value = 'Layout', onSelect }: { value?: string; onSelect: (icon: string) => void } = $props();

	const MARGIN = 8;
	const CELL = 36;
	const GAP = 4;
	const PADDING = 12;
	const SEARCH = 52;
	const PANEL_WIDTH = 320;
	const PANEL_HEIGHT = 340;

	let search = $state('');
	let open = $state(false);
	let triggerEl = $state<HTMLButtonElement | null>(null);
	let panelEl = $state<HTMLDivElement | null>(null);
	let place = $state({ left: 0, top: 0, bottom: 0, up: false, width: PANEL_WIDTH, gridHeight: 240 });
	let cellEls: HTMLButtonElement[] = [];

	const columns = $derived(Math.max(1, Math.floor((place.width - 2 * PADDING + GAP) / (CELL + GAP))));

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

	function measure() {
		if (!triggerEl) return;
		const rect = triggerEl.getBoundingClientRect();
		const width = Math.min(PANEL_WIDTH, window.innerWidth - 2 * MARGIN);
		const left = Math.min(Math.max(rect.left, MARGIN), window.innerWidth - width - MARGIN);
		const below = window.innerHeight - rect.bottom - MARGIN;
		const above = rect.top - MARGIN;
		const up = below < PANEL_HEIGHT && above > below;
		const room = Math.min(PANEL_HEIGHT, up ? above : below) - 4;
		place = {
			left,
			top: rect.bottom + 4,
			bottom: window.innerHeight - rect.top + 4,
			up,
			width,
			gridHeight: Math.max(CELL * 2, room - SEARCH - 2 * PADDING)
		};
	}

	function present(node: HTMLElement) {
		if (node.showPopover && !node.matches(':popover-open')) node.showPopover();
	}

	async function toggle() {
		if (open) {
			close();
			return;
		}
		measure();
		open = true;
		await tick();
		cellEls[focusIndex]?.focus();
	}

	function closeOnOutside(event: PointerEvent) {
		const target = event.target as Node;
		if (open && !panelEl?.contains(target) && !triggerEl?.contains(target)) close();
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
		focusAt((selectedIndex + delta + n * columns) % n);
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
				move(columns);
				break;
			case 'ArrowUp':
				event.preventDefault();
				move(-columns);
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
	onpointerdown={closeOnOutside}
	onresize={() => open && close()}
/>

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
		<span class="text-fc-fg-muted">{value || t('iconPicker.select')}</span>
		<iconify-icon icon={museIcons.chevronDown} width="14" height="14" class="block shrink-0"
		></iconify-icon>
	</button>

	{#if open}
		<div
			bind:this={panelEl}
			popover="manual"
			class="fixed inset-auto m-0 rounded-fc-lg bg-fc-component p-3 text-fc-fg shadow-lg"
			style:left="{place.left}px"
			style:width="{place.width}px"
			style:top={place.up ? 'auto' : `${place.top}px`}
			style:bottom={place.up ? `${place.bottom}px` : 'auto'}
			use:present
		>
			<Input type="text" placeholder={t('iconPicker.search')} bind:value={search} class="mb-2" />
			<div
				role="radiogroup"
				tabindex={-1}
				aria-label={t('iconPicker.label')}
				class="grid grid-cols-[repeat(auto-fill,2.25rem)] justify-between gap-1 overflow-y-auto overscroll-contain"
				style:max-height="{place.gridHeight}px"
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
							'flex size-9 items-center justify-center rounded-fc-md transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring',
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
