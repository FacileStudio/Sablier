<script lang="ts">
	import { getContext, onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { backend, type Project, type Task, type TimeEntry, type UserProfile } from '$lib/backend';
	import { onTimeEntriesChanged } from '$lib/time-entry-events';
	import { getEntryUserDisplayName } from '$lib/user-display';
	import UserAvatarBadge from '$lib/components/UserAvatarBadge.svelte';
	import UserColorSplitBar from '$lib/components/UserColorSplitBar.svelte';
	import ManualSessionDrawer from '$lib/components/ManualSessionDrawer.svelte';
	import {
		Alert,
		Badge,
		Button,
		Card,
		ConfirmModal,
		Drawer,
		EmptyState,
		Field,
		Input,
		Spinner,
		StatCard,
		StatusDot,
		Table,
		Tabs,
		icons
	} from '@facile/muse';
	import { formatDuration, getTimeEntryDurationMs, isTimeEntryPaused } from '$lib/utils';
	import IconPicker from '$lib/components/IconPicker.svelte';
	import { toIconify } from '$lib/icons';

	const ctx = getContext<{ token: string; userEmail: string; user: UserProfile | null }>('app');

	let loading = $state(true);
	let error = $state('');
	let project = $state<Project | null>(null);
	let tasks = $state<Task[]>([]);
	let entries = $state<TimeEntry[]>([]);
	let userRates = $state<Map<number, { rate: number; rate_type: 'daily' | 'hourly'; workday_hours: number }>>(new Map());
	let usersById = $state<Map<number, UserProfile>>(new Map());
	let projectEditDrawerOpen = $state(false);
	let editName = $state('');
	let editDescription = $state('');
	let editIcon = $state('Layout');
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
	let taskSearch = $state('');
	let now = $state(Date.now());
	let activeTab = $state('tasks');
	let ticker: ReturnType<typeof setInterval> | undefined;
	let stopTimeEntrySync: (() => void) | undefined;

	type UserTimeSegment = {
		key: string;
		label: string;
		color?: string;
		avatarUrl?: string;
		ms: number;
	};

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
		return getTimeEntryDurationMs(e, now);
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
				avatarUrl: entry.user_avatar_url,
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

	const projectValue = $derived.by(() => {
		if (userRates.size === 0) return null;
		let total = 0;
		let anyRate = false;
		for (const entry of entries) {
			const ur = userRates.get(entry.user_id);
			if (!ur || ur.rate <= 0) continue;
			anyRate = true;
			const hours = entryMs(entry) / 3_600_000;
			total += ur.rate_type === 'hourly' ? hours * ur.rate : (hours / ur.workday_hours) * ur.rate;
		}
		if (!anyRate) return null;
		return new Intl.NumberFormat(undefined, { style: 'currency', currency: 'EUR', maximumFractionDigits: 0 }).format(total);
	});

	const avgMs = $derived(entries.length > 0 ? totalMs / entries.length : 0);

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

	const filteredTasks = $derived(
		taskSearch.trim() === ''
			? tasksWithStats
			: tasksWithStats.filter((t) =>
					t.name.toLowerCase().includes(taskSearch.toLowerCase())
				)
	);

	const tabItems = $derived([
		{ id: 'tasks', label: 'Tasks', icon: icons.check, badge: tasksWithStats.length },
		{ id: 'sessions', label: 'Sessions', icon: icons.clock, badge: sortedEntries.length },
		{ id: 'repartition', label: 'User Repartition', icon: icons.usersGroup }
	]);

	function openEntryDeleteDialog(entry: TimeEntry) {
		deleteEntryTarget = entry;
		deleteError = '';
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
		} catch (e) {
			deleteError = e instanceof Error ? e.message : 'Failed to remove session.';
			throw e;
		} finally {
			deletingEntryId = null;
		}
	}

	function openTaskDeleteDialog(taskId: number, name: string, sessionCount: number) {
		deleteTaskTarget = { id: taskId, name, sessionCount };
		taskSaveError = '';
		taskDeleteError = '';
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
		} catch (e) {
			taskDeleteError = e instanceof Error ? e.message : 'Failed to delete task.';
			throw e;
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
			const updated = await backend.updateTask(ctx.token, project.id, taskId, { name: taskDraftName });
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

	const STATUS_CYCLE = ['to-do', 'in-progress', 'in-review', 'done'] as const;

	function nextTaskStatus(current: string): string {
		const idx = STATUS_CYCLE.indexOf(current as typeof STATUS_CYCLE[number]);
		return STATUS_CYCLE[(idx + 1) % STATUS_CYCLE.length];
	}

	function statusLabel(status: string): string {
		const labels: Record<string, string> = { 'to-do': 'Not started', 'in-progress': 'In progress', 'in-review': 'In review', 'done': 'Completed' };
		return labels[status] ?? 'Not started';
	}

	function statusDotClass(status: string): string {
		if (status === 'done') return 'border-fc-success bg-fc-success';
		if (status === 'in-review') return 'border-fc-info bg-fc-info';
		if (status === 'in-progress') return 'border-fc-warning bg-fc-warning';
		return 'border-fc-border hover:border-fc-fg';
	}

	function statusTone(status: string): 'success' | 'info' | 'warning' | 'neutral' {
		if (status === 'done') return 'success';
		if (status === 'in-review') return 'info';
		if (status === 'in-progress') return 'warning';
		return 'neutral';
	}

	async function toggleTaskStatus(taskId: number) {
		if (!project) return;
		const task = tasks.find((t) => t.id === taskId);
		if (!task) return;
		try {
			const updated = await backend.updateTask(ctx.token, project.id, taskId, { status: nextTaskStatus(task.status) });
			tasks = tasks.map((t) => (t.id === taskId ? updated : t));
		} catch {}
	}

	function startProjectEdit() {
		if (!project) {
			return;
		}
		projectActionError = '';
		editName = project.name;
		editDescription = project.description;
		editIcon = project.icon || 'Layout';
		projectEditDrawerOpen = true;
	}

	function cancelProjectEdit() {
		projectEditDrawerOpen = false;
		projectActionError = '';
		editName = '';
		editDescription = '';
		editIcon = 'Layout';
	}

	async function saveProject() {
		if (!project) {
			return;
		}
		savingProject = true;
		projectActionError = '';
		try {
			project = await backend.updateProject(ctx.token, project.id, editName, editDescription, editIcon);
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
			await goto('/projects');
		} catch (e) {
			projectActionError = e instanceof Error ? e.message : 'Failed to delete project.';
			deletingProject = false;
			throw e;
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
			const [proj, taskResult, ents, usersResult] = await Promise.all([
				backend.getProject(ctx.token, id),
				backend.listTasks(ctx.token, id),
				backend.listEntries(ctx.token, id),
				backend.listUsers(ctx.token)
			]);
			project = proj;
			tasks = taskResult.tasks;
			entries = ents.entries;
			const rateMap = new Map<number, { rate: number; rate_type: 'daily' | 'hourly'; workday_hours: number }>();
			const userMap = new Map<number, UserProfile>();
			for (const u of usersResult.users) {
				const uid = Number(u.id);
				rateMap.set(uid, { rate: u.rate ?? 0, rate_type: u.rate_type ?? 'daily', workday_hours: u.workday_hours > 0 ? u.workday_hours : 8 });
				userMap.set(uid, u);
			}
			userRates = rateMap;
			usersById = userMap;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load project.';
		} finally {
			loading = false;
		}
		stopTimeEntrySync = onTimeEntriesChanged(async () => {
			if (!project) return;
			const [taskResult, ents] = await Promise.all([
				backend.listTasks(ctx.token, project.id),
				backend.listEntries(ctx.token, project.id)
			]);
			tasks = taskResult.tasks;
			entries = ents.entries;
		});
	});

	onDestroy(() => {
		clearInterval(ticker);
		stopTimeEntrySync?.();
	});
</script>

<svelte:head>
	<title>{project?.name ?? 'Project'} — Sablier</title>
</svelte:head>

<div class="flex flex-col gap-10">
	<section class="flex flex-col gap-4">
		<div>
			<Button variant="ghost" href="/projects" icon={icons.chevronLeft} class="-ml-4">Projects</Button>
		</div>

		{#if loading}
			<div class="flex items-center gap-2 text-fc-sm text-fc-fg-muted">
				<Spinner size="sm" />
				Loading…
			</div>
		{:else if error}
			<Alert tone="danger">{error}</Alert>
		{:else if project}
			<div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
				<div class="flex min-w-0 flex-col gap-1">
					<div class="flex items-center gap-2.5">
						<iconify-icon
							icon={toIconify(project.icon)}
							width="24"
							height="24"
							class="block shrink-0 text-fc-fg-muted"
						></iconify-icon>
						<h1 class="truncate text-fc-2xl font-semibold text-fc-fg">{project.name}</h1>
					</div>
					<p class="text-fc-sm text-fc-fg-muted">
						{project.description || 'No description'}
					</p>
					<p class="text-fc-xs text-fc-fg-muted">
						Created {formatDateShort(project.created_at)}
					</p>
				</div>
				<div class="flex shrink-0 items-center gap-2">
					<Button variant="outline" icon={icons.edit} onclick={startProjectEdit}>Edit</Button>
					<Button
						variant="danger"
						icon={icons.remove}
						onclick={() => { projectDeleteDialogOpen = true; }}
						disabled={deletingProject}
					>
						{deletingProject ? 'Deleting…' : 'Delete'}
					</Button>
				</div>
			</div>

			{#if projectActionError}
				<Alert tone="danger">{projectActionError}</Alert>
			{/if}
		{/if}
	</section>

	{#if !loading && !error && project}
		<section class="flex flex-col gap-4">
			<div
				class="grid grid-cols-2 gap-4"
				class:sm:grid-cols-3={projectValue === null}
				class:sm:grid-cols-4={projectValue !== null}
			>
				<StatCard label="Total Time" value={formatDuration(totalMs)} class="tabular-nums" />
				<StatCard label="Tasks" value={tasks.length} class="tabular-nums" />
				<StatCard label="Avg Session" value={formatDuration(avgMs)} class="tabular-nums" />
				{#if projectValue !== null}
					<StatCard label="Project Value" value={projectValue} class="tabular-nums" />
				{/if}
			</div>
		</section>

		<section class="flex flex-col gap-4">
			<Tabs items={tabItems} bind:value={activeTab} panelId="project-panel" label="Project sections" />

			<div id="project-panel" class="flex flex-col gap-4">
				{#if activeTab === 'tasks'}
					<div class="flex flex-col gap-1">
						<h2 class="text-fc-lg font-semibold text-fc-fg">Tasks</h2>
						<p class="text-fc-sm text-fc-fg-muted">Every task tracked on this project.</p>
					</div>

					{#if tasksWithStats.length > 0}
						<div class="relative w-full sm:w-64">
							<iconify-icon
								icon={icons.search}
								width="16"
								height="16"
								class="pointer-events-none absolute top-1/2 left-3 block -translate-y-1/2 text-fc-fg-muted"
							></iconify-icon>
							<Input
								bind:value={taskSearch}
								placeholder="Filter tasks…"
								aria-label="Filter tasks"
								class="pl-9"
							/>
						</div>
					{/if}

					{#if taskSaveError}
						<Alert tone="danger">{taskSaveError}</Alert>
					{/if}

					{#if tasksWithStats.length === 0}
						<EmptyState
							icon={icons.check}
							title="No tasks yet."
							description="Tasks appear here as soon as a session is tracked against one."
						/>
					{:else if filteredTasks.length === 0}
						<EmptyState
							icon={icons.search}
							title={`No tasks matching "${taskSearch}".`}
							description="Try a different filter."
						/>
					{:else}
						<Table>
							<thead>
								<tr>
									<th aria-label="Toggle status"></th>
									<th>Task</th>
									<th>Status</th>
									<th>Tracked</th>
									<th aria-label="Assignee"></th>
									<th aria-label="Actions"></th>
								</tr>
							</thead>
							<tbody>
								{#each filteredTasks as task (task.id)}
									<tr class={task.status === 'done' ? 'opacity-60' : ''}>
										<td>
											<button
												type="button"
												class="mt-0.5 block size-4 shrink-0 rounded-fc-pill border transition-colors {statusDotClass(task.status)}"
												aria-label="Task status: {statusLabel(task.status)} — change"
												title={statusLabel(task.status)}
												onclick={() => toggleTaskStatus(task.id)}
											></button>
										</td>
										<td>
											{#if editingTaskId === task.id}
												<div class="flex min-w-56 flex-col gap-2">
													<Input bind:value={taskDraftName} maxlength={200} aria-label="Task name" />
													<div class="flex flex-wrap gap-2">
														<Button
															size="sm"
															icon={icons.check}
															onclick={() => saveTaskName(task.id)}
															disabled={savingTaskId === task.id}
														>
															{savingTaskId === task.id ? 'Saving…' : 'Save'}
														</Button>
														<Button
															variant="outline"
															size="sm"
															icon={icons.close}
															onclick={cancelTaskEdit}
															disabled={savingTaskId === task.id}
														>
															Cancel
														</Button>
													</div>
												</div>
											{:else}
												<div class="flex min-w-56 flex-col gap-2">
													<p
														class="truncate font-medium {task.status === 'done'
															? 'text-fc-fg-muted line-through'
															: 'text-fc-fg'}"
														title={task.name}
													>
														{task.name}
													</p>
													{#if task.userSegments.length > 0}
														<UserColorSplitBar segments={task.userSegments} showLegend={false} />
													{:else}
														<div class="h-3 w-full rounded-fc-pill bg-fc-surface"></div>
													{/if}
												</div>
											{/if}
										</td>
										<td>
											<Badge tone={statusTone(task.status)}>{statusLabel(task.status)}</Badge>
										</td>
										<td class="font-fc-mono whitespace-nowrap tabular-nums">
											{formatDuration(task.totalMs)}
										</td>
										<td>
											{#if task.actor_id}
												{@const actor = usersById.get(task.actor_id)}
												{#if actor}
													<UserAvatarBadge
														name={actor.name}
														avatarUrl={actor.avatar_url}
														color={actor.color}
														class="size-6 text-fc-xs"
													/>
												{/if}
											{/if}
										</td>
										<td>
											<div class="flex items-center justify-end gap-1">
												<Button
													variant="ghost"
													size="sm"
													icon={icons.edit}
													aria-label="Rename task"
													onclick={() => startTaskEdit(task.id, task.name)}
													disabled={editingTaskId !== null && editingTaskId !== task.id}
												/>
												<Button
													variant="ghost-danger"
													size="sm"
													icon={icons.remove}
													aria-label="Delete task"
													onclick={() => openTaskDeleteDialog(task.id, task.name, task.sessionCount)}
													disabled={deletingTaskId === task.id ||
														(editingTaskId !== null && editingTaskId !== task.id)}
												/>
											</div>
										</td>
									</tr>
								{/each}
							</tbody>
						</Table>
					{/if}
				{:else if activeTab === 'sessions'}
					<div class="flex flex-col gap-1">
						<h2 class="text-fc-lg font-semibold text-fc-fg">Sessions</h2>
						<p class="text-fc-sm text-fc-fg-muted">Every tracked session on this project.</p>
					</div>

					{#if deleteError}
						<Alert tone="danger">{deleteError}</Alert>
					{/if}

					{#if sortedEntries.length === 0}
						<EmptyState
							icon={icons.clock}
							title="No sessions yet."
							description="Start a timer on this project and its sessions land here."
						/>
					{:else}
						<Table>
							<thead>
								<tr>
									<th>User</th>
									<th>Task</th>
									<th>Started</th>
									<th>Duration</th>
									<th aria-label="Actions"></th>
								</tr>
							</thead>
							<tbody>
								{#each sortedEntries as entry (entry.id)}
									{@const isRunning = entry.stopped_at === null}
									{@const paused = isTimeEntryPaused(entry)}
									{@const durationMs = entryMs(entry)}
									<tr>
										<td class="text-fc-fg-muted">
											<div class="flex items-center gap-2 whitespace-nowrap">
												<UserAvatarBadge
													name={getEntryUserDisplayName(entry)}
													avatarUrl={entry.user_avatar_url}
													color={userColor(entry)}
												/>
												<span>{getEntryUserDisplayName(entry)}</span>
											</div>
										</td>
										<td class="text-fc-fg-muted">{entry.task_name || '—'}</td>
										<td class="whitespace-nowrap text-fc-fg-muted">{formatDate(entry.started_at)}</td>
										<td>
											{#if isRunning}
												{#if paused}
													<Badge tone="warning">Paused</Badge>
												{:else}
													<Badge tone="success">
														<StatusDot tone="success" pulse />
														Running
													</Badge>
												{/if}
											{:else}
												<span class="font-fc-mono whitespace-nowrap tabular-nums">
													{formatDuration(durationMs)}
												</span>
											{/if}
										</td>
										<td>
											{#if entry.user_id === Number(ctx.user?.id)}
												<div class="flex items-center justify-end gap-1">
													<Button
														variant="ghost"
														size="sm"
														icon={icons.edit}
														aria-label="Edit session"
														onclick={() => openEditDrawer(entry)}
													/>
													<Button
														variant="ghost-danger"
														size="sm"
														icon={icons.remove}
														aria-label="Delete session"
														onclick={() => openEntryDeleteDialog(entry)}
														disabled={deletingEntryId === entry.id}
													/>
												</div>
											{/if}
										</td>
									</tr>
								{/each}
							</tbody>
						</Table>
					{/if}
				{:else}
					<div class="flex flex-col gap-1">
						<h2 class="text-fc-lg font-semibold text-fc-fg">User Repartition</h2>
						<p class="text-fc-sm text-fc-fg-muted">
							Whole-project split by tracked time per user.
						</p>
					</div>

					{#if projectUserSegments.length === 0}
						<EmptyState
							icon={icons.usersGroup}
							title="No tracked time yet."
							description="The split appears once someone tracks time on this project."
						/>
					{:else}
						<Card>
							<UserColorSplitBar
								segments={projectUserSegments}
								barClass="h-4"
								showAvatars
								showDuration
							/>
						</Card>
					{/if}
				{/if}
			</div>
		</section>
	{/if}
</div>

<ConfirmModal
	bind:open={projectDeleteDialogOpen}
	tone="danger"
	title="Delete project?"
	description={project
		? `${project.name}, its ${tasks.length} task${tasks.length === 1 ? '' : 's'} and its ${entries.length} tracked session${entries.length === 1 ? '' : 's'} are deleted for everyone. This cannot be undone.`
		: 'This project and everything tracked on it are deleted for everyone. This cannot be undone.'}
	confirmLabel="Delete project"
	onConfirm={deleteProject}
/>

<ConfirmModal
	bind:open={taskDeleteDialogOpen}
	tone="danger"
	title="Delete task?"
	description={deleteTaskTarget
		? `${deleteTaskTarget.name} is removed from this project.${
				deleteTaskTarget.sessionCount > 0
					? ` Its ${deleteTaskTarget.sessionCount} tracked session${deleteTaskTarget.sessionCount === 1 ? '' : 's'} stay, but become unassigned.`
					: ''
			}`
		: undefined}
	confirmLabel="Delete task"
	onConfirm={confirmDeleteTask}
	onCancel={() => {
		deleteTaskTarget = null;
	}}
>
	{#if taskDeleteError}
		<Alert tone="danger">{taskDeleteError}</Alert>
	{/if}
</ConfirmModal>

<ConfirmModal
	bind:open={entryDeleteDialogOpen}
	tone="danger"
	title="Delete session?"
	description={deleteEntryTarget?.task_name
		? `The tracked time for ${deleteEntryTarget.task_name} is removed from this project's totals. This cannot be undone.`
		: "This session's tracked time is removed from this project's totals. This cannot be undone."}
	confirmLabel="Delete session"
	onConfirm={confirmDeleteEntry}
	onCancel={() => {
		deleteEntryTarget = null;
	}}
>
	{#if deleteError}
		<Alert tone="danger">{deleteError}</Alert>
	{/if}
</ConfirmModal>

<Drawer
	bind:open={projectEditDrawerOpen}
	title="Edit project"
	description="Update the project name and description from here."
	onClose={() => {
		projectActionError = '';
	}}
>
	<div class="flex flex-col gap-4">
		{#if projectActionError}
			<Alert tone="danger">{projectActionError}</Alert>
		{/if}

		<div class="flex flex-col gap-1.5">
			<span class="text-fc-sm text-fc-fg">Icon</span>
			<IconPicker value={editIcon} onSelect={(icon) => (editIcon = icon)} />
		</div>
		<Field label="Name">
			<Input bind:value={editName} />
		</Field>
		<Field label="Description">
			<Input bind:value={editDescription} placeholder="Optional" />
		</Field>
	</div>

	{#snippet footer()}
		<div class="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
			<Button variant="outline" size="lg" icon={icons.close} onclick={cancelProjectEdit} disabled={savingProject}>
				Cancel
			</Button>
			<Button size="lg" icon={icons.check} onclick={saveProject} disabled={savingProject}>
				{savingProject ? 'Saving…' : 'Save'}
			</Button>
		</div>
	{/snippet}
</Drawer>

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
