<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import { page } from '$app/state';
	import { backend, type Project, type Task, type TimeEntry } from '$lib/backend';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Clock, BarChart3, Calendar, ArrowLeft, Timer } from 'lucide-svelte';

	const ctx = getContext<{ token: string; userEmail: string }>('app');

	let loading = $state(true);
	let error = $state('');
	let project = $state<Project | null>(null);
	let tasks = $state<Task[]>([]);
	let entries = $state<TimeEntry[]>([]);
	let deletingEntryId = $state<number | null>(null);
	let deleteError = $state('');

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

	const sortedEntries = $derived(
		[...entries].sort((a, b) => new Date(b.started_at).getTime() - new Date(a.started_at).getTime())
	);

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

			<Card.Root class="mt-6">
				<Card.Header>
					<Card.Title>Tasks</Card.Title>
				</Card.Header>
				<Card.Content>
					{#if tasksWithStats.length === 0}
						<p class="text-sm text-muted-foreground">No tasks yet.</p>
					{:else}
						<div class="space-y-3">
							{#each tasksWithStats as task}
								<div class="rounded-xl border p-4">
									<div class="flex items-start justify-between gap-3">
										<div class="min-w-0">
											<p class="truncate font-medium">{task.name}</p>
											<p class="mt-1 text-sm text-muted-foreground">
												{task.sessionCount} {task.sessionCount === 1 ? 'session' : 'sessions'}
											</p>
										</div>
										<Badge variant="secondary" class="tabular-nums">
											{formatDuration(task.totalMs)}
										</Badge>
									</div>
									<p class="mt-2 text-xs text-muted-foreground">
										{task.lastStartedAt ? `Last session ${formatDate(task.lastStartedAt)}` : 'No sessions on this task yet'}
									</p>
								</div>
							{/each}
						</div>
					{/if}
				</Card.Content>
			</Card.Root>

			<Card.Root class="mt-6">
				<Card.Header>
					<Card.Title>Sessions</Card.Title>
				</Card.Header>
				<Card.Content>
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
										<Table.Cell class="text-muted-foreground">{entry.user_email ?? '—'}</Table.Cell>
										<Table.Cell class="text-muted-foreground">{entry.task_name || '—'}</Table.Cell>
										<Table.Cell class="text-muted-foreground">
											{formatDate(entry.started_at)}
										</Table.Cell>
										<Table.Cell class="text-right">
											{#if entry.stopped_at === null}
												<Badge variant="outline" class="gap-1.5 border-foreground font-medium">
													<span class="relative flex h-2 w-2">
														<span
															class="absolute inline-flex h-full w-full animate-ping rounded-full bg-foreground opacity-75"
														></span>
														<span
															class="relative inline-flex h-2 w-2 rounded-full bg-foreground"
														></span>
													</span>
													Running…
												</Badge>
											{:else}
												<span class="tabular-nums">{formatDuration(entryMs(entry))}</span>
											{/if}
										</Table.Cell>
										<Table.Cell class="text-right">
											<Button
												variant="ghost"
												size="sm"
												class="text-muted-foreground hover:text-destructive"
												onclick={() => handleDelete(entry.id)}
												disabled={deletingEntryId === entry.id}
											>
												{deletingEntryId === entry.id ? 'Removing…' : 'Remove'}
											</Button>
										</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					{/if}
				</Card.Content>
			</Card.Root>
		{/if}
	</div>
</div>
