<script lang="ts">
	import { onDestroy } from 'svelte';
	import { Input, cn } from '@facile/muse';
	import type { Task } from '$lib/backend';

	type Props = {
		tasks: Task[];
		value: string;
		disabled?: boolean;
		placeholder?: string;
		loading?: boolean;
	};

	let {
		tasks,
		value = $bindable(''),
		disabled = false,
		placeholder = 'Choose or create a task',
		loading = false
	}: Props = $props();

	const uid = $props.id();
	const listboxId = `${uid}-listbox`;
	const optionId = (index: number) => `${uid}-option-${index}`;

	let open = $state(false);
	let rootEl = $state<HTMLDivElement | null>(null);
	let inputEl = $state<HTMLInputElement | null>(null);
	let optionEls = $state<Array<HTMLDivElement | null>>([]);
	let activeIndex = $state(-1);
	let menuPlacement = $state<'up' | 'down'>('down');
	let menuMaxHeight = $state(224);
	let menuLeft = $state(0);
	let menuWidth = $state(0);
	let menuOffset = $state(0);

	const GAP = 8;
	const MARGIN = 12;
	const MIN_HEIGHT = 160;
	const MAX_HEIGHT = 224;

	let filtered = $derived(
		value.trim()
			? tasks.filter((t) => t.name.toLowerCase().includes(value.toLowerCase()))
			: tasks
	);

	let showCreate = $derived(
		value.trim().length > 0 && !tasks.some((t) => t.name.toLowerCase() === value.trim().toLowerCase())
	);
	let optionCount = $derived(filtered.length + (showCreate ? 1 : 0));

	function select(name: string) {
		value = name;
		open = false;
		activeIndex = -1;
		inputEl?.blur();
	}

	function openMenu() {
		if (disabled) {
			return;
		}
		open = true;
		updateMenuLayout();
		if (filtered.length > 0) {
			activeIndex = 0;
			return;
		}
		activeIndex = showCreate ? filtered.length : -1;
	}

	function updateMenuLayout() {
		const anchor = inputEl ?? rootEl;
		if (!anchor || typeof window === 'undefined') {
			return;
		}
		const box = anchor.getBoundingClientRect();
		const below = window.innerHeight - box.bottom - GAP - MARGIN;
		const above = box.top - GAP - MARGIN;
		const shouldOpenUp = below < MIN_HEIGHT && above > below;
		menuPlacement = shouldOpenUp ? 'up' : 'down';
		menuMaxHeight = Math.max(96, Math.min(MAX_HEIGHT, Math.floor(shouldOpenUp ? above : below)));
		menuLeft = box.left;
		menuWidth = box.width;
		menuOffset = shouldOpenUp ? window.innerHeight - box.top + GAP : box.bottom + GAP;
	}

	function handleFocus(event: FocusEvent) {
		inputEl = event.currentTarget as HTMLInputElement;
		openMenu();
	}

	function handleInput() {
		openMenu();
	}

	function handleBlur() {
		setTimeout(() => {
			open = false;
			activeIndex = -1;
		}, 120);
	}

	function handleKeydown(e: KeyboardEvent) {
		if ((e.key === 'ArrowDown' || e.key === 'ArrowUp') && optionCount > 0) {
			e.preventDefault();
			if (!open) {
				openMenu();
				return;
			}
			const delta = e.key === 'ArrowDown' ? 1 : -1;
			activeIndex = activeIndex === -1
				? (delta > 0 ? 0 : optionCount - 1)
				: (activeIndex + delta + optionCount) % optionCount;
			return;
		}
		if (e.key === 'Enter' && open) {
			if (activeIndex >= 0 && activeIndex < filtered.length) {
				e.preventDefault();
				select(filtered[activeIndex].name);
				return;
			}
			if (showCreate && activeIndex === filtered.length) {
				e.preventDefault();
				select(value.trim());
				return;
			}
		}
		if (e.key === 'Tab') {
			open = false;
			activeIndex = -1;
			return;
		}
		if (e.key === 'Escape') {
			e.stopPropagation();
			open = false;
			activeIndex = -1;
			inputEl?.focus();
		}
	}

	function handleWindowLayoutChange() {
		if (!open) {
			return;
		}
		updateMenuLayout();
	}

	$effect(() => {
		if (!open || activeIndex < 0) {
			return;
		}
		optionEls[activeIndex]?.scrollIntoView({
			block: 'nearest'
		});
	});

	$effect(() => {
		if (!open || typeof window === 'undefined') {
			return;
		}
		updateMenuLayout();
		window.addEventListener('resize', handleWindowLayoutChange);
		window.addEventListener('scroll', handleWindowLayoutChange, true);
		return () => {
			window.removeEventListener('resize', handleWindowLayoutChange);
			window.removeEventListener('scroll', handleWindowLayoutChange, true);
		};
	});

	onDestroy(() => {
		optionEls = [];
	});

	const menuVisible = $derived(open && (!loading || filtered.length > 0 || showCreate || Boolean(value.trim())));
</script>

<div bind:this={rootEl} class="relative w-full">
	<Input
		bind:value
		{disabled}
		{placeholder}
		onfocus={handleFocus}
		oninput={handleInput}
		onblur={handleBlur}
		onkeydown={handleKeydown}
		autocomplete="off"
		title={value.trim() || placeholder}
		role="combobox"
		aria-autocomplete="list"
		aria-expanded={menuVisible}
		aria-controls={listboxId}
		aria-activedescendant={menuVisible && activeIndex >= 0 ? optionId(activeIndex) : undefined}
	/>

	{#if menuVisible}
		<div
			class="fixed z-40 flex flex-col overflow-hidden rounded-fc-md border border-fc-border bg-fc-component shadow-lg"
			style:left="{menuLeft}px"
			style:width="{menuWidth}px"
			style:top={menuPlacement === 'down' ? `${menuOffset}px` : undefined}
			style:bottom={menuPlacement === 'up' ? `${menuOffset}px` : undefined}
			style:max-height="{menuMaxHeight}px"
		>
			<div
				id={listboxId}
				role="listbox"
				aria-label="Tasks"
				class="min-h-0 flex-1 overflow-y-auto overscroll-contain p-1"
			>
				{#each filtered as task, index (task.id)}
					<div
						bind:this={optionEls[index]}
						id={optionId(index)}
						role="option"
						aria-selected={value.toLowerCase() === task.name.toLowerCase()}
						tabindex="-1"
						class={cn(
							'flex w-full cursor-default items-center rounded-fc-sm px-2.5 py-2 text-left text-fc-sm transition-colors',
							activeIndex === index || value.toLowerCase() === task.name.toLowerCase()
								? 'bg-fc-accent text-fc-accent-fg font-medium'
								: 'text-fc-fg hover:bg-fc-surface'
						)}
						title={task.name}
						onmouseenter={() => (activeIndex = index)}
						onmousedown={(event) => {
							event.preventDefault();
							select(task.name);
						}}
					>
						<span class="block min-w-0 flex-1 truncate">{task.name}</span>
					</div>
				{/each}
				{#if showCreate}
					<div
						bind:this={optionEls[filtered.length]}
						id={optionId(filtered.length)}
						role="option"
						aria-selected="false"
						tabindex="-1"
						class={cn(
							'flex w-full cursor-default items-center gap-2 rounded-fc-sm px-2.5 py-2 text-left text-fc-sm transition-colors',
							activeIndex === filtered.length
								? 'bg-fc-accent text-fc-accent-fg font-medium'
								: 'text-fc-fg-muted hover:bg-fc-surface hover:text-fc-fg'
						)}
						title={value.trim()}
						onmouseenter={() => (activeIndex = filtered.length)}
						onmousedown={(event) => {
							event.preventDefault();
							select(value.trim());
						}}
					>
						<span class="min-w-0 flex-1 truncate">Create "{value.trim()}"</span>
					</div>
				{:else if !loading && filtered.length === 0}
					<div class="px-2.5 py-2 text-fc-sm text-fc-fg-muted">
						No matching task
					</div>
				{:else if loading}
					<div class="px-2.5 py-2 text-fc-sm text-fc-fg-muted">
						Loading tasks…
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>
