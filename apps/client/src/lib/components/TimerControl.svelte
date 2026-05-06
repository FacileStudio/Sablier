<script lang="ts">
	import { getContext, onMount, onDestroy } from 'svelte';
	import { backend, type Project, type Task, type TimeEntry } from '$lib/backend';
	import { findTaskByName, upsertTask } from '$lib/task-selection';
	import { formatDuration } from '$lib/utils';
	import { Button } from '$lib/components/ui/button';
	import { Label } from '$lib/components/ui/label';
	import * as Select from '$lib/components/ui/select';
	import * as Drawer from '$lib/components/ui/drawer';
	import ManualSessionDrawer from '$lib/components/ManualSessionDrawer.svelte';
	import TaskCombobox from '$lib/components/TaskCombobox.svelte';
	import { Play, Square, Pencil } from 'lucide-svelte';

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
	let error = $state('');
	let editDrawerOpen = $state(false);

	function projectName(id: number): string {
		return projects.find((p) => p.id === id)?.name ?? String(id);
	}

	function startTicker() {
		stopTicker();
		if (running) {
			elapsed = Math.max(0, Date.now() - new Date(running.started_at).getTime());
			ticker = setInterval(() => {
				elapsed = Math.max(0, Date.now() - new Date(running!.started_at).getTime());
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
			await backend.stopTimer(ctx.token);
			running = null;
			stopTicker();
			elapsed = 0;
			onchange?.();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to stop timer.';
		} finally {
			stopping = false;
		}
	}

	async function handleRunningEditChange() {
		const result = await backend.getRunning(ctx.token);
		running = result.entry;
		startTicker();
		onchange?.();
	}
</script>

{#if running}
	<div class="flex items-center gap-4">
		<span class="tabular-nums leading-none" style="font-family: var(--font-heading); font-size: clamp(1.75rem, 4vw, 2.5rem); font-weight: 700;">{formatDuration(elapsed, { includeSeconds: true })}</span>
		<div class="flex items-center gap-2">
			<Button
				class="gap-2 h-10 px-5 bg-red-600 hover:bg-red-700 text-white border-0"
				onclick={stopTimer}
				disabled={stopping}
			>
				<Square class="h-4 w-4" />
				{stopping ? 'Stopping…' : 'Stop'}
			</Button>
			<Button
				variant="outline"
				size="icon"
				class="h-10 w-10"
				onclick={() => (editDrawerOpen = true)}
				disabled={stopping}
			>
				<Pencil class="h-4 w-4" />
			</Button>
		</div>
		<ManualSessionDrawer
			{projects}
			editEntry={running}
			bind:open={editDrawerOpen}
			hideTrigger
			onchange={handleRunningEditChange}
		/>
	</div>
{:else}
	<Drawer.Root bind:open={drawerOpen} direction="bottom">
		<Drawer.Trigger>
			<Button class="gap-2 h-10 px-5" onclick={() => (drawerOpen = true)}>
				<Play class="h-4 w-4" />
				Start
			</Button>
		</Drawer.Trigger>
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
{/if}
