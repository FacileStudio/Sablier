<script lang="ts">
	import { getContext, onMount, onDestroy } from 'svelte';
	import { backend, type Project, type Task, type TimeEntry } from '$lib/backend';
	import { findTaskByName, upsertTask } from '$lib/task-selection';
	import { notifyTimeEntriesChanged } from '$lib/time-entry-events';
	import { getActiveSpaceId } from '$lib/space-context.svelte';
	import { formatDuration, getTimeEntryDurationMs, isTimeEntryPaused } from '$lib/utils';
	import { Button, Drawer, Field, IconButton, Select, icons } from '@facile/muse';
	import ManualSessionDrawer from '$lib/components/ManualSessionDrawer.svelte';
	import TaskCombobox from '$lib/components/TaskCombobox.svelte';
	import { NotificationService } from '$lib/notifications';
	import { t } from '$lib/i18n/index.svelte';
	import { fly } from 'svelte/transition';

	const timerIcons = {
		play: 'solar:play-linear',
		pause: 'solar:pause-linear',
		stop: 'solar:stop-linear',
		restart: 'solar:restart-linear'
	} as const;

	type Props = {
		projects: Project[];
		onchange?: () => void;
	};

	let { projects, onchange }: Props = $props();

	const ctx = getContext<{ token: string; userEmail: string }>('app');

	let running = $state<TimeEntry | null>(null);
	let elapsed = $state(0);
	let ticker: ReturnType<typeof setInterval> | undefined;

	let drawerOpen = $state(false);
	let selectedProjectId = $state('');
	let tasks = $state<Task[]>([]);
	let taskName = $state('');
	let taskProjectId = $state('');
	let starting = $state(false);
	let taskLoading = $state(false);
	let stopping = $state(false);
	let pausing = $state(false);
	let resuming = $state(false);
	let error = $state('');
	let editDrawerOpen = $state(false);
	let manualDrawerOpen = $state(false);
	let fabOpen = $state(false);

	const runningPaused = $derived(running ? isTimeEntryPaused(running) : false);
	const busy = $derived(stopping || pausing || resuming);

	function projectName(id: number): string {
		return projects.find((p) => p.id === id)?.name ?? String(id);
	}

	function startTicker() {
		stopTicker();
		if (running) {
			elapsed = getTimeEntryDurationMs(running);
			ticker = setInterval(() => {
				elapsed = getTimeEntryDurationMs(running!);
			}, 1000);
		}
	}

	function stopTicker() {
		if (ticker !== undefined) {
			clearInterval(ticker);
			ticker = undefined;
		}
	}

	onMount(async () => {
		const r = await backend.getRunning(ctx.token);
		running = r.entry;
		startTicker();
	});

	onDestroy(() => stopTicker());

	$effect(() => {
		const projectId = selectedProjectId;
		if (projectId === taskProjectId) {
			return;
		}
		taskProjectId = projectId;
		taskName = '';
		if (!projectId) {
			tasks = [];
			error = '';
			return;
		}
		taskLoading = true;
		error = '';
		void backend
			.listTasks(ctx.token, Number(projectId))
			.then((result) => {
				if (selectedProjectId !== projectId) {
					return;
				}
				tasks = result.tasks;
			})
			.catch((e) => {
				if (selectedProjectId !== projectId) {
					return;
				}
				error = e instanceof Error ? e.message : t('timer.failed.tasks');
				tasks = [];
			})
			.finally(() => {
				if (selectedProjectId !== projectId) {
					return;
				}
				taskLoading = false;
			});
	});

	async function resolveTaskId(projectId: number) {
		const trimmedTaskName = taskName.trim();
		if (!trimmedTaskName) {
			throw new Error(t('timer.typeTask'));
		}
		const existingTask = findTaskByName(tasks, trimmedTaskName);
		if (existingTask) {
			taskName = existingTask.name;
			return existingTask.id;
		}
		const task = await backend.createTask(ctx.token, projectId, trimmedTaskName);
		tasks = upsertTask(tasks, task);
		taskName = task.name;
		return task.id;
	}

	async function startTimer() {
		if (!selectedProjectId) {
			error = t('timer.pickProjectFirst');
			return;
		}
		error = '';
		starting = true;
		try {
			const projectId = Number(selectedProjectId);
			const taskId = await resolveTaskId(projectId);
			running = await backend.startTimer(ctx.token, projectId, taskId, getActiveSpaceId());
			selectedProjectId = '';
			taskName = '';
			drawerOpen = false;
			startTicker();
			notifyTimeEntriesChanged();
			NotificationService.triggerTimerStarted(projectName(projectId));
			onchange?.();
		} catch (e) {
			error = e instanceof Error ? e.message : t('timer.failed.start');
		} finally {
			starting = false;
		}
	}

	async function stopTimer() {
		stopping = true;
		try {
			const stoppedProjectId = running!.project_id;
			await backend.stopTimer(ctx.token);
			running = null;
			stopTicker();
			elapsed = 0;
			notifyTimeEntriesChanged();
			NotificationService.triggerTimerStopped(projectName(stoppedProjectId));
			onchange?.();
		} catch (e) {
			error = e instanceof Error ? e.message : t('timer.failed.stop');
		} finally {
			stopping = false;
		}
	}

	async function pauseTimer() {
		pausing = true;
		error = '';
		try {
			running = await backend.pauseTimer(ctx.token);
			startTicker();
			notifyTimeEntriesChanged();
			NotificationService.triggerTimerPaused(projectName(running!.project_id));
			onchange?.();
		} catch (e) {
			error = e instanceof Error ? e.message : t('timer.failed.pause');
		} finally {
			pausing = false;
		}
	}

	async function resumeTimer() {
		resuming = true;
		error = '';
		try {
			running = await backend.resumeTimer(ctx.token);
			startTicker();
			notifyTimeEntriesChanged();
			NotificationService.triggerTimerResumed(projectName(running!.project_id));
			onchange?.();
		} catch (e) {
			error = e instanceof Error ? e.message : t('timer.failed.resume');
		} finally {
			resuming = false;
		}
	}

	async function handleRunningEditChange() {
		const result = await backend.getRunning(ctx.token);
		running = result.entry;
		startTicker();
		notifyTimeEntriesChanged();
		onchange?.();
	}
</script>

<!-- Desktop timer panel -->
<div class="fixed top-3 left-1/2 z-50 hidden -translate-x-1/2 flex-col items-center gap-2 md:flex">
	<div
		class="flex items-center rounded-fc-pill bg-fc-bg/70 p-1.5 shadow-lg backdrop-blur-2xl backdrop-saturate-150"
	>
		{#if running}
			<div class="flex items-center gap-4">
				<span
					class="pl-4 font-fc-mono text-fc-2xl font-semibold tabular-nums leading-none text-fc-fg"
				>
					{formatDuration(elapsed, { includeSeconds: true })}
				</span>
				<div class="flex items-center gap-2">
					<Button
						variant={runningPaused ? 'primary' : 'outline'}
						icon={runningPaused ? timerIcons.restart : timerIcons.pause}
						onclick={runningPaused ? resumeTimer : pauseTimer}
						disabled={busy}
					>
						{#if runningPaused}
							{resuming ? t('timer.resuming') : t('timer.resume')}
						{:else}
							{pausing ? t('timer.pausing') : t('timer.pause')}
						{/if}
					</Button>
					<Button variant="danger" icon={timerIcons.stop} onclick={stopTimer} disabled={busy}>
						{stopping ? t('timer.stopping') : t('timer.stop')}
					</Button>
					<IconButton
						aria-label={t('timer.editEntry')}
						onclick={() => (editDrawerOpen = true)}
						disabled={busy}
					>
						<iconify-icon icon={icons.edit} width="18" height="18" class="block"></iconify-icon>
					</IconButton>
				</div>
			</div>
		{:else}
			<div class="flex items-center gap-2">
				<Button icon={timerIcons.play} onclick={() => (drawerOpen = true)}>{t('timer.start')}</Button>
				<IconButton aria-label={t('timer.addSession')} onclick={() => (manualDrawerOpen = true)}>
					<iconify-icon icon={icons.plus} width="18" height="18" class="block"></iconify-icon>
				</IconButton>
			</div>
		{/if}
	</div>
	{#if error}
		<p
			class="rounded-fc-pill bg-fc-bg/70 px-4 py-1.5 text-fc-sm text-fc-danger shadow-lg backdrop-blur-2xl backdrop-saturate-150"
			role="alert"
		>
			{error}
		</p>
	{/if}
</div>

<!-- Mobile FAB backdrop -->
{#if fabOpen}
	<div
		class="fixed inset-0 z-30 md:hidden"
		onclick={() => (fabOpen = false)}
		role="presentation"
	></div>
{/if}

<!-- Mobile FAB -->
<div class="fixed right-4 bottom-28 z-40 flex flex-col items-end gap-3 md:hidden">
	{#if running}
		{#if fabOpen}
			<div class="flex flex-col items-end gap-2" transition:fly={{ y: 10, duration: 150 }}>
				<Button
					size="lg"
					variant="outline"
					class="bg-fc-bg/70 shadow-lg backdrop-blur-2xl backdrop-saturate-150"
					icon={icons.edit}
					onclick={() => {
						fabOpen = false;
						editDrawerOpen = true;
					}}
					disabled={busy}
				>
					{t('timer.editEntry')}
				</Button>
				<Button
					size="lg"
					class="shadow-lg"
					icon={runningPaused ? timerIcons.restart : timerIcons.pause}
					onclick={() => {
						fabOpen = false;
						void (runningPaused ? resumeTimer() : pauseTimer());
					}}
					disabled={busy}
				>
					{#if runningPaused}
						{resuming ? t('timer.resuming') : t('timer.resume')}
					{:else}
						{pausing ? t('timer.pausing') : t('timer.pause')}
					{/if}
				</Button>
				<Button
					size="lg"
					variant="danger"
					class="bg-fc-bg/70 shadow-lg backdrop-blur-2xl backdrop-saturate-150"
					icon={timerIcons.stop}
					onclick={() => {
						fabOpen = false;
						void stopTimer();
					}}
					disabled={busy}
				>
					{stopping ? t('timer.stopping') : t('timer.stop')}
				</Button>
			</div>
		{/if}
		<button
			type="button"
			class="flex h-14 items-center gap-2.5 rounded-fc-pill bg-fc-accent px-5 text-fc-accent-fg shadow-lg focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
			onclick={() => (fabOpen = !fabOpen)}
			aria-expanded={fabOpen}
			aria-label={t('timer.actions')}
		>
			<span class="size-2 shrink-0 animate-pulse rounded-fc-full bg-fc-success"></span>
			<span class="font-fc-mono text-fc-lg font-semibold tabular-nums">
				{formatDuration(elapsed, { includeSeconds: true })}
			</span>
		</button>
	{:else}
		{#if fabOpen}
			<div class="flex flex-col items-end gap-2" transition:fly={{ y: 10, duration: 150 }}>
				<Button
					size="lg"
					variant="outline"
					class="bg-fc-bg/70 shadow-lg backdrop-blur-2xl backdrop-saturate-150"
					icon={icons.plus}
					onclick={() => {
						fabOpen = false;
						manualDrawerOpen = true;
					}}
				>
					{t('timer.addEntry')}
				</Button>
				<Button
					size="lg"
					class="shadow-lg"
					icon={timerIcons.play}
					onclick={() => {
						fabOpen = false;
						drawerOpen = true;
					}}
				>
					{t('timer.start')}
				</Button>
			</div>
		{/if}
		<button
			type="button"
			class="flex size-14 items-center justify-center rounded-fc-pill bg-fc-accent text-fc-accent-fg shadow-lg focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
			onclick={() => (fabOpen = !fabOpen)}
			aria-expanded={fabOpen}
			aria-label={t('timer.actions')}
		>
			<iconify-icon
				icon={fabOpen ? icons.close : timerIcons.play}
				width="24"
				height="24"
				class="block"
			></iconify-icon>
		</button>
	{/if}
</div>

<!-- Drawers (shared, portal-rendered) -->
<Drawer bind:open={drawerOpen} title={t('timer.drawer.title')}>
	<div class="flex flex-col gap-4">
		<Field label={t('common.project')}>
			<Select bind:value={selectedProjectId}>
				<option value="">{t('timer.selectProject')}</option>
				{#each projects as project (project.id)}
					<option value={String(project.id)}>{project.name}</option>
				{/each}
			</Select>
		</Field>
		<Field label={t('common.task')}>
			<TaskCombobox
				{tasks}
				bind:value={taskName}
				disabled={!selectedProjectId}
				loading={taskLoading}
				placeholder={!selectedProjectId ? t('timer.selectProjectFirst') : t('timer.taskPlaceholder')}
			/>
		</Field>
		{#if error}
			<p class="text-fc-sm text-fc-danger" role="alert">{error}</p>
		{/if}
	</div>

	{#snippet footer()}
		<Button
			size="lg"
			class="w-full"
			icon={timerIcons.play}
			onclick={startTimer}
			disabled={starting}
		>
			{starting ? t('timer.drawer.starting') : t('timer.drawer.start')}
		</Button>
	{/snippet}
</Drawer>

{#if running}
	<ManualSessionDrawer
		{projects}
		editEntry={running}
		bind:open={editDrawerOpen}
		hideTrigger
		onchange={handleRunningEditChange}
	/>
{/if}
<ManualSessionDrawer {projects} bind:open={manualDrawerOpen} hideTrigger />
