<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { backend, type Project, type Task, type TimeEntry, type UserProfile } from '$lib/backend';
	import { getEntryUserDisplayName } from '$lib/user-display';
	import UserColorDot from '$lib/components/UserColorDot.svelte';
	import UserColorSplitBar from '$lib/components/UserColorSplitBar.svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Clock, BarChart3, Calendar, ArrowLeft, Timer, Pencil, Trash2, Check, X } from 'lucide-svelte';

	const ctx = getContext<{ token: string; userEmail: string; user: UserProfile | null }>('app');

	let loading = $state(true);
	let error = $state('');
	let project = $state<Project | null>(null);
	let tasks = $state<Task[]>([]);
	let entries = $state<TimeEntry[]>([]);
	let editingProject = $state(false);
	let editName = $state('');
	let editDescription = $state('');
	let projectActionError = $state('');
	let savingProject = $state(false);
	let deletingProject = $state(false);
	let deletingEntryId = $state<number | null>(null);
	let deleteError = $state('');
	let deletingTaskId = $state<number | null>(null);
	let taskDeleteError = $state('');

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
		const end = e.stopped_at ? new Date(e.stopped_at).getTime() : Date.now();
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

	async function handleDelete(entryId: number) {
		if (!confirm('Remove this session?')) {
			return;
		}
		deletingEntryId = entryId;
		deleteError = '';
		try {
			await backend.deleteEntry(ctx.token, entryId);
			entries = entries.filter((entry) => entry.id !== entryId);
		} catch (e) {
			deleteError = e instanceof Error ? e.message : 'Failed to remove session.';
		} finally {
			deletingEntryId = null;
		}
	}

	async function handleDeleteTask(taskId: number, sessionCount: number) {
		const msg = sessionCount > 0
			? `Delete this task? ${sessionCount} ${sessionCount === 1 ? 'session' : 'sessions'} will become unassigned.`
			: 'Delete this task?';
		if (!confirm(msg)) return;
		deletingTaskId = taskId;
		taskDeleteError = '';
		try {
			await backend.deleteTask(ctx.token, project!.id, taskId);
			tasks = tasks.filter((t) => t.id !== taskId);
			entries = entries.map((e) => e.task_id === taskId ? { ...e, task_id: 0, task_name: '' } : e);
		} catch (e) {
			taskDeleteError = e instanceof Error ? e.message : 'Failed to delete task.';
		} finally {
			deletingTaskId = null;
		}
	}

	function startProjectEdit() {
		if (!project) {
			return;
		}
		projectActionError = '';
		editName = project.name;
		editDescription = project.description;
		editingProject = true;
	}

	function cancelProjectEdit() {
		editingProject = false;
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
			editingProject = false;
		} catch (e) {
			projectActionError = e instanceof Error ? e.message : 'Failed to save project.';
		} finally {
			savingProject = false;
		}
	}

	async function deleteProject() {
		if (!project || !confirm(`Delete project "${project.name}"?`)) {
			return;
		}
		deletingProject = true;
		projectActionError = '';
		try {
			await backend.deleteProject(ctx.token, project.id);
			await goto('/projects');
		} catch (e) {
			projectActionError = e instanceof Error ? e.message : 'Failed to delete project.';
			deletingProject = false;
		}
	}

	onMount(async () => {
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
						{#if !editingProject}
							<div class="flex shrink-0 gap-2">
								<Button variant="outline" size="sm" onclick={startProjectEdit}>
									<Pencil class="h-4 w-4" />
									Edit
								</Button>
								<Button
									variant="destructive"
									size="sm"
									onclick={deleteProject}
									disabled={deletingProject}
								>
									<Trash2 class="h-4 w-4" />
									{deletingProject ? 'Deleting…' : 'Delete'}
								</Button>
							</div>
						{/if}
					</div>
				</Card.Header>
				<Card.Content class="flex flex-col gap-4">
					{#if projectActionError}
						<p class="text-sm text-destructive">{projectActionError}</p>
					{/if}

					{#if editingProject}
						<div class="grid gap-4 md:grid-cols-2">
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
								<Check class="h-4 w-4" />
								{savingProject ? 'Saving…' : 'Save'}
							</Button>
							<Button variant="outline" onclick={cancelProjectEdit} disabled={savingProject}>
								<X class="h-4 w-4" />
								Cancel
							</Button>
						</div>
					{:else}
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
					{/if}
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
				<div class="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
					<div>
						<h2 class="text-lg font-semibold">User Repartition</h2>
						<p class="text-sm text-muted-foreground">
							Whole-project split by tracked time per user.
						</p>
					</div>
					<Badge variant="outline" class="w-fit tabular-nums">
						{formatDuration(totalMs)}
					</Badge>
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
					{#if tasksWithStats.length === 0}
						<p class="text-sm text-muted-foreground">No tasks yet.</p>
					{:else}
						<div class="space-y-3">
							{#each tasksWithStats as task}
								<div class="rounded-xl border p-4">
									<div class="flex items-start justify-between gap-3">
										<div class="min-w-0">
											<p class="truncate font-medium">{task.name}</p>
										</div>
										<div class="flex items-center gap-1">
											<Badge variant="secondary" class="tabular-nums">
												{formatDuration(task.totalMs)}
											</Badge>
											<Button
												variant="ghost"
												size="icon"
												class="h-7 w-7 text-muted-foreground opacity-50 hover:text-destructive hover:opacity-100"
												onclick={() => handleDeleteTask(task.id, task.sessionCount)}
												disabled={deletingTaskId === task.id}
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
									<Table.Head class="text-right">Duration</Table.Head>
									<Table.Head class="text-right">Actions</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each sortedEntries as entry}
									<Table.Row>
										<Table.Cell class="text-muted-foreground">
											<div class="flex items-center gap-2">
												<UserColorDot color={userColor(entry)} />
												<span>{getEntryUserDisplayName(entry)}</span>
											</div>
										</Table.Cell>
										<Table.Cell class="text-muted-foreground">{entry.task_name || '—'}</Table.Cell>
										<Table.Cell class="text-muted-foreground">
											{formatDate(entry.started_at)}
										</Table.Cell>
										<Table.Cell class="text-right">
											{#if entry.stopped_at === null}
												<span class="inline-flex items-center gap-1.5 rounded-full border border-green-500/30 bg-green-500/10 px-2.5 py-0.5 text-xs font-medium text-green-600 dark:text-green-400">
													<span class="relative flex h-2 w-2">
														<span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-500 opacity-75"></span>
														<span class="relative inline-flex h-2 w-2 rounded-full bg-green-500"></span>
													</span>
													Running
												</span>
											{:else}
												<span class="tabular-nums">{formatDuration(entryMs(entry))}</span>
											{/if}
										</Table.Cell>
										<Table.Cell class="text-right">
											{#if entry.user_id === Number(ctx.user?.id)}
												<Button
													variant="ghost"
													size="icon"
													class="h-8 w-8 text-destructive opacity-50 hover:opacity-100 hover:text-destructive"
													onclick={() => handleDelete(entry.id)}
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
</div>
