<script lang="ts">
	import { getContext } from 'svelte';
	import { CalendarDate, today, getLocalTimeZone } from '@internationalized/date';
	import type { DateValue } from '@internationalized/date';
	import { backend, type Project, type Task } from '$lib/backend';
	import { findTaskByName, upsertTask } from '$lib/task-selection';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Select from '$lib/components/ui/select';
	import * as Drawer from '$lib/components/ui/drawer';
	import * as Popover from '$lib/components/ui/popover';
	import * as Calendar from '$lib/components/ui/calendar';
	import TaskCombobox from '$lib/components/TaskCombobox.svelte';
	import { CalendarIcon, Plus } from 'lucide-svelte';
	import { cn } from '$lib/utils';

	type Props = {
		projects: Project[];
		onchange?: () => void;
	};

	let { projects, onchange }: Props = $props();

	const ctx = getContext<{ token: string; userEmail: string }>('app');

	let drawerOpen = $state(false);
	let selectedProjectId = $state('');
	let tasks = $state<Task[]>([]);
	let taskName = $state('');
	let taskProjectId = $state('');
	let startDate = $state<DateValue | undefined>(undefined);
	let startTime = $state('');
	let endDate = $state<DateValue | undefined>(undefined);
	let endTime = $state('');
	let startPopoverOpen = $state(false);
	let endPopoverOpen = $state(false);
	let saving = $state(false);
	let taskLoading = $state(false);
	let error = $state('');

	function projectName(id: number): string {
		return projects.find((p) => p.id === id)?.name ?? String(id);
	}

	function formatDate(date: DateValue | undefined): string {
		if (!date) return 'Pick a date';
		return new Date(date.year, date.month - 1, date.day).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	function buildIso(date: DateValue | undefined, time: string): string | null {
		if (!date || !time) return null;
		const [hours, minutes] = time.split(':').map(Number);
		const d = new Date(date.year, date.month - 1, date.day, hours, minutes);
		return d.toISOString();
	}

	function reset() {
		selectedProjectId = '';
		tasks = [];
		taskName = '';
		taskProjectId = '';
		startDate = undefined;
		startTime = '';
		endDate = undefined;
		endTime = '';
		error = '';
	}

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
		const stopIso = buildIso(endDate, endTime);
		if (!startIso) {
			error = 'Start date and time are required.';
			return;
		}
		if (!stopIso) {
			error = 'End date and time are required.';
			return;
		}
		if (new Date(stopIso) <= new Date(startIso)) {
			error = 'End time must be after start time.';
			return;
		}
		error = '';
		saving = true;
		try {
			const projectId = Number(selectedProjectId);
			const taskId = await resolveTaskId(projectId);
			await backend.createEntry(ctx.token, projectId, taskId, startIso, stopIso);
			reset();
			drawerOpen = false;
			onchange?.();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to save session.';
		} finally {
			saving = false;
		}
	}
</script>

<Drawer.Root bind:open={drawerOpen} direction="bottom">
	<Drawer.Trigger>
		<Button variant="outline" class="gap-2 h-10 px-5" onclick={() => (drawerOpen = true)}>
			<Plus class="h-4 w-4" />
			Add session
		</Button>
	</Drawer.Trigger>
	<Drawer.Portal>
		<Drawer.Overlay class="fixed inset-0 bg-black/40" />
		<Drawer.Content class="fixed bottom-0 left-0 right-0 flex flex-col rounded-t-2xl bg-background border-t">
			<div class="mx-auto w-12 h-1.5 rounded-full bg-muted mt-4 mb-6 shrink-0"></div>
			<div class="px-6 pb-8 flex flex-col gap-6 max-w-lg mx-auto w-full">
				<Drawer.Header class="p-0">
					<Drawer.Title>Add a session manually</Drawer.Title>
				</Drawer.Header>
				<div class="flex flex-col gap-4">
					<div class="flex flex-col gap-1.5">
						<Label for="manual-project-select">Project</Label>
						<Select.Root type="single" bind:value={selectedProjectId}>
							<Select.Trigger id="manual-project-select" class="w-full">
								{selectedProjectId ? projectName(Number(selectedProjectId)) : 'Select a project'}
							</Select.Trigger>
							<Select.Content>
								{#each projects as project}
									<Select.Item value={String(project.id)}>{project.name}</Select.Item>
								{/each}
							</Select.Content>
						</Select.Root>
					</div>
					<div class="grid grid-cols-2 gap-3">
						<div class="flex flex-col gap-1.5">
							<Label>Start</Label>
							<Popover.Root bind:open={startPopoverOpen}>
								<Popover.Trigger>
									<Button
										variant="outline"
										class={cn('w-full justify-start text-left font-normal gap-2', !startDate && 'text-muted-foreground')}
									>
										<CalendarIcon class="h-4 w-4 shrink-0" />
										{formatDate(startDate)}
									</Button>
								</Popover.Trigger>
								<Popover.Content class="w-auto p-0" align="start">
									<Calendar.Calendar
										type="single"
										bind:value={startDate}
										onValueChange={() => (startPopoverOpen = false)}
									/>
								</Popover.Content>
							</Popover.Root>
							<Input type="time" bind:value={startTime} class="w-full" />
						</div>
						<div class="flex flex-col gap-1.5">
							<Label>End</Label>
							<Popover.Root bind:open={endPopoverOpen}>
								<Popover.Trigger>
									<Button
										variant="outline"
										class={cn('w-full justify-start text-left font-normal gap-2', !endDate && 'text-muted-foreground')}
									>
										<CalendarIcon class="h-4 w-4 shrink-0" />
										{formatDate(endDate)}
									</Button>
								</Popover.Trigger>
								<Popover.Content class="w-auto p-0" align="start">
									<Calendar.Calendar
										type="single"
										bind:value={endDate}
										onValueChange={() => (endPopoverOpen = false)}
									/>
								</Popover.Content>
							</Popover.Root>
							<Input type="time" bind:value={endTime} class="w-full" />
						</div>
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
					<Button class="gap-2 w-full h-12 text-base" onclick={handleSave} disabled={saving}>
						<Plus class="h-4 w-4" />
						{saving ? 'Saving…' : 'Add session'}
					</Button>
				</div>
			</div>
		</Drawer.Content>
	</Drawer.Portal>
</Drawer.Root>
