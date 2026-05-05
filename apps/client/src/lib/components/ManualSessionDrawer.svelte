<script lang="ts">
	import { getContext } from 'svelte';
	import { backend, type Project, type Task } from '$lib/backend';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Select from '$lib/components/ui/select';
	import * as Drawer from '$lib/components/ui/drawer';
	import { Plus } from 'lucide-svelte';

	type Props = {
		projects: Project[];
		onchange?: () => void;
	};

	let { projects, onchange }: Props = $props();

	const ctx = getContext<{ token: string; userEmail: string }>('app');

	let drawerOpen = $state(false);
	let selectedProjectId = $state('');
	let tasks = $state<Task[]>([]);
	let selectedTaskId = $state('');
	let newTaskName = $state('');
	let taskProjectId = $state('');
	let startedAt = $state('');
	let stoppedAt = $state('');
	let saving = $state(false);
	let taskLoading = $state(false);
	let error = $state('');

	function projectName(id: number): string {
		return projects.find((p) => p.id === id)?.name ?? String(id);
	}

	function toIso(localDatetime: string): string {
		return new Date(localDatetime).toISOString();
	}

	function reset() {
		selectedProjectId = '';
		tasks = [];
		selectedTaskId = '';
		newTaskName = '';
		taskProjectId = '';
		startedAt = '';
		stoppedAt = '';
		error = '';
	}

	$effect(() => {
		const projectId = selectedProjectId;
		if (projectId === taskProjectId) {
			return;
		}
		taskProjectId = projectId;
		selectedTaskId = '';
		newTaskName = '';
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
		const taskName = newTaskName.trim();
		if (taskName !== '') {
			const task = await backend.createTask(ctx.token, projectId, taskName);
			if (!tasks.some((existing) => existing.id === task.id)) {
				tasks = [...tasks, task].sort((a, b) => a.name.localeCompare(b.name));
			}
			selectedTaskId = String(task.id);
			newTaskName = '';
			return task.id;
		}
		if (!selectedTaskId) {
			throw new Error('Pick a task or create one.');
		}
		return Number(selectedTaskId);
	}

	async function handleSave() {
		if (!selectedProjectId) {
			error = 'Pick a project.';
			return;
		}
		if (!startedAt) {
			error = 'Start time is required.';
			return;
		}
		if (!stoppedAt) {
			error = 'End time is required.';
			return;
		}
		const startIso = toIso(startedAt);
		const stopIso = toIso(stoppedAt);
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
							<Label for="manual-started-at">Start</Label>
							<Input
								id="manual-started-at"
								type="datetime-local"
								bind:value={startedAt}
							/>
						</div>
						<div class="flex flex-col gap-1.5">
							<Label for="manual-stopped-at">End</Label>
							<Input
								id="manual-stopped-at"
								type="datetime-local"
								bind:value={stoppedAt}
							/>
						</div>
					</div>
					<div class="flex flex-col gap-1.5">
						<Label for="manual-task-select">Task</Label>
						<Select.Root type="single" bind:value={selectedTaskId}>
							<Select.Trigger id="manual-task-select" class="w-full">
								{#if selectedTaskId}
									{tasks.find((task) => task.id === Number(selectedTaskId))?.name ?? 'Select a task'}
								{:else if taskLoading}
									Loading tasks…
								{:else if !selectedProjectId}
									Select a project first
								{:else}
									Select a task
								{/if}
							</Select.Trigger>
							<Select.Content>
								{#each tasks as task}
									<Select.Item value={String(task.id)}>{task.name}</Select.Item>
								{/each}
							</Select.Content>
						</Select.Root>
					</div>
					<div class="flex flex-col gap-1.5">
						<Label for="manual-task-create-input">New task</Label>
						<Input
							id="manual-task-create-input"
							placeholder="Create a task for this project"
							bind:value={newTaskName}
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
