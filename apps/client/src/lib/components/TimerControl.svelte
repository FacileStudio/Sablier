<script lang="ts">
	import { getContext, onMount, onDestroy } from 'svelte';
	import { backend, type Project, type Task, type TimeEntry } from '$lib/backend';
	import { findTaskByName, upsertTask } from '$lib/task-selection';
	import { notifyTimeEntriesChanged } from '$lib/time-entry-events';
	import { formatDuration, getTimeEntryDurationMs, isTimeEntryPaused } from '$lib/utils';
	import { Button } from '$lib/components/ui/button';
	import { Label } from '$lib/components/ui/label';
	import * as Select from '$lib/components/ui/select';
	import * as Drawer from '$lib/components/ui/drawer';
	import ManualSessionDrawer from '$lib/components/ManualSessionDrawer.svelte';
	import TaskCombobox from '$lib/components/TaskCombobox.svelte';
	import { Play, Square, Pencil, Pause, TimerReset, Plus, X } from 'lucide-svelte';
	import { NotificationService } from '$lib/notifications';
	import { fly } from 'svelte/transition';

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
				error = e instanceof Error ? e.message : 'Failed to load tasks.';
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
			throw new Error('Type a task name.');
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
			error = 'Pick a project first.';
			return;
		}
		error = '';
		starting = true;
		try {
			const projectId = Number(selectedProjectId);
			const taskId = await resolveTaskId(projectId);
			running = await backend.startTimer(ctx.token, projectId, taskId);
			selectedProjectId = '';
			taskName = '';
			drawerOpen = false;
			startTicker();
			notifyTimeEntriesChanged();
			NotificationService.triggerTimerStarted(projectName(projectId));
			onchange?.();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to start timer.';
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
			error = e instanceof Error ? e.message : 'Failed to stop timer.';
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
			error = e instanceof Error ? e.message : 'Failed to pause timer.';
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
			error = e instanceof Error ? e.message : 'Failed to resume timer.';
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
<div class="fixed top-0 left-1/2 z-50 -translate-x-1/2 hidden md:block">
	<div class="rounded-b-2xl border border-t-0 bg-background px-5 py-3 shadow-lg shadow-black/10">
		{#if running}
			<div class="flex items-center gap-4">
				<span class="leading-none" style="font-family: var(--font-mono); font-size: clamp(1.75rem, 4vw, 2.5rem); font-weight: 700;">{formatDuration(elapsed, { includeSeconds: true })}</span>
				<div class="flex items-center gap-2">
					<Button
						variant={runningPaused ? 'default' : 'outline'}
						class="gap-2 h-10 px-5"
						onclick={runningPaused ? resumeTimer : pauseTimer}
						disabled={stopping || pausing || resuming}
					>
						{#if runningPaused}
							<TimerReset class="h-4 w-4" />
							{resuming ? 'Resuming…' : 'Resume'}
						{:else}
							<Pause class="h-4 w-4" />
							{pausing ? 'Pausing…' : 'Pause'}
						{/if}
					</Button>
					<Button
						class="gap-2 h-10 px-5 bg-red-600 hover:bg-red-700 text-white border-0"
						onclick={stopTimer}
						disabled={stopping || pausing || resuming}
					>
						<Square class="h-4 w-4" />
						{stopping ? 'Stopping…' : 'Stop'}
					</Button>
					<Button
						variant="outline"
						size="icon"
						class="h-10 w-10"
						onclick={() => (editDrawerOpen = true)}
						disabled={stopping || pausing || resuming}
					>
						<Pencil class="h-4 w-4" />
					</Button>
				</div>
			</div>
		{:else}
			<div class="flex items-center gap-2">
				<Button class="gap-2 h-10 px-5" onclick={() => (drawerOpen = true)}>
					<Play class="h-4 w-4" />
					Start Session
				</Button>
				<Button variant="outline" size="icon" class="h-10 w-10" onclick={() => (manualDrawerOpen = true)}>
					<Plus class="h-4 w-4" />
				</Button>
			</div>
		{/if}
		{#if error}
			<p class="text-sm text-destructive">{error}</p>
		{/if}
	</div>
</div>

<!-- Mobile FAB backdrop -->
{#if fabOpen}
	<div class="fixed inset-0 z-30 md:hidden" onclick={() => (fabOpen = false)} role="presentation"></div>
{/if}

<!-- Mobile FAB -->
<div class="fixed bottom-6 right-6 z-40 md:hidden flex flex-col items-end gap-3">
	{#if running}
		{#if fabOpen}
			<div class="flex flex-col items-end gap-2" transition:fly={{ y: 10, duration: 150 }}>
				<button
					class="flex items-center gap-2 rounded-2xl bg-foreground text-background px-4 h-12 text-sm font-medium shadow-lg whitespace-nowrap disabled:opacity-50"
					onclick={() => { fabOpen = false; editDrawerOpen = true; }}
					disabled={stopping || pausing || resuming}
				>
					<Pencil class="h-4 w-4" />
					Edit entry
				</button>
				<button
					class="flex items-center gap-2 rounded-2xl bg-foreground text-background px-4 h-12 text-sm font-medium shadow-lg whitespace-nowrap disabled:opacity-50"
					onclick={() => { fabOpen = false; void (runningPaused ? resumeTimer() : pauseTimer()); }}
					disabled={stopping || pausing || resuming}
				>
					{#if runningPaused}
						<TimerReset class="h-4 w-4" />
						{resuming ? 'Resuming…' : 'Resume'}
					{:else}
						<Pause class="h-4 w-4" />
						{pausing ? 'Pausing…' : 'Pause'}
					{/if}
				</button>
				<button
					class="flex items-center gap-2 rounded-2xl bg-red-600 text-white px-4 h-12 text-sm font-medium shadow-lg whitespace-nowrap disabled:opacity-50"
					onclick={() => { fabOpen = false; void stopTimer(); }}
					disabled={stopping || pausing || resuming}
				>
					<Square class="h-4 w-4" />
					{stopping ? 'Stopping…' : 'Stop'}
				</button>
			</div>
		{/if}
		<button
			class="flex items-center gap-2.5 rounded-2xl bg-foreground text-background px-5 h-14 font-medium shadow-xl"
			onclick={() => (fabOpen = !fabOpen)}
		>
			<span class="h-2 w-2 rounded-full bg-green-400 animate-pulse shrink-0"></span>
			<span style="font-family: var(--font-mono); font-size: 1.125rem; font-weight: 700;">
				{formatDuration(elapsed, { includeSeconds: true })}
			</span>
		</button>
	{:else}
		{#if fabOpen}
			<div class="flex flex-col items-end gap-2" transition:fly={{ y: 10, duration: 150 }}>
				<Button
					variant="outline"
					class="px-4 h-12 text-sm shadow-lg whitespace-nowrap"
					onclick={() => { fabOpen = false; manualDrawerOpen = true; }}
				>
					<Plus class="h-4 w-4" />
					Add entry
				</Button>
				<Button
				variant="default"
					class="px-4 h-12 text-sm shadow-lg whitespace-nowrap"
					onclick={() => { fabOpen = false; drawerOpen = true; }}
				>
					<Play class="h-4 w-4" />
					Start session
				</Button>
			</div>
		{/if}
		<button
			class="flex h-14 w-14 items-center justify-center rounded-2xl bg-foreground text-background shadow-xl"
			onclick={() => (fabOpen = !fabOpen)}
			aria-label="Timer actions"
		>
			{#if fabOpen}
				<X class="h-6 w-6" />
			{:else}
				<Play class="h-6 w-6" />
			{/if}
		</button>
	{/if}
</div>

<!-- Drawers (shared, portal-rendered) -->
<Drawer.Root bind:open={drawerOpen} direction="bottom">
	<Drawer.Portal>
		<Drawer.Overlay class="fixed inset-0 bg-black/40" />
		<Drawer.Content class="fixed bottom-0 left-0 right-0 flex flex-col rounded-t-2xl bg-background border-t">
			<div class="mx-auto w-12 h-1.5 rounded-full bg-muted mt-4 mb-6 shrink-0"></div>
			<div class="px-6 pb-8 flex flex-col gap-6 max-w-lg mx-auto w-full">
				<Drawer.Header class="p-0">
					<Drawer.Title>Start a timer</Drawer.Title>
				</Drawer.Header>
				<div class="flex flex-col gap-4">
					<div class="flex flex-col gap-1.5">
						<Label for="timer-project-select">Project</Label>
						<Select.Root type="single" bind:value={selectedProjectId}>
							<Select.Trigger id="timer-project-select" class="w-full">
								{selectedProjectId ? projectName(Number(selectedProjectId)) : 'Select a project'}
							</Select.Trigger>
							<Select.Content>
								{#each projects as project}
									<Select.Item value={String(project.id)}>{project.name}</Select.Item>
								{/each}
							</Select.Content>
						</Select.Root>
					</div>
					<div class="flex flex-col gap-1.5">
						<Label>Task</Label>
						<TaskCombobox
							{tasks}
							bind:value={taskName}
							disabled={!selectedProjectId}
							loading={taskLoading}
							placeholder={!selectedProjectId ? 'Select a project first' : 'Choose or create a task'}
						/>
					</div>
					{#if error}
						<p class="text-sm text-destructive">{error}</p>
					{/if}
					<Button class="gap-2 w-full h-12 text-base" onclick={startTimer} disabled={starting}>
						<Play class="h-4 w-4" />
						{starting ? 'Starting…' : 'Start timer'}
					</Button>
				</div>
			</div>
		</Drawer.Content>
	</Drawer.Portal>
</Drawer.Root>

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
