<script lang="ts">
	import { getContext } from 'svelte';
	import { backend, type Project, type Task, type TimeEntry } from '$lib/backend';
	import { findTaskByName, upsertTask } from '$lib/task-selection';
	import { notifyTimeEntriesChanged } from '$lib/time-entry-events';
	import { getActiveSpaceId } from '$lib/space-context.svelte';
	import { Button, Drawer, Field, Input, Select, icons } from '@facile/muse';
	import TaskCombobox from '$lib/components/TaskCombobox.svelte';

	type Props = {
		projects: Project[];
		editEntry?: TimeEntry | null;
		open?: boolean;
		hideTrigger?: boolean;
		onchange?: () => void;
		onclose?: () => void;
	};

	let { projects, editEntry = null, open = $bindable(false), hideTrigger = false, onchange, onclose }: Props = $props();

	const ctx = getContext<{ token: string; userEmail: string }>('app');
	const uid = $props.id();

	let drawerOpen = $state(false);
	let selectedProjectId = $state('');
	let tasks = $state<Task[]>([]);
	let taskName = $state('');
	let taskProjectId = $state('');
	let startDate = $state('');
	let startTime = $state('');
	let endDate = $state('');
	let endTime = $state('');
	let saving = $state(false);
	let taskLoading = $state(false);
	let error = $state('');

	const isEditMode = $derived(editEntry != null);
	const isRunningEdit = $derived(editEntry != null && editEntry.stopped_at == null);

	function toDateInput(d: Date): string {
		return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
	}

	function isoToDateValue(iso: string): string {
		return toDateInput(new Date(iso));
	}

	function isoToTime(iso: string): string {
		const d = new Date(iso);
		return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`;
	}

	function buildIso(date: string, time: string): string | null {
		if (!date || !time) return null;
		const [year, month, day] = date.split('-').map(Number);
		const [hours, minutes] = time.split(':').map(Number);
		if ([year, month, day, hours, minutes].some((n) => Number.isNaN(n))) return null;
		return new Date(year, month - 1, day, hours, minutes).toISOString();
	}

	function reset() {
		selectedProjectId = '';
		tasks = [];
		taskName = '';
		taskProjectId = '';
		startDate = '';
		startTime = '';
		endDate = '';
		endTime = '';
		error = '';
	}

	function populateFromEntry(entry: TimeEntry) {
		selectedProjectId = String(entry.project_id);
		taskName = entry.task_name ?? '';
		startDate = isoToDateValue(entry.started_at);
		startTime = isoToTime(entry.started_at);
		if (entry.stopped_at) {
			endDate = isoToDateValue(entry.stopped_at);
			endTime = isoToTime(entry.stopped_at);
		} else {
			endDate = '';
			endTime = '';
		}
	}

	function nowTime(): string {
		const n = new Date();
		return `${String(n.getHours()).padStart(2, '0')}:${String(n.getMinutes()).padStart(2, '0')}`;
	}

	function todayDate(): string {
		return toDateInput(new Date());
	}

	$effect(() => {
		if (open) {
			if (editEntry) {
				populateFromEntry(editEntry);
			} else {
				startDate = todayDate();
				startTime = nowTime();
				endDate = todayDate();
				endTime = nowTime();
				if (projects.length === 1) {
					selectedProjectId = String(projects[0].id);
				}
			}
			drawerOpen = true;
		}
	});

	$effect(() => {
		if (!drawerOpen) {
			open = false;
			onclose?.();
		}
	});

	$effect(() => {
		const projectId = selectedProjectId;
		if (projectId === taskProjectId) {
			return;
		}
		taskProjectId = projectId;
		taskName = '';
		tasks = [];
		if (!projectId) {
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
				if (editEntry && String(editEntry.project_id) === projectId) {
					taskName = editEntry.task_name ?? '';
				}
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

	async function handleSave() {
		if (!selectedProjectId) {
			error = 'Pick a project.';
			return;
		}
		const startIso = buildIso(startDate, startTime);
		const stopIso = isRunningEdit ? null : buildIso(endDate, endTime);
		if (!startIso) {
			error = 'Start date and time are required.';
			return;
		}
		if (!isRunningEdit && !stopIso) {
			error = 'End date and time are required.';
			return;
		}
		if (stopIso && new Date(stopIso) <= new Date(startIso)) {
			error = 'End time must be after start time.';
			return;
		}
		error = '';
		saving = true;
		try {
			const projectId = Number(selectedProjectId);
			const taskId = await resolveTaskId(projectId);
			if (isEditMode && editEntry) {
				await backend.updateEntry(ctx.token, editEntry.id, projectId, taskId, startIso, stopIso);
			} else {
				if (!stopIso) {
					throw new Error('End date and time are required.');
				}
				await backend.createEntry(ctx.token, projectId, taskId, startIso, stopIso, getActiveSpaceId());
			}
			reset();
			drawerOpen = false;
			notifyTimeEntriesChanged();
			onchange?.();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to save session.';
		} finally {
			saving = false;
		}
	}
</script>

{#if !hideTrigger}
	<Button variant="outline" icon={icons.plus} onclick={() => (drawerOpen = true)}>Add session</Button>
{/if}

<Drawer bind:open={drawerOpen} title={isEditMode ? 'Edit session' : 'Add a session'}>
	<div class="flex flex-col gap-4">
		<Field label="Project">
			<Select bind:value={selectedProjectId}>
				<option value="">Select a project</option>
				{#each projects as project (project.id)}
					<option value={String(project.id)}>{project.name}</option>
				{/each}
			</Select>
		</Field>

		<div class={isRunningEdit ? 'grid grid-cols-1 gap-3' : 'grid grid-cols-2 gap-3'}>
			<Field label="Start">
				<Input type="date" bind:value={startDate} />
				<Input id="{uid}-start-time" type="time" bind:value={startTime} aria-label="Start time" />
			</Field>
			{#if !isRunningEdit}
				<Field label="End">
					<Input type="date" bind:value={endDate} />
					<Input id="{uid}-end-time" type="time" bind:value={endTime} aria-label="End time" />
				</Field>
			{/if}
		</div>

		<Field label="Task">
			<TaskCombobox
				{tasks}
				bind:value={taskName}
				disabled={!selectedProjectId}
				loading={taskLoading}
				placeholder={!selectedProjectId ? 'Select a project first' : 'Choose or create a task'}
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
			icon={isEditMode ? icons.edit : icons.plus}
			onclick={handleSave}
			disabled={saving}
		>
			{#if isEditMode}
				{saving ? 'Saving…' : isRunningEdit ? 'Update session' : 'Save changes'}
			{:else}
				{saving ? 'Saving…' : 'Add session'}
			{/if}
		</Button>
	{/snippet}
</Drawer>
