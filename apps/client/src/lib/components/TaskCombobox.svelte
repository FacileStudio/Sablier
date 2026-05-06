<script lang="ts">
	import { cn } from '$lib/utils';
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

	let open = $state(false);
	let inputEl = $state<HTMLInputElement | null>(null);
	let optionEls = $state<Array<HTMLButtonElement | null>>([]);
	let activeIndex = $state(-1);

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
		if (filtered.length > 0) {
			activeIndex = 0;
			return;
		}
		activeIndex = showCreate ? filtered.length : -1;
	}

	function handleFocus() {
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
			open = false;
			activeIndex = -1;
			inputEl?.blur();
		}
	}

	$effect(() => {
		if (!open || activeIndex < 0) {
			return;
		}
		optionEls[activeIndex]?.scrollIntoView({
			block: 'nearest'
		});
	});
</script>

<div class="relative w-full">
	<input
		bind:this={inputEl}
		{disabled}
		{placeholder}
		bind:value
		onfocus={handleFocus}
		oninput={handleInput}
		onblur={handleBlur}
		onkeydown={handleKeydown}
		autocomplete="off"
		title={value.trim() || placeholder}
		class={cn(
			'dark:bg-input/30 border-input focus-visible:border-ring focus-visible:ring-ring/50 disabled:bg-input/50 dark:disabled:bg-input/80 h-10 w-full min-w-0 rounded-xl border bg-transparent px-3 text-sm outline-none transition-colors focus-visible:ring-3 disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50 placeholder:text-muted-foreground'
		)}
	/>

	{#if open && (!loading || filtered.length > 0 || showCreate || value.trim())}
		<div
			class="absolute left-0 right-0 top-full z-50 mt-2 overflow-hidden rounded-xl border border-border/80 bg-popover text-popover-foreground shadow-lg"
		>
			<div class="max-h-[min(14rem,38vh)] overflow-y-auto overscroll-contain p-1">
				{#each filtered as task, index}
					<button
						bind:this={optionEls[index]}
						type="button"
						class={cn(
							'flex w-full cursor-default items-center rounded-lg px-3 py-2 text-left text-sm transition-colors hover:bg-accent hover:text-accent-foreground',
							(activeIndex === index || value.toLowerCase() === task.name.toLowerCase()) && 'bg-accent text-accent-foreground'
						)}
						title={task.name}
						onmouseenter={() => (activeIndex = index)}
						onmousedown={(event) => {
							event.preventDefault();
							select(task.name);
						}}
					>
						<span class="block min-w-0 flex-1 truncate">{task.name}</span>
					</button>
				{/each}
				{#if showCreate}
					<button
						bind:this={optionEls[filtered.length]}
						type="button"
						class={cn(
							'flex w-full cursor-default items-center gap-2 rounded-lg px-3 py-2 text-left text-sm text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground',
							activeIndex === filtered.length && 'bg-accent text-accent-foreground'
						)}
						title={value.trim()}
						onmouseenter={() => (activeIndex = filtered.length)}
						onmousedown={(event) => {
							event.preventDefault();
							select(value.trim());
						}}
					>
						<span class="min-w-0 flex-1 truncate">Create "{value.trim()}"</span>
					</button>
				{:else if !loading && filtered.length === 0}
					<div class="px-3 py-2 text-sm text-muted-foreground">
						No matching task
					</div>
				{:else if loading}
					<div class="px-3 py-2 text-sm text-muted-foreground">
						Loading tasks…
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>
