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

	let filtered = $derived(
		value.trim()
			? tasks.filter((t) => t.name.toLowerCase().includes(value.toLowerCase()))
			: tasks
	);

	let showCreate = $derived(
		value.trim().length > 0 && !tasks.some((t) => t.name.toLowerCase() === value.trim().toLowerCase())
	);

	function select(name: string) {
		value = name;
		open = false;
		inputEl?.blur();
	}

	function handleFocus() {
		if (!disabled) open = true;
	}

	function handleInput() {
		open = true;
	}

	function handleBlur() {
		setTimeout(() => {
			open = false;
		}, 120);
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			open = false;
			inputEl?.blur();
		}
	}
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
			'dark:bg-input/30 border-input focus-visible:border-ring focus-visible:ring-ring/50 disabled:bg-input/50 dark:disabled:bg-input/80 h-8 w-full min-w-0 truncate rounded-lg border bg-transparent px-2.5 py-1 text-base outline-none transition-colors focus-visible:ring-3 disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50 placeholder:text-muted-foreground md:text-sm'
		)}
	/>

	{#if open && !loading && (filtered.length > 0 || showCreate)}
		<div
			class="absolute left-0 right-0 top-full z-50 mt-1 max-h-48 overflow-y-auto rounded-lg border bg-popover text-popover-foreground shadow-md"
		>
			{#each filtered as task}
				<button
					type="button"
					class={cn(
						'flex w-full cursor-default items-center px-3 py-2 text-sm hover:bg-accent hover:text-accent-foreground',
						value.toLowerCase() === task.name.toLowerCase() && 'bg-accent text-accent-foreground'
					)}
					title={task.name}
					onmousedown={() => select(task.name)}
				>
					<span class="block min-w-0 truncate">{task.name}</span>
				</button>
			{/each}
			{#if showCreate}
				<button
					type="button"
					class="flex w-full cursor-default items-center gap-1.5 border-t px-3 py-2 text-sm text-muted-foreground hover:bg-accent hover:text-accent-foreground"
					title={value.trim()}
					onmousedown={() => select(value.trim())}
				>
					<span class="text-xs">Create</span>
					<span class="min-w-0 truncate font-medium text-foreground">"{value.trim()}"</span>
				</button>
			{/if}
		</div>
	{/if}
</div>
