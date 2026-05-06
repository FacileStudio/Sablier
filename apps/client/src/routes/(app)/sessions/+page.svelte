<script lang="ts">
	import { getContext, onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { backend, type Project, type TimeEntry, type UserProfile } from '$lib/backend';
	import { getEntryUserDisplayName } from '$lib/user-display';
	import UserAvatarBadge from '$lib/components/UserAvatarBadge.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import * as AlertDialog from '$lib/components/ui/alert-dialog';
	import * as Table from '$lib/components/ui/table';
	import * as Select from '$lib/components/ui/select';
	import TimerControl from '$lib/components/TimerControl.svelte';
	import ManualSessionDrawer from '$lib/components/ManualSessionDrawer.svelte';
	import { Trash2, Pencil } from 'lucide-svelte';

	const ctx = getContext<{ token: string; userEmail: string; user: UserProfile | null }>('app');

	let projects = $state<Project[]>([]);
	let entries = $state<TimeEntry[]>([]);
	let selectedProjectId = $state<string>('all');
	let now = $state(Date.now());
	let deleteError = $state('');
	let deletingEntryId = $state<number | null>(null);
	let deleteDialogOpen = $state(false);
	let deleteTarget = $state<TimeEntry | null>(null);
	let editingEntry = $state<TimeEntry | null>(null);
	let editDrawerOpen = $state(false);

	let ticker: ReturnType<typeof setInterval> | undefined;

	let filteredEntries = $derived(
		selectedProjectId === 'all'
			? entries
			: entries.filter((e) => String(e.project_id) === selectedProjectId)
	);

	function formatDuration(ms: number): string {
		const totalSecs = Math.floor(ms / 1000);
		const h = Math.floor(totalSecs / 3600);
		const m = Math.floor((totalSecs % 3600) / 60);
		const s = totalSecs % 60;
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

	function projectName(id: number): string {
		return projects.find((p) => p.id === id)?.name ?? '—';
	}

	function userColor(entry: TimeEntry) {
		return (entry as TimeEntry & { user_color?: string }).user_color;
	}

	async function loadEntries() {
		const result = await backend.listEntries(ctx.token);
		entries = result.entries;
	}

	async function loadAll() {
		const [projectsRes, entriesRes] = await Promise.all([
			backend.listProjects(ctx.token),
			backend.listEntries(ctx.token)
		]);
		projects = projectsRes.projects;
		entries = entriesRes.entries;
	}

	function openDeleteDialog(entry: TimeEntry) {
		deleteTarget = entry;
		deleteDialogOpen = true;
	}

	async function confirmDelete() {
		if (!deleteTarget) {
			return;
		}
		deletingEntryId = deleteTarget.id;
		deleteError = '';
		try {
			await backend.deleteEntry(ctx.token, deleteTarget.id);
			await loadEntries();
			deleteDialogOpen = false;
			deleteTarget = null;
		} catch (e) {
			deleteError = e instanceof Error ? e.message : 'Failed to remove session.';
		} finally {
			deletingEntryId = null;
		}
	}

	async function handleTimerChange() {
		await loadEntries();
	}

	onMount(async () => {
		await loadAll();
		ticker = setInterval(() => { now = Date.now(); }, 1000);
	});

	onDestroy(() => clearInterval(ticker));
</script>

<svelte:head>
	<title>Sessions — Sablier</title>
</svelte:head>

<div class="flex flex-col gap-6 p-6">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-bold tracking-tight">Sessions</h1>
		<div class="flex items-center gap-3">
			<TimerControl {projects} onchange={handleTimerChange} />
			<ManualSessionDrawer {projects} onchange={handleTimerChange} />
		<ManualSessionDrawer {projects} editEntry={editingEntry} bind:open={editDrawerOpen} hideTrigger onchange={handleTimerChange} onclose={() => { editingEntry = null; }} />
		</div>
	</div>

	<div class="flex items-center gap-3">
		<Select.Root type="single" bind:value={selectedProjectId}>
			<Select.Trigger class="w-48">
				{#if selectedProjectId === 'all'}
					All projects
				{:else}
					{projectName(Number(selectedProjectId))}
				{/if}
			</Select.Trigger>
			<Select.Content>
				<Select.Item value="all">All projects</Select.Item>
				{#each projects as project}
					<Select.Item value={String(project.id)}>{project.name}</Select.Item>
				{/each}
			</Select.Content>
		</Select.Root>
	</div>

	{#if deleteError}
		<p class="text-sm text-destructive">{deleteError}</p>
	{/if}

	{#if filteredEntries.length === 0}
		<div class="flex justify-center py-16 text-muted-foreground text-sm">
			No sessions yet.
		</div>
	{:else}
		<Table.Root>
			<Table.Header>
				<Table.Row>
					<Table.Head>Project</Table.Head>
					<Table.Head>User</Table.Head>
					<Table.Head>Task</Table.Head>
					<Table.Head>Started</Table.Head>
					<Table.Head>Duration</Table.Head>
					<Table.Head class="text-right">Actions</Table.Head>
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{#each filteredEntries as entry}
					{@const isRunning = entry.stopped_at === null}
					{@const startMs = new Date(entry.started_at).getTime()}
					{@const durationMs = isRunning ? now - startMs : new Date(entry.stopped_at!).getTime() - startMs}
					<Table.Row
						class="cursor-pointer"
						onclick={() => goto(`/projects/${entry.project_id}`)}
					>
						<Table.Cell class="font-medium">
							<span class="hover:underline">{projectName(entry.project_id)}</span>
						</Table.Cell>
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
									onclick={(e) => { e.stopPropagation(); editingEntry = entry; editDrawerOpen = true; }}
									class="h-8 w-8 opacity-50 hover:opacity-100"
								>
									<Pencil class="h-4 w-4" />
								</Button>
								<Button
									variant="ghost"
									size="icon"
									onclick={(e) => { e.stopPropagation(); openDeleteDialog(entry); }}
									class="h-8 w-8 text-destructive opacity-50 hover:opacity-100 hover:text-destructive"
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

	<AlertDialog.Root
		bind:open={deleteDialogOpen}
		onOpenChange={(open) => {
			if (!open) {
				deleteTarget = null;
			}
		}}
	>
		<AlertDialog.Content>
			<AlertDialog.Header>
				<AlertDialog.Title>Delete session?</AlertDialog.Title>
				<AlertDialog.Description>
					This will permanently remove the session
					{#if deleteTarget?.task_name}
						for <span class="font-medium text-foreground">{deleteTarget.task_name}</span>
					{/if}
					.
				</AlertDialog.Description>
			</AlertDialog.Header>
			<AlertDialog.Footer>
				<AlertDialog.Cancel disabled={deletingEntryId !== null}>Cancel</AlertDialog.Cancel>
				<AlertDialog.Action
					variant="destructive"
					disabled={deletingEntryId !== deleteTarget?.id && deletingEntryId !== null}
					onclick={(e) => {
						e.preventDefault();
						void confirmDelete();
					}}
				>
					{deletingEntryId === deleteTarget?.id ? 'Deleting…' : 'Delete'}
				</AlertDialog.Action>
			</AlertDialog.Footer>
		</AlertDialog.Content>
	</AlertDialog.Root>
</div>
