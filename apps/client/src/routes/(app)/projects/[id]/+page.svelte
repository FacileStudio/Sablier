<script lang="ts">
	import { getContext, onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { backend, type Project, type Task, type TimeEntry, type UserProfile } from '$lib/backend';
	import { getEntryUserDisplayName } from '$lib/user-display';
	import UserAvatarBadge from '$lib/components/UserAvatarBadge.svelte';
	import UserColorSplitBar from '$lib/components/UserColorSplitBar.svelte';
	import ManualSessionDrawer from '$lib/components/ManualSessionDrawer.svelte';
	import * as AlertDialog from '$lib/components/ui/alert-dialog';
	import * as Drawer from '$lib/components/ui/drawer';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Clock, BarChart3, Calendar, ArrowLeft, Timer, Pencil, Trash2, Check, X, Save } from 'lucide-svelte';

	const ctx = getContext<{ token: string; userEmail: string; user: UserProfile | null }>('app');

	let loading = $state(true);
	let error = $state('');
	let project = $state<Project | null>(null);
	let tasks = $state<Task[]>([]);
	let entries = $state<TimeEntry[]>([]);
	let projectEditDrawerOpen = $state(false);
	let editName = $state('');
	let editDescription = $state('');
	let projectActionError = $state('');
	let savingProject = $state(false);
	let deletingProject = $state(false);
	let projectDeleteDialogOpen = $state(false);
	let deletingEntryId = $state<number | null>(null);
	let deleteError = $state('');
	let entryDeleteDialogOpen = $state(false);
	let deleteEntryTarget = $state<TimeEntry | null>(null);
	let deletingTaskId = $state<number | null>(null);
	let taskDeleteError = $state('');
	let taskDeleteDialogOpen = $state(false);
	let deleteTaskTarget = $state<{ id: number; name: string; sessionCount: number } | null>(null);
	let editingTaskId = $state<number | null>(null);
	let taskDraftName = $state('');
	let taskSaveError = $state('');
	let savingTaskId = $state<number | null>(null);
	let editingEntry = $state<TimeEntry | null>(null);
	let editDrawerOpen = $state(false);
	let now = $state(Date.now());
	let ticker: ReturnType<typeof setInterval> | undefined;

	type UserTimeSegment = {
		key: string;
		label: string;
		color?: string;
		ms: number;
	};

	function formatDuration(ms: number): string {
		const totalSeconds = Math.floor(ms / 1000);
		const h = Math.floor(totalSeconds / 3600);
		const m = Math.floor((totalSeconds % 3600) / 60);
		const s = totalSeconds % 60;
		return [h, m, s].map((v) => String(v).padStart(2, '0')).join(':');
	}

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleString(undefined, {
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function formatDateShort(iso: string): string {
		return new Date(iso).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	function entryMs(e: TimeEntry): number {
		const start = new Date(e.started_at).getTime();
		const end = e.stopped_at ? new Date(e.stopped_at).getTime() : now;
		return end - start;
	}

	function userColor(entry: TimeEntry) {
		return (entry as TimeEntry & { user_color?: string }).user_color;
	}

	function userLabel(entry: TimeEntry) {
		return getEntryUserDisplayName(entry);
	}

	function aggregateUserTimeSegments(entryList: TimeEntry[]): UserTimeSegment[] {
		const segments = new Map<string, UserTimeSegment>();

		for (const entry of entryList) {
			const key = String(entry.user_id ?? entry.user_email ?? entry.id);
			const existing = segments.get(key);
			const ms = entryMs(entry);

			if (existing) {
				existing.ms += ms;
				if (!existing.color) {
					existing.color = userColor(entry);
				}
				if (existing.label.startsWith('User ') && (entry.user_name || entry.user_email)) {
					existing.label = userLabel(entry);
				}
				continue;
			}

			segments.set(key, {
				key,
				label: userLabel(entry),
				color: userColor(entry),
				ms
			});
		}

		return [...segments.values()].sort((a, b) => b.ms - a.ms || a.label.localeCompare(b.label));
	}

	const sortedEntries = $derived(
		[...entries].sort((a, b) => new Date(b.started_at).getTime() - new Date(a.started_at).getTime())
	);

	const projectUserSegments = $derived(aggregateUserTimeSegments(entries));

	const totalMs = $derived(entries.reduce((acc, e) => acc + entryMs(e), 0));

	const avgMs = $derived(entries.length > 0 ? totalMs / entries.length : 0);

	const lastSession = $derived(
		entries.length > 0
			? entries.reduce((latest, e) =>
					new Date(e.started_at) > new Date(latest.started_at) ? e : latest
				)
			: null
	);

	const tasksWithStats = $derived(
		[...tasks]
			.sort((a, b) => a.name.localeCompare(b.name))
			.map((task) => {
				const taskEntries = entries.filter((entry) => entry.task_id === task.id);
				const taskTotalMs = taskEntries.reduce((acc, entry) => acc + entryMs(entry), 0);
				return {
					...task,
					sessionCount: taskEntries.length,
					totalMs: taskTotalMs,
					userSegments: aggregateUserTimeSegments(taskEntries),
					lastStartedAt:
						taskEntries.length > 0
							? taskEntries.reduce((latest, entry) =>
									new Date(entry.started_at) > new Date(latest.started_at) ? entry : latest
								).started_at
							: null
				};
			})
	);

	function openEntryDeleteDialog(entry: TimeEntry) {
		deleteEntryTarget = entry;
		entryDeleteDialogOpen = true;
	}

	async function confirmDeleteEntry() {
		const target = deleteEntryTarget;
		if (!target) {
			return;
		}
		deletingEntryId = target.id;
		deleteError = '';
		try {
			await backend.deleteEntry(ctx.token, target.id);
			entries = entries.filter((entry) => entry.id !== target.id);
			entryDeleteDialogOpen = false;
			deleteEntryTarget = null;
		} catch (e) {
			deleteError = e instanceof Error ? e.message : 'Failed to remove session.';
		} finally {
			deletingEntryId = null;
		}
	}

	function openTaskDeleteDialog(taskId: number, name: string, sessionCount: number) {
		deleteTaskTarget = { id: taskId, name, sessionCount };
		taskSaveError = '';
		taskDeleteDialogOpen = true;
	}

	function openEditDrawer(entry: TimeEntry) {
		editingEntry = entry;
		editDrawerOpen = true;
	}

	function startTaskEdit(taskId: number, name: string) {
		editingTaskId = taskId;
		taskDraftName = name;
		taskSaveError = '';
	}

	function cancelTaskEdit() {
		editingTaskId = null;
		taskDraftName = '';
		taskSaveError = '';
	}

	async function confirmDeleteTask() {
		const target = deleteTaskTarget;
		if (!target || !project) {
			return;
		}
		deletingTaskId = target.id;
		taskDeleteError = '';
		try {
			await backend.deleteTask(ctx.token, project.id, target.id);
			tasks = tasks.filter((t) => t.id !== target.id);
			entries = entries.map((e) => e.task_id === target.id ? { ...e, task_id: 0, task_name: '' } : e);
			taskDeleteDialogOpen = false;
			deleteTaskTarget = null;
		} catch (e) {
			taskDeleteError = e instanceof Error ? e.message : 'Failed to delete task.';
		} finally {
			deletingTaskId = null;
		}
	}

	async function saveTaskName(taskId: number) {
		if (!project) {
			return;
		}
		savingTaskId = taskId;
		taskSaveError = '';
		try {
			const updated = await backend.updateTask(ctx.token, project.id, taskId, taskDraftName);
			tasks = tasks
				.map((task) => task.id === taskId ? updated : task)
				.sort((a, b) => a.name.localeCompare(b.name));
			entries = entries.map((entry) =>
				entry.task_id === taskId ? { ...entry, task_name: updated.name } : entry
			);
			cancelTaskEdit();
		} catch (e) {
			taskSaveError = e instanceof Error ? e.message : 'Failed to rename task.';
		} finally {
			savingTaskId = null;
		}
	}

	function startProjectEdit() {
		if (!project) {
			return;
		}
		projectActionError = '';
		editName = project.name;
		editDescription = project.description;
		projectEditDrawerOpen = true;
	}

	function cancelProjectEdit() {
		projectEditDrawerOpen = false;
		projectActionError = '';
		editName = '';
		editDescription = '';
	}

	async function saveProject() {
		if (!project) {
			return;
		}
		savingProject = true;
		projectActionError = '';
		try {
			project = await backend.updateProject(ctx.token, project.id, editName, editDescription);
			projectEditDrawerOpen = false;
			editName = '';
			editDescription = '';
		} catch (e) {
			projectActionError = e instanceof Error ? e.message : 'Failed to save project.';
		} finally {
			savingProject = false;
		}
	}

	async function deleteProject() {
		if (!project) {
			return;
		}
		deletingProject = true;
		projectActionError = '';
		try {
			await backend.deleteProject(ctx.token, project.id);
			projectDeleteDialogOpen = false;
			await goto('/projects');
		} catch (e) {
			projectActionError = e instanceof Error ? e.message : 'Failed to delete project.';
			deletingProject = false;
		}
	}

	async function handleEntryChange() {
		if (!project) {
			return;
		}
		const result = await backend.listEntries(ctx.token, project.id);
		entries = result.entries;
	}

	onMount(async () => {
		ticker = setInterval(() => { now = Date.now(); }, 1000);
		try {
			const id = Number(page.params.id);
			const [proj, taskResult, ents] = await Promise.all([
				backend.getProject(ctx.token, id),
				backend.listTasks(ctx.token, id),
				backend.listEntries(ctx.token, id)
			]);
			project = proj;
			tasks = taskResult.tasks;
			entries = ents.entries;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load project.';
		} finally {
			loading = false;
		}
	});

	onDestroy(() => clearInterval(ticker));
</script>

<svelte:head>
	<title>{project?.name ?? 'Project'} — Sablier</title>
</svelte:head>

<div class="flex flex-col gap-6 p-6">
	<div>
		<Button variant="ghost" href="/projects" class="mb-4 gap-2 pl-0 text-muted-foreground">
			<ArrowLeft class="h-4 w-4" />
			Projects
		</Button>

		{#if loading}
			<p class="text-sm text-muted-foreground">Loading…</p>
		{:else if error}
			<p class="text-sm text-destructive">{error}</p>
		{:else if project}
			<div class="flex flex-col gap-1">
				<h1 class="text-2xl font-bold tracking-tight">{project.name}</h1>
				<p class="text-sm text-muted-foreground">
					{project.description || 'No description'}
				</p>
				<p class="mt-1 text-xs text-muted-foreground">
					Created {formatDateShort(project.created_at)}
				</p>
			</div>

			<Card.Root class="mt-6">
				<Card.Header class="gap-4">
					<div class="flex items-start justify-between gap-3">
						<div>
							<Card.Title>Project Settings</Card.Title>
							<Card.Description>Edit metadata or delete the project from here, not from the grid.</Card.Description>
						</div>
						<div class="flex shrink-0 gap-2">
							<Button variant="outline" size="sm" onclick={startProjectEdit}>
								<Pencil class="h-4 w-4" />
								Edit
							</Button>
							<Button
								size="sm"
								class="border-destructive bg-destructive text-white hover:bg-destructive/90 hover:text-white"
								onclick={() => { projectDeleteDialogOpen = true; }}
								disabled={deletingProject}
							>
								<Trash2 class="h-4 w-4" />
								{deletingProject ? 'Deleting…' : 'Delete'}
							</Button>
						</div>
					</div>
				</Card.Header>
				<Card.Content class="flex flex-col gap-4">
					{#if projectActionError}
						<p class="text-sm text-destructive">{projectActionError}</p>
					{/if}
					<div class="grid gap-4 md:grid-cols-2">
						<div>
							<p class="text-xs uppercase tracking-wide text-muted-foreground">Name</p>
							<p class="mt-1 font-medium">{project.name}</p>
						</div>
						<div>
							<p class="text-xs uppercase tracking-wide text-muted-foreground">Description</p>
							<p class="mt-1 text-sm text-muted-foreground">
								{project.description || 'No description'}
							</p>
						</div>
					</div>
				</Card.Content>
			</Card.Root>

			<div class="mt-6 grid grid-cols-2 gap-4 sm:grid-cols-4">
				<Card.Root>
					<Card.Header class="pb-2">
						<Card.Title class="text-sm font-medium text-muted-foreground">Total Time</Card.Title>
					</Card.Header>
					<Card.Content>
						<div class="flex items-center gap-2">
							<Clock class="h-4 w-4 text-muted-foreground" />
							<span class="text-2xl font-bold tabular-nums">{formatDuration(totalMs)}</span>
						</div>
					</Card.Content>
				</Card.Root>

				<Card.Root>
					<Card.Header class="pb-2">
						<Card.Title class="text-sm font-medium text-muted-foreground">Tasks</Card.Title>
					</Card.Header>
					<Card.Content>
						<div class="flex items-center gap-2">
							<BarChart3 class="h-4 w-4 text-muted-foreground" />
							<span class="text-2xl font-bold">{tasks.length}</span>
						</div>
					</Card.Content>
				</Card.Root>

				<Card.Root>
					<Card.Header class="pb-2">
						<Card.Title class="text-sm font-medium text-muted-foreground">Avg Session</Card.Title>
					</Card.Header>
					<Card.Content>
						<div class="flex items-center gap-2">
							<Timer class="h-4 w-4 text-muted-foreground" />
							<span class="text-2xl font-bold tabular-nums">{formatDuration(avgMs)}</span>
						</div>
					</Card.Content>
				</Card.Root>

				<Card.Root>
					<Card.Header class="pb-2">
						<Card.Title class="text-sm font-medium text-muted-foreground">Last Session</Card.Title>
					</Card.Header>
					<Card.Content>
						<div class="flex items-center gap-2">
							<Calendar class="h-4 w-4 text-muted-foreground" />
							<span class="text-sm font-medium">
								{lastSession ? formatDate(lastSession.started_at) : 'Never'}
							</span>
						</div>
					</Card.Content>
				</Card.Root>
			</div>

			<section class="mt-6 rounded-2xl border p-5">
				<div>
					<div>
						<h2 class="text-lg font-semibold">User Repartition</h2>
						<p class="text-sm text-muted-foreground">
							Whole-project split by tracked time per user.
						</p>
					</div>
				</div>

				<div class="mt-4">
					{#if projectUserSegments.length === 0}
						<p class="text-sm text-muted-foreground">No tracked time yet.</p>
					{:else}
						<UserColorSplitBar segments={projectUserSegments} barClass="h-4" />
					{/if}
				</div>
			</section>

			<section class="mt-6">
				<div class="mb-4 flex items-start justify-between gap-3">
					<h2 class="text-lg font-semibold">Tasks</h2>
					<p class="text-right text-xs text-muted-foreground">
						Time shown is total time spent per task.
					</p>
				</div>
				<div>
					{#if taskSaveError}
						<p class="mb-4 text-sm text-destructive">{taskSaveError}</p>
					{/if}
					{#if tasksWithStats.length === 0}
						<p class="text-sm text-muted-foreground">No tasks yet.</p>
					{:else}
						<div class="space-y-3">
							{#each tasksWithStats as task}
								<div class="rounded-xl border p-4">
									<div class="flex items-start justify-between gap-3">
										<div class="min-w-0">
											{#if editingTaskId === task.id}
												<div class="flex flex-col gap-2">
													<Input
														bind:value={taskDraftName}
														class="h-8"
														maxlength={200}
													/>
													<div class="flex flex-wrap gap-2">
														<Button
															size="sm"
															onclick={() => saveTaskName(task.id)}
															disabled={savingTaskId === task.id}
														>
															<Check class="h-4 w-4" />
															{savingTaskId === task.id ? 'Saving…' : 'Save'}
														</Button>
														<Button
															variant="outline"
															size="sm"
															onclick={cancelTaskEdit}
															disabled={savingTaskId === task.id}
														>
															<X class="h-4 w-4" />
															Cancel
														</Button>
													</div>
												</div>
											{:else}
												<p class="truncate font-medium" title={task.name}>{task.name}</p>
											{/if}
										</div>
										<div class="flex items-center gap-1">
											<Badge variant="secondary" class="tabular-nums">
												{formatDuration(task.totalMs)}
											</Badge>
											<Button
												variant="ghost"
												size="icon"
												class="h-7 w-7 text-muted-foreground opacity-50 hover:opacity-100"
												onclick={() => startTaskEdit(task.id, task.name)}
												disabled={editingTaskId !== null && editingTaskId !== task.id}
											>
												<Pencil class="h-3.5 w-3.5" />
											</Button>
											<Button
												variant="ghost"
												size="icon"
												class="h-7 w-7 text-muted-foreground opacity-50 hover:text-destructive hover:opacity-100"
												onclick={() => openTaskDeleteDialog(task.id, task.name, task.sessionCount)}
												disabled={deletingTaskId === task.id || (editingTaskId !== null && editingTaskId !== task.id)}
											>
												<Trash2 class="h-3.5 w-3.5" />
											</Button>
										</div>
									</div>
									<div class="mt-3">
										{#if task.userSegments.length > 0}
											<UserColorSplitBar segments={task.userSegments} />
										{:else}
											<div class="h-3 w-full rounded-full bg-muted/40"></div>
										{/if}
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			</section>

			<section class="mt-6">
				<div class="mb-4">
					<h2 class="text-lg font-semibold">Sessions</h2>
				</div>
				<div>
					{#if deleteError}
						<p class="mb-4 text-sm text-destructive">{deleteError}</p>
					{/if}

					{#if sortedEntries.length === 0}
						<p class="text-sm text-muted-foreground">No sessions yet.</p>
					{:else}
						<Table.Root>
							<Table.Header>
								<Table.Row>
									<Table.Head>User</Table.Head>
									<Table.Head>Task</Table.Head>
								<Table.Head>Started</Table.Head>
								<Table.Head>Duration</Table.Head>
								<Table.Head class="text-right">Actions</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each sortedEntries as entry}
									{@const isRunning = entry.stopped_at === null}
									{@const durationMs = isRunning ? now - new Date(entry.started_at).getTime() : entryMs(entry)}
									<Table.Row>
										<Table.Cell class="text-muted-foreground">
											<div class="flex items-center gap-2">
												<UserAvatarBadge
													name={getEntryUserDisplayName(entry)}
													avatarUrl={entry.user_avatar_url}
													color={userColor(entry)}
												/>
												<span>{getEntryUserDisplayName(entry)}</span>
											</div>
										</Table.Cell>
										<Table.Cell class="text-muted-foreground">{entry.task_name || '—'}</Table.Cell>
										<Table.Cell class="text-muted-foreground">{formatDate(entry.started_at)}</Table.Cell>
										<Table.Cell>
											{#if isRunning}
												<span class="inline-flex items-center gap-1.5 rounded-full border border-green-500/30 bg-green-500/10 px-2.5 py-0.5 text-xs font-medium text-green-600 dark:text-green-400">
													<span class="relative flex h-2 w-2">
														<span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-500 opacity-75"></span>
														<span class="relative inline-flex h-2 w-2 rounded-full bg-green-500"></span>
													</span>
													Running
												</span>
											{:else}
												<span class="font-mono text-sm tabular-nums">{formatDuration(durationMs)}</span>
											{/if}
										</Table.Cell>
										<Table.Cell class="text-right">
											{#if entry.user_id === Number(ctx.user?.id)}
												<Button
													variant="ghost"
													size="icon"
													class="h-8 w-8 opacity-50 hover:opacity-100"
													onclick={() => openEditDrawer(entry)}
												>
													<Pencil class="h-4 w-4" />
												</Button>
												<Button
													variant="ghost"
													size="icon"
													class="h-8 w-8 text-destructive opacity-50 hover:opacity-100 hover:text-destructive"
													onclick={() => openEntryDeleteDialog(entry)}
													disabled={deletingEntryId === entry.id}
												>
													<Trash2 class="h-4 w-4 text-destructive" />
												</Button>
											{/if}
										</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					{/if}
				</div>
			</section>
		{/if}
	</div>

	<AlertDialog.Root
		bind:open={projectDeleteDialogOpen}
		onOpenChange={(open) => {
			if (!open) {
				projectDeleteDialogOpen = false;
			}
		}}
	>
		<AlertDialog.Content>
			<AlertDialog.Header>
				<AlertDialog.Title>Delete project?</AlertDialog.Title>
				<AlertDialog.Description>
					This will permanently delete
					{#if project}
						<span class="font-medium text-foreground"> {project.name}</span>
					{/if}
					.
				</AlertDialog.Description>
			</AlertDialog.Header>
			<AlertDialog.Footer>
				<AlertDialog.Cancel disabled={deletingProject}>Cancel</AlertDialog.Cancel>
				<AlertDialog.Action
					variant="destructive"
					class="bg-destructive text-white hover:bg-destructive/90 hover:text-white"
					disabled={deletingProject}
					onclick={(e) => {
						e.preventDefault();
						void deleteProject();
					}}
				>
					<Trash2 class="h-4 w-4" />
					{deletingProject ? 'Deleting…' : 'Delete'}
				</AlertDialog.Action>
			</AlertDialog.Footer>
		</AlertDialog.Content>
	</AlertDialog.Root>

	<Drawer.Root bind:open={projectEditDrawerOpen} direction="bottom">
		<Drawer.Portal>
			<Drawer.Overlay class="fixed inset-0 bg-black/40" />
			<Drawer.Content class="fixed bottom-0 left-0 right-0 flex flex-col rounded-t-2xl bg-background border-t">
				<div class="mx-auto mt-4 mb-6 h-1.5 w-12 shrink-0 rounded-full bg-muted"></div>
				<div class="mx-auto flex w-full max-w-lg flex-col gap-6 px-6 pb-8">
					<Drawer.Header class="p-0">
						<Drawer.Title>Edit project</Drawer.Title>
						<Drawer.Description>
							Update the project name and description from here.
						</Drawer.Description>
					</Drawer.Header>

					{#if projectActionError}
						<p class="text-sm text-destructive">{projectActionError}</p>
					{/if}

					<div class="flex flex-col gap-4">
						<div class="flex flex-col gap-1.5">
							<Label for="project-edit-name">Name</Label>
							<Input id="project-edit-name" bind:value={editName} />
						</div>
						<div class="flex flex-col gap-1.5">
							<Label for="project-edit-description">Description</Label>
							<Input
								id="project-edit-description"
								bind:value={editDescription}
								placeholder="Optional"
							/>
						</div>
					</div>

					<div class="flex flex-wrap gap-2">
						<Button onclick={saveProject} disabled={savingProject}>
							<Save class="h-4 w-4" />
							{savingProject ? 'Saving…' : 'Save'}
						</Button>
						<Button variant="outline" onclick={cancelProjectEdit} disabled={savingProject}>
							<X class="h-4 w-4" />
							Cancel
						</Button>
					</div>
				</div>
			</Drawer.Content>
		</Drawer.Portal>
	</Drawer.Root>

	<ManualSessionDrawer
		projects={project ? [project] : []}
		editEntry={editingEntry}
		bind:open={editDrawerOpen}
		hideTrigger
		onchange={handleEntryChange}
		onclose={() => {
			editingEntry = null;
		}}
	/>

	<AlertDialog.Root
		bind:open={taskDeleteDialogOpen}
		onOpenChange={(open) => {
			if (!open) {
				deleteTaskTarget = null;
			}
		}}
	>
		<AlertDialog.Content>
			<AlertDialog.Header>
				<AlertDialog.Title>Delete task?</AlertDialog.Title>
				<AlertDialog.Description>
					{#if deleteTaskTarget}
						This will permanently delete
						<span class="font-medium text-foreground"> {deleteTaskTarget.name}</span>.
						{#if deleteTaskTarget.sessionCount > 0}
							{deleteTaskTarget.sessionCount}
							{deleteTaskTarget.sessionCount === 1 ? ' session will' : ' sessions will'} become unassigned.
						{/if}
					{/if}
				</AlertDialog.Description>
			</AlertDialog.Header>
			<AlertDialog.Footer>
				<AlertDialog.Cancel disabled={deletingTaskId !== null}>Cancel</AlertDialog.Cancel>
				<AlertDialog.Action
					variant="destructive"
					class="bg-destructive text-white hover:bg-destructive/90 hover:text-white"
					disabled={deletingTaskId !== deleteTaskTarget?.id && deletingTaskId !== null}
					onclick={(e) => {
						e.preventDefault();
						void confirmDeleteTask();
					}}
				>
					<Trash2 class="h-4 w-4" />
					{deletingTaskId === deleteTaskTarget?.id ? 'Deleting…' : 'Delete'}
				</AlertDialog.Action>
			</AlertDialog.Footer>
		</AlertDialog.Content>
	</AlertDialog.Root>

	<AlertDialog.Root
		bind:open={entryDeleteDialogOpen}
		onOpenChange={(open) => {
			if (!open) {
				deleteEntryTarget = null;
			}
		}}
	>
		<AlertDialog.Content>
			<AlertDialog.Header>
				<AlertDialog.Title>Delete session?</AlertDialog.Title>
				<AlertDialog.Description>
					This will permanently remove the session
					{#if deleteEntryTarget?.task_name}
						for <span class="font-medium text-foreground">{deleteEntryTarget.task_name}</span>
					{/if}
					.
				</AlertDialog.Description>
			</AlertDialog.Header>
			<AlertDialog.Footer>
				<AlertDialog.Cancel disabled={deletingEntryId !== null}>Cancel</AlertDialog.Cancel>
				<AlertDialog.Action
					variant="destructive"
					class="bg-destructive text-white hover:bg-destructive/90 hover:text-white"
					disabled={deletingEntryId !== deleteEntryTarget?.id && deletingEntryId !== null}
					onclick={(e) => {
						e.preventDefault();
						void confirmDeleteEntry();
					}}
				>
					<Trash2 class="h-4 w-4" />
					{deletingEntryId === deleteEntryTarget?.id ? 'Deleting…' : 'Delete'}
				</AlertDialog.Action>
			</AlertDialog.Footer>
		</AlertDialog.Content>
	</AlertDialog.Root>
</div>
