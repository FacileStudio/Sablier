<script lang="ts">
	import { getContext } from 'svelte';
	import { Badge, Modal, cn, controlLine } from '@facile/muse';
	import { backend, type Project, type TaskSearchResult } from '$lib/backend';
	import { flatOptions, groupResults, type PaletteOption } from '$lib/task-search';
	import { toIconify } from '$lib/icons';
	import { t } from '$lib/i18n/index.svelte';

	type Props = {
		open?: boolean;
		title: string;
		projects: Project[];
		onselect: (result: TaskSearchResult) => Promise<unknown> | void;
	};

	let { open = $bindable(false), title, projects, onselect }: Props = $props();

	const ctx = getContext<{ token: string }>('app');
	const uid = $props.id();
	const listboxId = `${uid}-listbox`;
	const optionId = (index: number) => `${uid}-option-${index}`;

	let query = $state('');
	let results = $state<TaskSearchResult[]>([]);
	let activeIndex = $state(0);
	let loading = $state(false);
	let pending = $state(false);
	let error = $state('');
	let searchSeq = 0;

	const fallback = $derived(projects.map((p) => ({ id: p.id, name: p.name, icon: p.icon, space: null })));
	const groups = $derived(groupResults(results, query, fallback, t('spaceFilter.personal')));
	const options = $derived(flatOptions(groups));

	$effect(() => {
		if (!open) return;
		query = '';
		error = '';
	});

	$effect(() => {
		if (!open) return;
		const q = query.trim();
		const seq = ++searchSeq;
		loading = true;
		const timer = setTimeout(async () => {
			try {
				const found = (await backend.searchTasks(ctx.token, q)).results;
				if (seq === searchSeq) results = found;
			} catch (e) {
				if (seq === searchSeq) error = e instanceof Error ? e.message : t('palette.failed');
			} finally {
				if (seq === searchSeq) loading = false;
				activeIndex = 0;
			}
		}, 150);
		return () => clearTimeout(timer);
	});

	function groupLabel(key: string, label: string | null) {
		if (key === 'recent') return t('palette.recent');
		if (key === 'create') return t('palette.create');
		return label ?? '';
	}

	async function resolve(option: PaletteOption): Promise<TaskSearchResult> {
		if (option.kind === 'task') return option.result;
		const task = await backend.createTask(ctx.token, option.project.id, option.name);
		const { space, ...project } = option.project;
		return { task: { id: task.id, name: task.name, facile_id: null }, project, space, last_used_at: null, opus: false };
	}

	async function choose(option: PaletteOption | undefined) {
		if (!option || pending) return;
		pending = true;
		error = '';
		try {
			await onselect(await resolve(option));
			open = false;
		} catch (e) {
			error = e instanceof Error ? e.message : t('palette.failed');
		} finally {
			pending = false;
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		const count = options.length;
		if (e.key === 'Enter') {
			e.preventDefault();
			void choose(options[activeIndex]);
			return;
		}
		const moves: Record<string, number> = { ArrowDown: activeIndex + 1, ArrowUp: activeIndex - 1, Home: 0, End: count - 1 };
		if (!(e.key in moves) || count === 0) return;
		e.preventDefault();
		activeIndex = (moves[e.key] + count) % count;
		document.getElementById(optionId(activeIndex))?.scrollIntoView({ block: 'nearest' });
	}
</script>

<Modal bind:open {title} size="lg">
	<div class="flex flex-col gap-3">
		<input
			class={controlLine}
			bind:value={query}
			onkeydown={handleKeydown}
			placeholder={t('palette.placeholder')}
			autocomplete="off"
			spellcheck="false"
			role="combobox"
			aria-label={t('palette.label')}
			aria-autocomplete="list"
			aria-expanded="true"
			aria-controls={listboxId}
			aria-activedescendant={options.length ? optionId(activeIndex) : undefined}
			aria-busy={loading || pending}
		/>
		<div
			id={listboxId}
			role="listbox"
			aria-label={t('combobox.tasks')}
			class="max-h-[min(60dvh,26rem)] min-h-24 overflow-y-auto overscroll-contain"
		>
			{#each groups as group (group.key)}
				<div role="group" aria-labelledby="{uid}-{group.key}" class="flex flex-col pb-2">
					<p id="{uid}-{group.key}" class="px-2.5 py-1.5 text-fc-xs font-medium text-fc-fg-muted">
						{groupLabel(group.key, group.label)}
					</p>
					{#each group.options as option, i (i)}
						{@const index = group.start + i}
						<div
							id={optionId(index)}
							role="option"
							tabindex="-1"
							aria-selected={index === activeIndex}
							class={cn(
								'flex min-h-11 cursor-default items-center gap-2.5 rounded-fc-sm px-2.5 text-fc-sm',
								index === activeIndex ? 'bg-fc-accent text-fc-accent-fg' : 'text-fc-fg'
							)}
							onmouseenter={() => (activeIndex = index)}
							onmousedown={(event) => event.preventDefault()}
							onclick={() => choose(option)}
							onkeydown={handleKeydown}
						>
							{#if option.kind === 'task'}
								<iconify-icon icon={toIconify(option.result.project.icon)} width="16" height="16" class="block shrink-0"></iconify-icon>
								<span class="min-w-0 flex-1 truncate">
									<span class="opacity-70">{option.result.project.name} ›</span>
									{option.result.task.name}
								</span>
								{#if group.key === 'recent' && option.result.space}
									<span class="hidden shrink-0 truncate text-fc-xs opacity-70 sm:block">{option.result.space.name}</span>
								{/if}
								{#if option.result.opus}
									<Badge tone="info" title={t('palette.opusTitle')}>Opus</Badge>
								{/if}
							{:else}
								<iconify-icon icon="solar:add-circle-linear" width="16" height="16" class="block shrink-0"></iconify-icon>
								<span class="min-w-0 flex-1 truncate">
									{t('palette.newTask', { name: option.name, project: option.project.name })}
								</span>
							{/if}
						</div>
					{/each}
				</div>
			{:else}
				<p class="px-2.5 py-2 text-fc-sm text-fc-fg-muted">
					{loading ? t('combobox.loading') : t('palette.empty')}
				</p>
			{/each}
		</div>
		{#if error}
			<p class="text-fc-sm text-fc-danger" role="alert">{error}</p>
		{/if}
	</div>
</Modal>
